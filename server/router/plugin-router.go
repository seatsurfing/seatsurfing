package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/seatsurfing/seatsurfing/server/plugin"
)

type PluginRouter struct {
}

type PluginRouterAdminMenuItem struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Source     string `json:"src"`
	Visibility string `json:"visibility"`
	Icon       string `json:"icon"`
}

func (router *PluginRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/admin-menu-items/", router.getAdminMenuItems).Methods("GET")
}

func (router *PluginRouter) getAdminMenuItems(w http.ResponseWriter, r *http.Request) {
	res := []PluginRouterAdminMenuItem{}
	for _, plg := range plugin.GetPlugins() {
		for _, item := range (*plg).GetAdminUIMenuItems() {
			resItem := PluginRouterAdminMenuItem{
				ID:         item.ID,
				Title:      item.Title,
				Source:     item.Source,
				Visibility: item.Visibility,
				Icon:       item.Icon,
			}
			res = append(res, resItem)
		}
	}
	SendJSON(w, res)
}
