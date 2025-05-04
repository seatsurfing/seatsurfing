package router

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type SearchRouter struct {
}

type GetSearchResultsResponse struct {
	Users     []*GetUserResponse     `json:"users"`
	Locations []*GetLocationResponse `json:"locations"`
	Spaces    []*GetSpaceResponse    `json:"spaces"`
}

func (router *SearchRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/", router.getResults).Methods("GET")
}

func (router *SearchRouter) getResults(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	keyword := r.URL.Query().Get("query")
	res := &GetSearchResultsResponse{
		Users: []*GetUserResponse{},
	}
	if r.URL.Query().Get("includeUsers") == "1" {
		if CanAdminOrg(user, user.OrganizationID) {
			if err := router.addUserResults(user, keyword, res); err != nil {
				log.Println(err)
				SendInternalServerError(w)
				return
			}
		}
	}
	if r.URL.Query().Get("includeLocations") == "1" {
		if err := router.addLocationResults(user, keyword, res); err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
	}
	if r.URL.Query().Get("includeSpaces") == "1" {
		if err := router.addSpaceResults(user, keyword, res); err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
	}
	SendJSON(w, res)
}

func (router *SearchRouter) addUserResults(user *User, keyword string, res *GetSearchResultsResponse) error {
	list, err := GetUserRepository().GetByKeyword(user.OrganizationID, keyword)
	if err != nil {
		return err
	}
	userRouter := &UserRouter{}
	for _, e := range list {
		m := userRouter.copyToRestModel(e, true)
		res.Users = append(res.Users, m)
	}
	return nil
}

func (router *SearchRouter) addLocationResults(user *User, keyword string, res *GetSearchResultsResponse) error {
	list, err := GetLocationRepository().GetByKeyword(user.OrganizationID, keyword)
	if err != nil {
		return err
	}
	locationRouter := &LocationRouter{}
	for _, e := range list {
		m := locationRouter.copyToRestModel(e)
		res.Locations = append(res.Locations, m)
	}
	return nil
}

func (router *SearchRouter) addSpaceResults(user *User, keyword string, res *GetSearchResultsResponse) error {
	list, err := GetSpaceRepository().GetByKeyword(user.OrganizationID, keyword)
	if err != nil {
		return err
	}
	spaceRouter := &SpaceRouter{}
	for _, e := range list {
		m := spaceRouter.copyToRestModel(e, nil)
		res.Spaces = append(res.Spaces, m)
	}
	return nil
}
