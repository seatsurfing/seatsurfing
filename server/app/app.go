package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/config"
	"github.com/seatsurfing/seatsurfing/server/plugin"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

var _appInstance *App
var _appOnce sync.Once

func GetApp() *App {
	_appOnce.Do(func() {
		_appInstance = &App{}
	})
	return _appInstance
}

type App struct {
	Router           *mux.Router
	PublicHttpServer *http.Server
	CleanupTicker    *time.Ticker
}

func (a *App) InitializeDatabases() {
	RunDBSchemaUpdates()
	InitDefaultOrgSettings()
	InitDefaultUserPreferences()
	// Set up email logging callback
	SetEmailLogCallback(func(subject, recipient, organizationID string) error {
		return GetMailLogRepository().LogEmail(subject, recipient, organizationID)
	})
}

func (a *App) InitializePlugins() {
	for _, plg := range plugin.GetPlugins() {
		(*plg).OnInit()
	}
}

func (a *App) InitializeRouter() {
	a.Router = mux.NewRouter()
	routers := make(map[string]Route)
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
	for _, plg := range plugin.GetPlugins() {
		for route, router := range (*plg).GetPublicRoutes() {
			routers[route] = router
		}
	}
	for route, router := range routers {
		subRouter := a.Router.PathPrefix(route).Subrouter()
		router.SetupRoutes(subRouter)
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
		email := config.InitOrgUser + "@seatsurfing.local"
		org := &Organization{
			Name:             config.InitOrgName,
			ContactEmail:     email,
			ContactFirstname: "Organization",
			ContactLastname:  "Admin",
			Language:         strings.ToLower(config.InitOrgLanguage),
			SignupDate:       time.Now().UTC(),
		}
		GetOrganizationRepository().Create(org)
		user := &User{
			OrganizationID: org.ID,
			Email:          email,
			HashedPassword: NullString(GetUserRepository().GetHashedPassword(config.InitOrgPass)),
			Role:           UserRoleOrgAdmin,
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
		log.Println("Updating settings for primary organization...")
		orgs, err := GetOrganizationRepository().GetAll()
		if err != nil {
			log.Println("Error while getting primary organization:", err)
			return
		}
		org := orgs[0]
		GetSettingsRepository().Set(org.ID, SettingFeatureNoUserLimit.Name, "1")
		GetSettingsRepository().Set(org.ID, SettingFeatureCustomDomains.Name, "1")
		GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")
		GetSettingsRepository().Set(org.ID, SettingFeatureAuthProviders.Name, "1")
		GetSettingsRepository().Set(org.ID, SettingFeatureRecurringBookings.Name, "1")
	}
}

func (a *App) onTimerTick() {
	if err := GetAuthStateRepository().DeleteExpired(); err != nil {
		log.Println(err)
	}
	if err := GetRefreshTokenRepository().DeleteExpired(); err != nil {
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

	for _, plg := range plugin.GetPlugins() {
		(*plg).OnTimer()
	}
	// Check domain accessibility once per hour
	if time.Now().Minute() == 0 {
		go a.CheckDomainAccessibilityTimer()
	}
}

func (a *App) InitializeTimers() {
	a.onTimerTick()
	installID, _ := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	GetUpdateChecker().InitializeVersionUpdateTimer(installID)
	a.CleanupTicker = time.NewTicker(time.Minute * 1)
	go func() {
		for {
			<-a.CleanupTicker.C
			a.onTimerTick()
		}
	}()
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
	fs := http.FileServer(http.Dir(GetConfig().StaticUiPath))
	for _, attrPath := range attributesPaths {
		path := strings.ReplaceAll(attrPath, "[", "{")
		path = strings.ReplaceAll(path, "]", "}/")
		router.Path(basePath + path).Handler(a.attributePathHandler(fs, basePath+"/", basePath+attrPath+"/"))
	}
	router.Path(basePath + "/").Handler(a.stripStaticPrefix(fs, basePath+"/"))
	router.PathPrefix(basePath + "/").Handler(http.StripPrefix(basePath+"/", fs))
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
	log.Println("Initializing Public REST services...")
	a.PublicHttpServer = &http.Server{
		Addr:         GetConfig().PublicListenAddr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.Router,
	}
	go func() {
		if err := a.PublicHttpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}
	}()
	log.Println("Public HTTP Server listening on", GetConfig().PublicListenAddr)
}

func (a *App) Run() {
	a.startPublicHttpServer()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	a.PublicHttpServer.Shutdown(ctx)
}
