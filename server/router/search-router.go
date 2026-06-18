package router

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type SearchRouter struct {
}

type GetUserSearchResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type GetLocationSearchResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GetSpaceSearchResponse struct {
	ID       string                     `json:"id"`
	Name     string                     `json:"name"`
	Location *GetLocationSearchResponse `json:"location,omitempty"`
}

type GetGroupSearchResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organizationId"`
}

type GetSearchResultsResponse struct {
	Users     []*GetUserSearchResponse     `json:"users"`
	Locations []*GetLocationSearchResponse `json:"locations"`
	Spaces    []*GetSpaceSearchResponse    `json:"spaces"`
	Groups    []*GetGroupSearchResponse    `json:"groups"`
}

func (router *SearchRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/", router.getResults).Methods("GET")
}

func (router *SearchRouter) getResults(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("query")
	if len(keyword) > 64 {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}

	res := &GetSearchResultsResponse{}
	if r.URL.Query().Get("includeUsers") == "1" {
		if err := router.addUserResults(user, keyword, res); err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
	}
	if r.URL.Query().Get("includeGroups") == "1" {
		if err := router.addGroupResults(user, keyword, res); err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
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
		if err := router.addSpaceResults(user, keyword, r.URL.Query().Get("expandLocations") == "1", res); err != nil {
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
	for _, e := range list {
		m := &GetUserSearchResponse{
			ID:        e.ID,
			Email:     e.Email,
			Firstname: e.Firstname,
			Lastname:  e.Lastname,
		}
		res.Users = append(res.Users, m)
	}
	return nil
}

func (router *SearchRouter) addGroupResults(user *User, keyword string, res *GetSearchResultsResponse) error {
	list, err := GetGroupRepository().GetByKeyword(user.OrganizationID, keyword)
	if err != nil {
		return err
	}
	for _, e := range list {
		m := &GetGroupSearchResponse{
			ID:             e.ID,
			Name:           e.Name,
			OrganizationID: e.OrganizationID,
		}
		res.Groups = append(res.Groups, m)
	}
	return nil
}

func (router *SearchRouter) addLocationResults(user *User, keyword string, res *GetSearchResultsResponse) error {
	list, err := GetLocationRepository().GetByKeyword(user.OrganizationID, keyword)
	if err != nil {
		return err
	}
	for _, e := range list {
		m := &GetLocationSearchResponse{
			ID:          e.ID,
			Name:        e.Name,
			Description: e.Description,
		}
		res.Locations = append(res.Locations, m)
	}
	return nil
}

func (router *SearchRouter) addSpaceResults(user *User, keyword string, expandLocations bool, res *GetSearchResultsResponse) error {
	list, err := GetSpaceRepository().GetByKeyword(user.OrganizationID, keyword)
	if err != nil {
		return err
	}

	var locationMap map[string]*Location
	if expandLocations {
		locations, err := GetLocationRepository().GetAll(user.OrganizationID)
		if err != nil {
			return err
		}
		locationMap = make(map[string]*Location)
		for _, loc := range locations {
			locationMap[loc.ID] = loc
		}
	}

	for _, e := range list {
		m := &GetSpaceSearchResponse{
			ID:   e.ID,
			Name: e.Name,
		}
		if expandLocations {
			if loc, ok := locationMap[e.LocationID]; ok {
				m.Location = &GetLocationSearchResponse{
					ID:          loc.ID,
					Name:        loc.Name,
					Description: loc.Description,
				}
			}
		}
		res.Spaces = append(res.Spaces, m)
	}
	return nil
}
