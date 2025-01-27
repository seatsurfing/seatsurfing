package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
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
	Router        *mux.Router
	CleanupTicker *time.Ticker
}

func (a *App) InitializeDatabases() {
	RunDBSchemaUpdates()
	InitDefaultOrgSettings()
	InitDefaultUserPreferences()
}

func (a *App) InitializeRouter() {
	config := GetConfig()
	a.Router = mux.NewRouter()
	routers := make(map[string]Route)
	routers["/location/{locationId}/space/"] = &SpaceRouter{}
	routers["/location/"] = &LocationRouter{}
	routers["/booking/"] = &BookingRouter{}
	routers["/buddy/"] = &BuddyRouter{}
	routers["/organization/"] = &OrganizationRouter{}
	routers["/auth-provider/"] = &AuthProviderRouter{}
	routers["/auth/"] = &AuthRouter{}
	routers["/user/"] = &UserRouter{}
	routers["/preference/"] = &UserPreferencesRouter{}
	routers["/stats/"] = &StatsRouter{}
	routers["/search/"] = &SearchRouter{}
	routers["/setting/"] = &SettingsRouter{}
	routers["/space-attribute/"] = &SpaceAttributeRouter{}
	routers["/confluence/"] = &ConfluenceRouter{}
	routers["/uc/"] = &CheckUpdateRouter{}
	if config.OrgSignupEnabled {
		routers["/signup/"] = &SignupRouter{}
	}
	for route, router := range routers {
		subRouter := a.Router.PathPrefix(route).Subrouter()
		router.setupRoutes(subRouter)
	}
	if !GetConfig().DisableUiProxy {
		a.setupBookingUIProxy(a.Router)
		a.setupAdminUIProxy(a.Router)
	}
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
		GetSettingsRepository().Set(org.ID, SettingSubscriptionMaxUsers.Name, "10000")
		GetOrganizationRepository().AddDomain(org, config.InitOrgDomain, true)
		user := &User{
			OrganizationID: org.ID,
			Email:          config.InitOrgUser + "@" + config.InitOrgDomain,
			HashedPassword: NullString(GetUserRepository().GetHashedPassword(config.InitOrgPass)),
			Role:           UserRoleSuperAdmin,
		}
		GetUserRepository().Create(user)
		GetOrganizationRepository().createSampleData(org)
	}
}

func (a *App) InitializeTimers() {
	GetUpdateChecker().InitializeVersionUpdateTimer()
	a.CleanupTicker = time.NewTicker(time.Minute * 1)
	go func() {
		for {
			<-a.CleanupTicker.C
			log.Println("Cleaning up expired database entries...")
			if err := GetAuthStateRepository().DeleteExpired(); err != nil {
				log.Println(err)
			}
			if err := GetSignupRepository().DeleteExpired(); err != nil {
				log.Println(err)
			}
			if err := GetRefreshTokenRepository().DeleteExpired(); err != nil {
				log.Println(err)
			}
			if err := GetUserRepository().enableUsersWithExpiredBan(); err != nil {
				log.Println(err)
			}
			num, err := GetUserRepository().DeleteObsoleteConfluenceAnonymousUsers()
			if err != nil {
				log.Println(err)
			}
			if num > 0 {
				log.Printf("Deleted %d anonymous Confluence users", num)
			}
		}
	}()
}

func (a *App) bookingUIProxyHandler(w http.ResponseWriter, r *http.Request) {
	a.proxyHandler(w, r, GetConfig().BookingUiBackend)
}

func (a *App) adminUIProxyHandler(w http.ResponseWriter, r *http.Request) {
	a.proxyHandler(w, r, GetConfig().AdminUiBackend)
}

func (a *App) proxyHandler(w http.ResponseWriter, r *http.Request, backend string) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
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
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	bodyRes, err := ioutil.ReadAll(resp.Body)
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

func (a *App) setupBookingUIProxy(router *mux.Router) {
	const basePath = "/ui"
	router.Path(basePath).HandlerFunc(a.bookingUIProxyHandler)
	router.Path(basePath + "/").HandlerFunc(a.bookingUIProxyHandler)
	router.PathPrefix(basePath + "/").HandlerFunc(a.bookingUIProxyHandler)
}

func (a *App) setupAdminUIProxy(router *mux.Router) {
	const basePath = "/admin"
	router.Path(basePath).HandlerFunc(a.adminUIProxyHandler)
	router.Path(basePath + "/").HandlerFunc(a.adminUIProxyHandler)
	router.PathPrefix(basePath + "/").HandlerFunc(a.adminUIProxyHandler)
}

func (a *App) Run(publicListenAddr string) {
	log.Println("Initializing REST services...")
	httpServer := &http.Server{
		Addr:         publicListenAddr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.Router,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}
	}()
	log.Println("HTTP Server listening on", publicListenAddr)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	httpServer.Shutdown(ctx)
}
