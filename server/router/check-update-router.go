package router

import (
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/util"
)

type CheckUpdateRouter struct {
}

func (router *CheckUpdateRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/", router.checkUpdate).Methods("GET")
}

func (router *CheckUpdateRouter) checkUpdate(w http.ResponseWriter, r *http.Request) {
	latest := GetUpdateChecker().Latest
	if latest == nil {
		SendNotFound(w)
		return
	}
	SendJSON(w, latest)
}
