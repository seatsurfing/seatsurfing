package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type LocationRouter struct {
}

type CreateLocationRequest struct {
	Name                  string `json:"name" validate:"required"`
	Description           string `json:"description"`
	MaxConcurrentBookings uint   `json:"maxConcurrentBookings"`
	Timezone              string `json:"timezone"`
	Enabled               bool   `json:"enabled"`
}

type GetLocationResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	MapWidth       uint   `json:"mapWidth"`
	MapHeight      uint   `json:"mapHeight"`
	MapMimeType    string `json:"mapMimeType"`
	CreateLocationRequest
}

type GetMapResponse struct {
	Width    uint   `json:"width"`
	Height   uint   `json:"height"`
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type SetSpaceAttributeValueRequest struct {
	Value string `json:"value"`
}

type GetSpaceAttributeValueResponse struct {
	AttributeID string `json:"attributeId"`
	Value       string `json:"value"`
}

type SearchLocationRequest struct {
	Enter      time.Time         `json:"enter" validate:"required"`
	Leave      time.Time         `json:"leave" validate:"required"`
	Attributes []SearchAttribute `json:"attributes"`
}

type SearchAttribute struct {
	AttributeID string `json:"attributeId"`
	Comparator  string `json:"comparator"`
	Value       string `json:"value"`
}

const (
	SearchAttributeNumSpaces     string = "numSpaces"
	SearchAttributeNumFreeSpaces string = "numFreeSpaces"
	SearchAttributeBuddyOnSite   string = "buddyOnSite"
)

func (router *LocationRouter) setupRoutes(s *mux.Router) {
	s.HandleFunc("/search", router.search).Methods("POST")
	s.HandleFunc("/loadsampledata", router.loadSampleData).Methods("POST")
	s.HandleFunc("/{id}/attribute", router.getAttributes).Methods("GET")
	s.HandleFunc("/{id}/attribute/{attributeId}", router.setAttribute).Methods("POST")
	s.HandleFunc("/{id}/attribute/{attributeId}", router.deleteAttribute).Methods("DELETE")
	s.HandleFunc("/{id}/map", router.getMap).Methods("GET")
	s.HandleFunc("/{id}/map", router.setMap).Methods("POST")
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *LocationRouter) getAttributes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetLocationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	list, err := GetSpaceAttributeValueRepository().GetAllForEntity(e.ID, SpaceAttributeValueEntityTypeLocation)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetSpaceAttributeValueResponse{}
	for _, val := range list {
		m := &GetSpaceAttributeValueResponse{
			AttributeID: val.AttributeID,
			Value:       val.Value,
		}
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *LocationRouter) setAttribute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetLocationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	attribute, err := GetSpaceAttributeRepository().GetOne(vars["attributeId"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	if !attribute.LocationApplicable {
		SendBadRequest(w)
		return
	}
	var m SetSpaceAttributeValueRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	if err := GetSpaceAttributeValueRepository().Set(attribute.ID, e.ID, SpaceAttributeValueEntityTypeLocation, m.Value); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *LocationRouter) deleteAttribute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetLocationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	GetSpaceAttributeValueRepository().Delete(vars["attributeId"], e.ID, SpaceAttributeValueEntityTypeLocation)
	SendUpdated(w)
}

func (router *LocationRouter) getOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetLocationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAccessOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	res := router.copyToRestModel(e)
	SendJSON(w, res)
}

func (router *LocationRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	list, err := GetLocationRepository().GetAll(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetLocationResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (rouer *LocationRouter) matchesSearchAttributes(entityID string, m *[]SearchAttribute, attributeValues []*SpaceAttributeValue) bool {
	var matchString = func(a, b, comparator string) bool {
		if comparator == "eq" {
			return a == b
		} else if comparator == "neq" {
			return a != b
		} else if comparator == "contains" {
			return strings.Contains(a, b)
		} else if comparator == "ncontains" {
			return !strings.Contains(a, b)
		} else if comparator == "gt" {
			searchAttrInt, err := strconv.Atoi(a)
			if err != nil {
				return false
			}
			attrValInt, err := strconv.Atoi(b)
			if err != nil {
				return false
			}
			return searchAttrInt > attrValInt
		} else if comparator == "lt" {
			searchAttrInt, err := strconv.Atoi(a)
			if err != nil {
				return false
			}
			attrValInt, err := strconv.Atoi(b)
			if err != nil {
				return false
			}
			return searchAttrInt < attrValInt
		} else if comparator == "gte" {
			searchAttrInt, err := strconv.Atoi(a)
			if err != nil {
				return false
			}
			attrValInt, err := strconv.Atoi(b)
			if err != nil {
				return false
			}
			return searchAttrInt >= attrValInt
		} else if comparator == "lte" {
			searchAttrInt, err := strconv.Atoi(a)
			if err != nil {
				return false
			}
			attrValInt, err := strconv.Atoi(b)
			if err != nil {
				return false
			}
			return searchAttrInt <= attrValInt
		}
		return false
	}

	var matchArray = func(a []string, b, comparator string) bool {
		if comparator == "contains" {
			if b == "*" {
				return len(a) > 0
			}
			return slices.Contains(a, b)
		} else if comparator == "ncontains" {
			if b == "*" {
				return len(a) == 0
			}
			return !slices.Contains(a, b)
		}
		return false
	}

	for _, searchAttr := range *m {
		found := false
		for _, attrVal := range attributeValues {
			if (attrVal.AttributeID == searchAttr.AttributeID) && (attrVal.EntityID == entityID) {
				if strings.Index(attrVal.Value, "[") == 0 && strings.Index(attrVal.Value, "]") == len(attrVal.Value)-1 {
					var arr []string
					if err := json.Unmarshal([]byte(attrVal.Value), &arr); err != nil {
						log.Println(err)
						return false
					}
					found = matchArray(arr, searchAttr.Value, searchAttr.Comparator)
				} else {
					found = matchString(attrVal.Value, searchAttr.Value, searchAttr.Comparator)
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (router *LocationRouter) searchInputContains(m *[]SearchAttribute, attributeID string) bool {
	for _, e := range *m {
		if e.AttributeID == attributeID {
			return true
		}
	}
	return false
}

func (router *LocationRouter) searchAttachNumSpaces(attributeValues []*SpaceAttributeValue, organizationID string) ([]*SpaceAttributeValue, error) {
	totalSpaces, err := GetSpaceRepository().GetTotalCountMap(organizationID)
	if err != nil {
		return nil, err
	}
	for k, v := range totalSpaces {
		attributeValues = append(attributeValues, &SpaceAttributeValue{
			AttributeID: SearchAttributeNumSpaces,
			EntityID:    k,
			EntityType:  SpaceAttributeValueEntityTypeLocation,
			Value:       strconv.Itoa(v),
		})
	}
	return attributeValues, nil
}

func (router *LocationRouter) searchAttachNumFreeSpaces(attributeValues []*SpaceAttributeValue, organizationID string, enter, leave time.Time) ([]*SpaceAttributeValue, error) {
	freeSpaces, err := GetSpaceRepository().GetFreeCountMap(organizationID, enter, leave)
	if err != nil {
		return nil, err
	}
	for k, v := range freeSpaces {
		attributeValues = append(attributeValues, &SpaceAttributeValue{
			AttributeID: SearchAttributeNumFreeSpaces,
			EntityID:    k,
			EntityType:  SpaceAttributeValueEntityTypeLocation,
			Value:       strconv.Itoa(v),
		})
	}
	return attributeValues, nil
}

func (router *LocationRouter) searchAttachBuddiesOnSite(attributeValues []*SpaceAttributeValue, user *User, enter, leave time.Time) ([]*SpaceAttributeValue, error) {
	buddies, err := GetBuddyRepository().GetAllByOwner(user.ID)
	if err != nil {
		return nil, err
	}
	log.Println(buddies)
	usersOnSite, err := GetSpaceRepository().GetBookingUserIDMap(user.OrganizationID, enter, leave)
	if err != nil {
		return nil, err
	}
	buddiesOnSite := make(map[string][]string)
	for locationID, userIDs := range usersOnSite {
		buddiesOnSite[locationID] = []string{}
		for _, buddy := range buddies {
			if slices.Contains(userIDs, buddy.BuddyID) {
				buddiesOnSite[locationID] = append(buddiesOnSite[locationID], buddy.ID)
			}
		}
	}
	for k, v := range buddiesOnSite {
		json, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		log.Println(string(json))
		attributeValues = append(attributeValues, &SpaceAttributeValue{
			AttributeID: SearchAttributeBuddyOnSite,
			EntityID:    k,
			EntityType:  SpaceAttributeValueEntityTypeLocation,
			Value:       string(json),
		})
	}
	return attributeValues, nil
}

func (router *LocationRouter) search(w http.ResponseWriter, r *http.Request) {
	var m SearchLocationRequest
	if err := UnmarshalValidateBody(r, &m); err != nil {
		log.Println(err)
		SendBadRequest(w)
		return
	}
	if len(m.Attributes) == 0 {
		router.getAll(w, r)
		return
	}
	user := GetRequestUser(r)
	list, err := GetLocationRepository().GetAll(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	attributeValues, err := GetSpaceAttributeValueRepository().GetAll(user.OrganizationID, SpaceAttributeValueEntityTypeLocation)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if router.searchInputContains(&m.Attributes, SearchAttributeNumSpaces) {
		attributeValues, err = router.searchAttachNumSpaces(attributeValues, user.OrganizationID)
		if err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
	}
	if router.searchInputContains(&m.Attributes, SearchAttributeNumFreeSpaces) {
		attributeValues, err = router.searchAttachNumFreeSpaces(attributeValues, user.OrganizationID, m.Enter, m.Leave)
		if err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
	}
	if router.searchInputContains(&m.Attributes, SearchAttributeBuddyOnSite) {
		attributeValues, err = router.searchAttachBuddiesOnSite(attributeValues, user, m.Enter, m.Leave)
		if err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
	}
	res := []*GetLocationResponse{}
	for _, e := range list {
		if router.matchesSearchAttributes(e.ID, &m.Attributes, attributeValues) {
			m := router.copyToRestModel(e)
			res = append(res, m)
		}
	}
	SendJSON(w, res)
}

func (router *LocationRouter) update(w http.ResponseWriter, r *http.Request) {
	var m CreateLocationRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetLocationRepository().GetOne(vars["id"])
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	if m.Timezone != "" {
		if !isValidTimeZone(m.Timezone) {
			SendBadRequest(w)
			return
		}
	}
	eNew := router.copyFromRestModel(&m)
	eNew.ID = e.ID
	eNew.OrganizationID = e.OrganizationID
	if err := GetLocationRepository().Update(eNew); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *LocationRouter) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetLocationRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	if err := GetLocationRepository().Delete(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *LocationRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateLocationRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	e := router.copyFromRestModel(&m)
	e.OrganizationID = user.OrganizationID
	if !CanSpaceAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	if m.Timezone != "" {
		if !isValidTimeZone(m.Timezone) {
			SendBadRequest(w)
			return
		}
	}
	if err := GetLocationRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}

func (router *LocationRouter) getMap(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetLocationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAccessOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	locationMap, err := GetLocationRepository().GetMap(e)
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	res := &GetMapResponse{
		Width:    locationMap.Width,
		Height:   locationMap.Height,
		MimeType: locationMap.MimeType,
		Data:     base64.StdEncoding.EncodeToString(locationMap.Data),
	}
	SendJSON(w, res)
}

func (router *LocationRouter) setMap(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetLocationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		SendBadRequest(w)
		return
	}
	image, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		log.Println(err)
		SendBadRequest(w)
		return
	}
	locationMap := &LocationMap{
		Width:    uint(image.Width),
		Height:   uint(image.Height),
		MimeType: format,
		Data:     data,
	}
	if err := GetLocationRepository().SetMap(e, locationMap); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *LocationRouter) loadSampleData(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	org, err := GetOrganizationRepository().GetOne(user.OrganizationID)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	GetOrganizationRepository().createSampleData(org)
}

func (router *LocationRouter) copyFromRestModel(m *CreateLocationRequest) *Location {
	e := &Location{}
	e.Name = m.Name
	e.Description = m.Description
	e.MaxConcurrentBookings = m.MaxConcurrentBookings
	e.Timezone = m.Timezone
	e.Enabled = m.Enabled
	return e
}

func (router *LocationRouter) copyToRestModel(e *Location) *GetLocationResponse {
	m := &GetLocationResponse{}
	m.ID = e.ID
	m.OrganizationID = e.OrganizationID
	m.Name = e.Name
	m.MapMimeType = e.MapMimeType
	m.MapWidth = e.MapWidth
	m.MapHeight = e.MapHeight
	m.Description = e.Description
	m.MaxConcurrentBookings = e.MaxConcurrentBookings
	m.Timezone = e.Timezone
	m.Enabled = e.Enabled
	return m
}
