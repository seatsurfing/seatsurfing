package app

import (
	"context"
	"encoding/json"
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
	a.setupStaticAdminUIRoutes(a.Router)
	a.setupStaticBookingUIRoutes(a.Router)
	//a.Router.Path("/robots.txt").Methods("GET").HandlerFunc(a.RobotsTxtHandler)
	a.Router.Path("/").Methods("GET").HandlerFunc(a.RedirectRootPath)
	a.Router.PathPrefix("/").Methods("OPTIONS").HandlerFunc(CorsHandler)
	a.Router.Use(CorsMiddleware)
	a.Router.Use(VerifyAuthMiddleware)
}

func (a *App) RobotsTxtHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("User-agent: *\nDisallow: /\n"))
}

func (a *App) RedirectRootPath(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "/ui/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *App) InitializeDefaultOrg() {
	numOrgs, err := GetOrganizationRepository().GetNumOrgs()
	if err == nil && numOrgs == 0 {
		log.Println("Creating first organization...")
		config := GetConfig()
		org := &Organization{
			Name:       config.InitOrgName,
			Language:   strings.ToLower(config.InitOrgLanguage),
			SignupDate: time.Now().UTC(),
		}
		GetOrganizationRepository().Create(org)
		user := &User{
			OrganizationID: org.ID,
			Email:          config.InitOrgUser + "@seatsurfing.local",
			HashedPassword: NullString(GetUserRepository().GetHashedPassword(config.InitOrgPass)),
			Role:           UserRoleOrgAdmin,
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
			log.Println("Error while getting first organization:", err)
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

func (a *App) InitializeTimers() {
	installID, _ := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	GetUpdateChecker().InitializeVersionUpdateTimer(installID)
	a.CleanupTicker = time.NewTicker(time.Minute * 1)
	go func() {
		for {
			<-a.CleanupTicker.C
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
			for _, plg := range plugin.GetPlugins() {
				(*plg).OnTimer()
			}
			// Check domain accessibility once per hour
			if time.Now().Minute() == 0 {
				go a.CheckDomainAccessibilityTimer()
			}
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

func (a *App) setupStaticBookingUIRoutes(router *mux.Router) {
	const basePath = "/ui"
	attributesPaths := a.getAttributePaths(GetConfig().StaticBookingUiPath)
	fs := http.FileServer(http.Dir(GetConfig().StaticBookingUiPath))
	for _, attrPath := range attributesPaths {
		path := strings.ReplaceAll(attrPath, "[", "{")
		path = strings.ReplaceAll(path, "]", "}/")
		router.Path(basePath + path).Handler(a.attributePathHandler(fs, basePath+"/", basePath+attrPath+"/"))
	}
	router.Path(basePath + "/").Handler(a.stripStaticPrefix(fs, basePath+"/"))
	router.PathPrefix(basePath + "/").Handler(http.StripPrefix(basePath+"/", fs))
}

func (a *App) setupStaticAdminUIRoutes(router *mux.Router) {
	const basePath = "/admin"
	attributesPaths := a.getAttributePaths(GetConfig().StaticAdminUiPath)
	fs := http.FileServer(http.Dir(GetConfig().StaticAdminUiPath))
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
