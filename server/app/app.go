package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/seatsurfing/seatsurfing/server/api"
	"github.com/seatsurfing/seatsurfing/server/api/hostapipb"
	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var _appInstance *App
var _appOnce sync.Once

// PluginInstance represents one connected (or reconnecting) plugin. Unlike
// the old net/rpc/go-plugin subprocess model, a plugin here is an
// independent network process the host dials as a gRPC client - Ready
// reflects whether the connect-driven registration sequence
// (RunSchemaUpdates -> unauthorized routes -> OnInit) has succeeded on the
// current connection; forwardToPlugin/hooks treat a not-ready instance as
// unavailable rather than crashing or silently dropping it.
type PluginInstance struct {
	Instance   api.SeatsurfingPlugin
	client     *api.PluginGRPC // same value as Instance, typed concretely for RunSchemaUpdatesErr/OnInitErr
	Conn       *grpc.ClientConn
	Name       string
	Config     RemotePluginConfig
	Ready      atomic.Bool
	registered atomic.Bool // guards the one-time api.RegisterPlugin call
}

func GetApp() *App {
	_appOnce.Do(func() {
		_appInstance = &App{}
	})
	return _appInstance
}

type App struct {
	Router            *mux.Router
	PublicHttpServer  *http.Server
	CleanupTicker     *time.Ticker
	PluginInstances   []*PluginInstance
	hostAPIGRPCServer *grpc.Server
}

func (a *App) InitializeDatabases() {
	RunDBSchemaUpdates()
	// Plugin schema updates are NOT run here anymore: this ran synchronously
	// against every plugin in the old subprocess model, but now plugins are
	// independent gRPC connections that may not even be dialed yet (see
	// loadGRPCPlugin's comment - Connect() is deferred to NotifyPlugins).
	// Plugin schema updates now happen inside registerPluginOnConnect, gated
	// by actual connection readiness, once NotifyPlugins starts the
	// connection watchers.
	InitDefaultOrgSettings()
	InitDefaultUserPreferences()
	// Set up email logging callback
	SetEmailLogCallback(func(subject, recipient, organizationID string) error {
		return GetMailLogRepository().LogEmail(subject, recipient, organizationID)
	})
	// Set up email footer provider: DB value takes precedence over file fallback
	SetGlobalEmailFooterProvider(func(language string) (string, error) {
		return GetSettingsRepository().GetGlobalStringLocalized(api.SettingEmailFooterPrefix, language)
	})
}

// InitializePlugins dials every plugin listed in PLUGINS_CONFIG over gRPC.
// Unlike the old directory-scan model, dialing is non-blocking (grpc.NewClient
// does not wait for the connection to become ready) and each plugin's
// registration (schema updates, unauthorized routes, OnInit) happens
// asynchronously once connected, driven by watchPluginConnection - so a
// plugin that is down at host startup, or restarts later, does not block
// host startup and recovers on its own once reachable again.
func (a *App) InitializePlugins() {
	for _, pc := range GetConfig().Plugins {
		a.loadGRPCPlugin(pc)
	}
}

func (a *App) loadGRPCPlugin(pc RemotePluginConfig) {
	log.Println("Loading plugin", pc.Name, "at", pc.Address)

	var transportCreds credentials.TransportCredentials
	if pc.TLS {
		transportCreds = credentials.NewTLS(nil)
	} else {
		transportCreds = insecure.NewCredentials()
	}

	conn, err := grpc.NewClient(pc.Address,
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithPerRPCCredentials(api.NewTokenCredentials(pc.Token, pc.TLS)),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(GetConfig().PluginMaxMsgSize),
			grpc.MaxCallSendMsgSize(GetConfig().PluginMaxMsgSize),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                20 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		log.Println("Error creating gRPC client for plugin", pc.Name, ":", err)
		return
	}

	plg := api.NewPluginGRPC(conn, GetConfig().PluginCallTimeout)
	instance := &PluginInstance{
		Instance: plg,
		client:   plg,
		Conn:     conn,
		Name:     pc.Name,
		Config:   pc,
	}
	// Deliberately do not Connect()/start the connection watcher yet: that
	// happens in NotifyPlugins, which main.go calls only after the host's
	// own database migrations/default org/settings are initialized, so a
	// plugin's OnInit (which may call back into HostAPI) can never race
	// ahead of the host's own startup sequence.
	a.PluginInstances = append(a.PluginInstances, instance)
}

// watchPluginConnection runs for the lifetime of the process, observing the
// plugin's gRPC connection state. Every time the connection (re)enters
// READY - including the very first connect - it re-runs the registration
// sequence and only then marks the instance Ready; every time it leaves
// READY, the instance is immediately marked not-ready so forwardToPlugin and
// hook call sites stop routing to it until it recovers.
func (a *App) watchPluginConnection(inst *PluginInstance) {
	ctx := context.Background()
	state := inst.Conn.GetState()
	for {
		for state != connectivity.Ready {
			if state == connectivity.Idle {
				inst.Conn.Connect()
			}
			if !inst.Conn.WaitForStateChange(ctx, state) {
				return
			}
			state = inst.Conn.GetState()
		}
		a.registerPluginOnConnect(inst)
		for state == connectivity.Ready {
			if !inst.Conn.WaitForStateChange(ctx, state) {
				return
			}
			state = inst.Conn.GetState()
		}
		if inst.Ready.Swap(false) {
			log.Printf("Plugin %s: connection lost, marking not ready", inst.Name)
		}
	}
}

// registerPluginOnConnect runs RunSchemaUpdates -> unauthorized routes ->
// OnInit against a freshly (re)connected plugin. If schema updates fail, the
// instance is left not-ready (rather than silently treated as succeeded) and
// the next reconnect will retry - RunSchemaUpdates has no error return on
// the SeatsurfingPlugin interface itself, so PluginGRPC exposes
// RunSchemaUpdatesErr for exactly this purpose.
func (a *App) registerPluginOnConnect(inst *PluginInstance) {
	log.Printf("Plugin %s: connected, running registration...", inst.Name)

	if err := inst.client.RunSchemaUpdatesErr(); err != nil {
		log.Printf("Plugin %s: RunSchemaUpdates failed, will retry on next reconnect: %v", inst.Name, err)
		return
	}

	AddUnauthorizedRoutes(inst.Instance.GetUnauthorizedRoutes())
	if inst.registered.CompareAndSwap(false, true) {
		api.RegisterPlugin(inst.Instance)
	}

	if err := inst.client.OnInitErr(); err != nil {
		log.Printf("Plugin %s: OnInit failed, will retry on next reconnect: %v", inst.Name, err)
		return
	}

	inst.Ready.Store(true)
	log.Printf("Plugin %s: ready", inst.Name)
}

// hostAPIImpl is the host-side implementation of pluginapi.HostAPI.
// It wraps the real repository singletons and utility functions.
type hostAPIImpl struct{}

func (h *hostAPIImpl) GetSettingsRepository() api.SettingsRepository {
	return GetSettingsRepository()
}
func (h *hostAPIImpl) GetUserRepository() api.UserRepository {
	return GetUserRepository()
}
func (h *hostAPIImpl) GetOrganizationRepository() api.OrganizationRepository {
	return GetOrganizationRepository()
}
func (h *hostAPIImpl) GetGroupRepository() api.GroupRepository {
	return GetGroupRepository()
}
func (h *hostAPIImpl) GetBookingRepository() api.BookingRepository {
	return GetBookingRepository()
}
func (h *hostAPIImpl) GetSpaceRepository() api.SpaceRepository {
	return GetSpaceRepository()
}
func (h *hostAPIImpl) GetLocationRepository() api.LocationRepository {
	return GetLocationRepository()
}
func (h *hostAPIImpl) GetAuthProviderRepository() api.AuthProviderRepository {
	return GetAuthProviderRepository()
}
func (h *hostAPIImpl) GetAuthStateRepository() api.AuthStateRepository {
	return GetAuthStateRepository()
}
func (h *hostAPIImpl) SendEmail(recipient, subject, body, language, orgID string) error {
	return SendEmailWithBodyAndOrg(&MailAddress{Address: recipient}, subject, body, language, orgID)
}
func (h *hostAPIImpl) Encrypt(plaintext string) (string, error) {
	return EncryptString(plaintext)
}
func (h *hostAPIImpl) Decrypt(ciphertext string) (string, error) {
	return DecryptString(ciphertext)
}
func (h *hostAPIImpl) IsValidLanguageCode(code string) bool {
	return GetConfig().IsValidLanguageCode(code)
}
func (h *hostAPIImpl) DisablePasswordLogin() bool {
	return GetConfig().DisablePasswordLogin
}
func (h *hostAPIImpl) FormatPublicURL(domain string) string {
	return FormatURL(domain)
}
func (h *hostAPIImpl) IsDevelopmentMode() bool {
	return GetConfig().Development
}
func (h *hostAPIImpl) GetPostgresURL() string {
	return GetConfig().PostgresURL
}

func (a *App) forwardToPlugin(inst *PluginInstance, w http.ResponseWriter, r *http.Request) {
	if !inst.Ready.Load() {
		http.Error(w, "plugin unavailable", http.StatusServiceUnavailable)
		return
	}
	plg := inst.Instance
	var body []byte
	if r.Body != nil {
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading request body", http.StatusInternalServerError)
			return
		}
	}

	userID := ""
	if u := GetRequestUser(r); u != nil {
		userID = u.ID
	}

	req := api.PluginHTTPRequest{
		Method:   r.Method,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
		Headers:  map[string][]string(r.Header),
		Body:     body,
		UserID:   userID,
	}

	resp := plg.HandleHTTPRequest(req)

	for key, vals := range resp.Headers {
		for _, v := range vals {
			w.Header().Add(key, v)
		}
	}
	if resp.StatusCode != 0 {
		w.WriteHeader(resp.StatusCode)
	}
	w.Write(resp.Body)
}

// NotifyPlugins starts the HostAPI gRPC server (so plugins have somewhere to
// call back into once they connect) and then kicks off connection and
// registration for every configured plugin. Called from main.go only after
// the host's own database migrations/default org/settings are initialized,
// so a plugin's OnInit can never race ahead of the host's own startup.
func (a *App) NotifyPlugins() {
	a.StartHostAPIGRPCServer()
	for _, inst := range a.PluginInstances {
		inst.Conn.Connect()
		go a.watchPluginConnection(inst)
	}
}

// StartHostAPIGRPCServer binds GetConfig().HostAPIListenAddr and serves
// HostAPI over gRPC so any number of plugin processes can dial in and call
// back into the host. A no-op when no plugins are configured.
func (a *App) StartHostAPIGRPCServer() {
	if len(GetConfig().Plugins) == 0 {
		return
	}
	lis, err := net.Listen("tcp", GetConfig().HostAPIListenAddr)
	if err != nil {
		log.Println("Error starting HostAPI gRPC listener:", err)
		return
	}
	a.hostAPIGRPCServer = grpc.NewServer(
		grpc.UnaryInterceptor(api.TokenAuthUnaryInterceptor(GetConfig().HostAPIToken)),
		grpc.MaxRecvMsgSize(GetConfig().PluginMaxMsgSize),
		grpc.MaxSendMsgSize(GetConfig().PluginMaxMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    20 * time.Second,
			Timeout: 10 * time.Second,
		}),
	)
	hostapipb.RegisterHostAPIServiceServer(a.hostAPIGRPCServer, api.NewHostAPIGRPCServer(&hostAPIImpl{}))
	go func() {
		log.Println("HostAPI gRPC server listening on", GetConfig().HostAPIListenAddr)
		if err := a.hostAPIGRPCServer.Serve(lis); err != nil {
			log.Println("HostAPI gRPC server stopped:", err)
		}
	}()
}

func (a *App) KillPlugins() {
	log.Println("Killing plugins …")
	for _, plg := range a.PluginInstances {
		log.Println("Killing plugin", plg.Name)
		plg.Conn.Close()
	}
	if a.hostAPIGRPCServer != nil {
		a.hostAPIGRPCServer.GracefulStop()
	}
}

type notFoundResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *notFoundResponseWriter) WriteHeader(status int) {
	w.status = status
	if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *notFoundResponseWriter) Write(b []byte) (int, error) {
	if w.status == http.StatusNotFound || w.status == http.StatusMethodNotAllowed {
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

func (a *App) globalNotFoundMiddleware(next http.Handler) http.Handler {

	content404, content404Err := os.ReadFile(GetConfig().StaticUiPath + "/404.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := &notFoundResponseWriter{ResponseWriter: w, status: 0}
		next.ServeHTTP(wrapped, r)

		isNotFound := wrapped.status == http.StatusNotFound || wrapped.status == http.StatusMethodNotAllowed
		if !isNotFound || r.URL.Path == "/ui/404/" {
			return
		}

		if GetConfig().Development {
			a.proxyHandler(w, r, "localhost:3000/ui/404/")
			return
		}
		if content404Err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write(content404)
	})
}

func pluginPrefixConflicts(pluginPrefix string, builtInPrefixes []string) bool {
	for _, b := range builtInPrefixes {
		if strings.HasPrefix(pluginPrefix, b) || strings.HasPrefix(b, pluginPrefix) {
			return true
		}
	}
	return false
}

func (a *App) InitializeRouter() {
	a.Router = mux.NewRouter()
	routers := make(map[string]api.Route)
	routers["/location/{locationId}/space/"] = &SpaceRouter{}
	routers["/location/"] = &LocationRouter{}
	routers["/booking/"] = &BookingRouter{}
	routers["/buddy/"] = &BuddyRouter{}
	routers["/organization/"] = &OrganizationRouter{}
	routers["/auth-provider/"] = &AuthProviderRouter{}
	routers["/auth/"] = &AuthRouter{}
	routers["/group/"] = &GroupRouter{}
	routers["/user/"] = &UserRouter{}
	routers["/preference/"] = &UserPreferencesRouter{}
	routers["/recurring-booking/"] = &RecurringBookingRouter{}
	routers["/stats/"] = &StatsRouter{}
	routers["/search/"] = &SearchRouter{}
	routers["/setting/"] = &SettingsRouter{}
	routers["/space-attribute/"] = &SpaceAttributeRouter{}
	routers["/confluence/"] = &ConfluenceRouter{}
	routers["/uc/"] = &CheckUpdateRouter{}
	routers["/healthcheck"] = &HealthcheckRouter{}
	routers["/kiosk/"] = &KioskRouter{}
	builtInPrefixes := make([]string, 0, len(routers))
	for route, router := range routers {
		builtInPrefixes = append(builtInPrefixes, route)
		subRouter := a.Router.PathPrefix(route).Subrouter()
		router.SetupRoutes(subRouter)
	}
	// Route prefixes come from PLUGINS_CONFIG, not a live call to the plugin,
	// so the router is complete even when a plugin is unreachable at startup
	// - forwardToPlugin's Ready gate serves 503 for those routes until the
	// plugin connects (see the connect-driven lifecycle in InitializePlugins/
	// NotifyPlugins).
	for _, inst := range a.PluginInstances {
		for _, prefix := range inst.Config.RoutePrefixes {
			if pluginPrefixConflicts(prefix, builtInPrefixes) {
				log.Printf("Plugin route prefix %q conflicts with a built-in route and will not be registered", prefix)
				continue
			}
			prefix := prefix
			inst := inst
			subRouter := a.Router.PathPrefix(prefix).Subrouter()
			subRouter.Methods("OPTIONS").PathPrefix("/").HandlerFunc(CorsHandler)
			subRouter.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				a.forwardToPlugin(inst, w, r)
			})
		}
	}
	a.setupStaticUIRoutes(a.Router)
	//a.Router.Path("/robots.txt").Methods("GET").HandlerFunc(a.RobotsTxtHandler)
	a.Router.PathPrefix("/admin/").Methods("GET").HandlerFunc(a.RedirectAdminPath)
	a.Router.Path("/admin").Methods("GET").HandlerFunc(a.RedirectAdminPath)
	a.Router.Path("/").Methods("GET").HandlerFunc(a.RedirectRootPath)
	a.Router.PathPrefix("/").Methods("OPTIONS").HandlerFunc(CorsHandler)
	a.Router.Use(SecurityHeaderMiddleware)
	a.Router.Use(VerifyAuthMiddleware)
	a.Router.Use(GetRateLimiterMiddleware())
	notFoundBase := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	a.Router.MethodNotAllowedHandler = a.globalNotFoundMiddleware(notFoundBase)
	a.Router.Use(a.globalNotFoundMiddleware)
}

func (a *App) RobotsTxtHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("User-agent: *\nDisallow: /\n"))
}

func (a *App) RedirectRootPath(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "/ui/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *App) RedirectAdminPath(w http.ResponseWriter, r *http.Request) {
	// Extract the path after /admin and redirect to /ui/admin with the same path
	adminPath := strings.TrimPrefix(r.URL.Path, "/admin")
	if adminPath == "" || adminPath == "/" {
		adminPath = "/"
	}
	redirectURL := "/ui/admin" + adminPath
	if r.URL.RawQuery != "" {
		redirectURL += "?" + r.URL.RawQuery
	}
	w.Header().Set("Location", redirectURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *App) InitializeDefaultOrg() {
	numOrgs, err := GetOrganizationRepository().GetNumOrgs()
	if err == nil && numOrgs == 0 {
		log.Println("Creating default organization...")
		config := GetConfig()
		domain := config.InitOrgDomain
		email := config.InitOrgUser + "@" + domain
		if domain == "localhost" {
			email = config.InitOrgUser + "@" + "seatsurfing.local"
		}
		org := &api.Organization{
			Name:             config.InitOrgName,
			ContactEmail:     email,
			ContactFirstname: "Organization",
			ContactLastname:  "Admin",
			Language:         strings.ToLower(config.InitOrgLanguage),
			SignupDate:       time.Now().UTC(),
		}
		GetOrganizationRepository().Create(org)
		GetOrganizationRepository().AddDomain(org, domain, true)
		GetOrganizationRepository().SetPrimaryDomain(org, domain)
		user := &api.User{
			OrganizationID: org.ID,
			Email:          email,
			HashedPassword: api.NullString(GetUserRepository().GetHashedPassword(config.InitOrgPass)),
			Role:           api.UserRoleOrgAdmin,
			Firstname:      "Organization",
			Lastname:       "Admin",
		}
		GetUserRepository().Create(user)
		GetOrganizationRepository().CreateSampleData(org)
	}
}

func (a *App) InitializeSingleOrgSettings() {
	numOrgs, err := GetOrganizationRepository().GetNumOrgs()
	if err == nil && numOrgs == 1 {
		log.Println("Updating settings for primary organization …")
		orgs, err := GetOrganizationRepository().GetAll()
		if err != nil {
			log.Println("Error while getting primary organization:", err)
			return
		}
		org := orgs[0]
		GetSettingsRepository().Set(org.ID, api.SettingFeatureNoUserLimit.Name, "1")
		GetSettingsRepository().Set(org.ID, api.SettingFeatureCustomDomains.Name, "1")
		GetSettingsRepository().Set(org.ID, api.SettingFeatureGroups.Name, "1")
		GetSettingsRepository().Set(org.ID, api.SettingFeatureAuthProviders.Name, "1")
		GetSettingsRepository().Set(org.ID, api.SettingFeatureRecurringBookings.Name, "1")
		GetSettingsRepository().Set(org.ID, api.SettingFeatureKioskMode.Name, "1")
	}
}

func (a *App) onTimerTick() {
	if err := GetAuthStateRepository().DeleteExpired(); err != nil {
		log.Println(err)
	}
	if err := GetRefreshTokenRepository().DeleteExpired(); err != nil {
		log.Println(err)
	}
	if err := GetSessionRepository().DeleteExpired(); err != nil {
		log.Println(err)
	}
	if err := GetUserRepository().EnableUsersWithExpiredBan(); err != nil {
		log.Println(err)
	}
	num, err := GetUserRepository().DeleteObsoleteConfluenceAnonymousUsers()
	if err != nil {
		log.Println(err)
	}
	if num > 0 {
		log.Printf("Deleted %d anonymous Confluence users", num)
	}

	// purge max. 100 bookings after retention period (if enabled)
	num, err = GetBookingRepository().PurgeOldBookings(100)
	if err != nil {
		log.Println(err)
	}
	if num > 0 {
		log.Printf("Purged %d old bookings", num)
	}

	// send booking reminder emails (~24h before booking start)
	// run every 5 minutes to not cause overlapping runs
	if time.Now().Minute()%5 == 0 {
		go a.sendBookingReminders()
	}

	for _, inst := range a.PluginInstances {
		inst.Instance.OnTimer()
	}
	// Check domain accessibility once per hour
	if time.Now().Minute() == 0 {
		go a.CheckDomainAccessibilityTimer()
	}
	// Update install stats once per hour
	if time.Now().Minute() == 0 {
		go a.UpdateInstallStats()
	}
}

var bookingReminderMu sync.Mutex

func (a *App) sendBookingReminders() {
	bookingReminderMu.Lock()
	defer bookingReminderMu.Unlock()

	bookings, err := GetBookingRepository().GetBookingsDueForReminder(25)
	if err != nil {
		log.Println(err)
		return
	}
	var wg sync.WaitGroup
	for _, booking := range bookings {
		wg.Add(1)
		go func(b *api.BookingDetails) {
			defer wg.Done()
			a.sendBookingReminderEmail(b)
		}(booking)
	}
	wg.Wait()

	num := len(bookings)
	if num > 0 {
		log.Printf("Sent %d booking reminder emails", num)
	}
}

func (a *App) sendBookingReminderEmail(e *api.BookingDetails) {
	active, err := GetUserPreferencesRepository().GetBool(e.UserID, PreferenceMailReminder.Name)
	if err != nil || !active {
		return
	}
	org, err := GetOrganizationRepository().GetOne(e.Space.Location.OrganizationID)
	if err != nil || org == nil {
		log.Println(err)
		return
	}
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		log.Println(err)
		return
	}
	recipientName := e.UserFirstname
	if recipientName == "" {
		recipientName = GetLocalPartFromEmailAddress(e.UserEmail)
	}
	subject := e.Subject
	if subject == "" {
		subject = "—"
	}
	vars := map[string]string{
		"orgDomain":     FormatURL(domain.DomainName) + "/",
		"recipientName": recipientName,
		"date":          e.Enter.Format("2006-01-02 15:04") + " - " + e.Leave.Format("2006-01-02 15:04"),
		"areaName":      e.Space.Location.Name,
		"spaceName":     e.Space.Name,
		"subject":       subject,
	}
	language := org.Language
	if userLang, err := GetUserPreferencesRepository().Get(e.UserID, PreferenceMailLanguage.Name); err == nil && userLang != "" {
		language = userLang
	}
	if err := SendEmailWithOrg(&MailAddress{Address: e.UserEmail}, GetEmailTemplatePathBookingReminder(), language, vars, org.ID); err != nil {
		log.Println(err)
		return
	}

	now := time.Now().UTC()
	if err := GetBookingRepository().SetReminderSent(e.ID, &now); err != nil {
		log.Println(err)
	}
}

func (a *App) InitializeTimers() {
	a.UpdateInstallStats()
	a.onTimerTick()
	if GetConfig().DisableVersionCheck {
		log.Println("ℹ️  Version check is disabled.")
	} else {
		installID, _ := GetSettingsRepository().GetGlobalString(api.SettingInstallID.Name)
		go GetUpdateChecker().InitializeVersionUpdateTimer(installID)
	}
	a.CleanupTicker = time.NewTicker(time.Minute * 1)
	go func() {
		for {
			<-a.CleanupTicker.C
			a.onTimerTick()
		}
	}()
}

func (a *App) UpdateInstallStats() {
	if GetConfig().DisableAnonymousUsageStats {
		return
	}
	stats := &InstallStats{}
	stats.NumLocations, _ = GetLocationRepository().GetCountAll()
	stats.NumUsers, _ = GetUserRepository().GetCountAll()
	stats.NumBookings, _ = GetBookingRepository().GetCountAll()
	stats.NumSpaces, _ = GetSpaceRepository().GetCountAll()
	SetInstallStats(stats)
}

func (a *App) CheckDomainAccessibilityTimer() {
	domains, err := GetOrganizationRepository().GetAllDomains()
	if err != nil {
		log.Println(err)
		return
	}
	for _, domain := range domains {
		if strings.HasSuffix(domain.DomainName, ".seatsurfing.app") || strings.HasSuffix(domain.DomainName, ".seatsurfing.io") {
			GetOrganizationRepository().SetDomainAccessibility(domain.OrganizationID, domain.DomainName, true, time.Now().UTC())
			continue
		}
		success, err := IsDomainAccessible(domain.DomainName, domain.OrganizationID)
		if err != nil {
			log.Println("Error while performing domain accessibility check for domain:", domain.DomainName, err)
			continue
		}
		if !success {
			log.Println("Domain is not accessible:", domain.DomainName)
			GetOrganizationRepository().SetDomainAccessibility(domain.OrganizationID, domain.DomainName, false, time.Now().UTC())
			continue
		}
		GetOrganizationRepository().SetDomainAccessibility(domain.OrganizationID, domain.DomainName, true, time.Now().UTC())
	}
}

func (a *App) attributePathHandler(fs http.Handler, prefix, path string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = path
		r2.URL.Path = strings.TrimPrefix(r2.URL.Path, prefix)
		fs.ServeHTTP(w, r2)
	})
}

func (a *App) stripStaticPrefix(fs http.Handler, prefix string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		fs.ServeHTTP(w, r2)
	})
}

func (a *App) proxyHandler(w http.ResponseWriter, r *http.Request, backend string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	url := fmt.Sprintf("%s://%s%s", "http", backend, r.RequestURI)
	proxyReq, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	proxyReq.Header = make(http.Header)
	for h, val := range r.Header {
		proxyReq.Header[h] = val
	}
	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	bodyRes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for h, vals := range resp.Header {
		for _, val := range vals {
			w.Header().Set(h, val)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyRes)
}

func (a *App) setupStaticUIRoutes(router *mux.Router) {
	const basePath = "/ui"
	if GetConfig().Development {
		router.PathPrefix(basePath + "/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.proxyHandler(w, r, "localhost:3000")
		})
		return
	}
	attributesPaths := a.getAttributePaths(GetConfig().StaticUiPath)
	fs := http.FileServer(neuteredFileSystem{http.Dir(GetConfig().StaticUiPath)})
	for _, attrPath := range attributesPaths {
		path := strings.ReplaceAll(attrPath, "[", "{")
		path = strings.ReplaceAll(path, "]", "}/")
		router.Path(basePath + path).Handler(a.attributePathHandler(fs, basePath+"/", basePath+attrPath+"/"))
	}
	router.Path(basePath + "/").Handler(a.stripStaticPrefix(fs, basePath+"/"))
	router.PathPrefix(basePath + "/").Handler(http.StripPrefix(basePath+"/", fs))
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(p string) (http.File, error) {
	f, err := nfs.fs.Open(p)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if s.IsDir() {
		if _, err := nfs.fs.Open(path.Join(p, "index.html")); err != nil {
			f.Close()
			return nil, os.ErrNotExist
		}
	}
	return f, nil
}

func (a *App) getAttributePaths(dir string) []string {
	b, err := os.ReadFile(dir + "_attr.json")
	if err != nil {
		return []string{}
	}
	var res []string
	json.Unmarshal(b, &res)
	return res
}

func (a *App) startPublicHttpServer() {
	log.Println("Initializing Public REST services …")
	a.PublicHttpServer = &http.Server{
		Addr:         GetConfig().PublicListenAddr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.Router,
	}
	go func() {
		if err := a.PublicHttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
			os.Exit(-1)
		}
	}()
	log.Println("Public HTTP Server listening on", GetConfig().PublicListenAddr)
}

func (a *App) Run() {
	a.startPublicHttpServer()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	a.PublicHttpServer.Shutdown(ctx)
	a.KillPlugins()
}
