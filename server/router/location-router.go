package router

import (
	"bytes"
	"database/sql"
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
	"github.com/rustyoz/svg"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	"github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type LocationRouter struct {
}

type CreateLocationRequest struct {
	Name                  string   `json:"name" validate:"required,max=128"`
	Description           string   `json:"description" validate:"max=512"`
	MaxConcurrentBookings uint     `json:"maxConcurrentBookings"`
	Timezone              string   `json:"timezone" validate:"max=32"`
	Enabled               bool     `json:"enabled"`
	MapScale              float64  `json:"mapScale"`
	MapType               string   `json:"mapType" validate:"omitempty,oneof=designed"`
	AllowedBookerGroupIDs []string `json:"allowedBookerGroupIds" validate:"dive,uuid"`
	BookableDays          []int    `json:"bookableDays" validate:"dive,min=0,max=6"`
}

type GetLocationResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	MapWidth       uint   `json:"mapWidth"`
	MapHeight      uint   `json:"mapHeight"`
	MapMimeType    string `json:"mapMimeType"`
	CreateLocationRequest
}

type GetFloorPlanDesignResponse struct {
	DesignData string `json:"designData"`
}

type SetFloorPlanDesignRequest struct {
	DesignData string `json:"designData" validate:"required"`
}

type GetMapResponse struct {
	Width    uint    `json:"width"`
	Height   uint    `json:"height"`
	Scale    float64 `json:"scale"`
	MimeType string  `json:"mimeType"`
	Data     string  `json:"data"`
}

type SetSpaceAttributeValueRequest struct {
	Value string `json:"value" validate:"max=256"`
}

type GetSpaceAttributeValueResponse struct {
	AttributeID string `json:"attributeId"`
	Value       string `json:"value"`
}

type SearchLocationRequest struct {
	Enter      time.Time         `json:"enter" validate:"required"`
	Leave      time.Time         `json:"leave" validate:"required"`
	Attributes []SearchAttribute `json:"attributes" validate:"dive"`
}

const (
	SearchAttributeNumSpaces     = service.SearchAttributeNumSpaces
	SearchAttributeNumFreeSpaces = service.SearchAttributeNumFreeSpaces
	SearchAttributeBuddyOnSite   = service.SearchAttributeBuddyOnSite
)

func (router *LocationRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/search", router.search).Methods("POST")
	s.HandleFunc("/loadsampledata", router.loadSampleData).Methods("POST")
	s.HandleFunc("/{id}/attribute", router.getAttributes).Methods("GET")
	s.HandleFunc("/{id}/attribute/{attributeId}", router.setAttribute).Methods("POST")
	s.HandleFunc("/{id}/attribute/{attributeId}", router.deleteAttribute).Methods("DELETE")
	s.HandleFunc("/{id}/map", router.getMap).Methods("GET")
	s.HandleFunc("/{id}/map", router.setMap).Methods("POST")
	s.HandleFunc("/{id}/floorplan-design", router.getFloorPlanDesign).Methods("GET")
	s.HandleFunc("/{id}/floorplan-design", router.setFloorPlanDesign).Methods("POST")
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
	user := GetRequestUser(r)
	if !CanAccessOrg(user, e.OrganizationID) {
		SendForbidden(w)
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
	if !HasPermission(user, e.OrganizationID, PermissionAreas, PermissionLevelAdmin) {
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
	if !HasPermission(user, e.OrganizationID, PermissionAreas, PermissionLevelAdmin) {
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

	allowedBookers, err := GetLocationRepository().GetAllAllowedBookersForLocation(e.ID)
	res := router.copyToRestModel(e, allowedBookers)
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

	locationIDs := []string{}
	for _, e := range list {
		locationIDs = append(locationIDs, e.ID)
	}
	allowedBookers, err := GetLocationRepository().GetAllAllowedBookersForLocationList(locationIDs)

	res := []*GetLocationResponse{}
	for _, e := range list {
		filteredLocationGroup := []*LocationGroup{}
		for _, ab := range allowedBookers {
			if ab.LocationID == e.ID {
				filteredLocationGroup = append(filteredLocationGroup, ab)
			}
		}
		m := router.copyToRestModel(e, filteredLocationGroup)
		res = append(res, m)
	}
	SendJSON(w, res)
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
	list, err := service.GetLocationService().SearchLocationsForUser(GetRequestUser(r), m.Enter, m.Leave, m.Attributes)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetLocationResponse{}
	for _, e := range list {
		res = append(res, router.copyToRestModel(e.Location, e.AllowedBookers))
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
	if !HasPermission(user, e.OrganizationID, PermissionAreas, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	if m.Timezone != "" {
		if !IsValidTimeZone(m.Timezone) {
			SendBadRequest(w)
			return
		}
	}
	if len(m.BookableDays) > 0 {
		if !IsValidWeekdaysList(weekdaysToString(m.BookableDays)) {
			SendBadRequest(w)
			return
		}
	}
	previousMapType := e.MapType
	eNew := router.copyFromRestModel(&m)
	eNew.ID = e.ID
	eNew.OrganizationID = e.OrganizationID
	// Clear stale floor plan data before flipping map_type so that a failure
	// here does not leave the location in an inconsistent state (map_type
	// already updated but stale map data still present).
	if previousMapType == "designed" && eNew.MapType == "" {
		if err := GetLocationRepository().ClearMapData(eNew); err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
	}
	if err := GetLocationRepository().Update(eNew); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}

	err = GetLocationRepository().ReplaceAllowedBookers(eNew, m.AllowedBookerGroupIDs)
	if err != nil {
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
	if !HasPermission(user, e.OrganizationID, PermissionAreas, PermissionLevelAdmin) {
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
	if !HasPermission(user, e.OrganizationID, PermissionAreas, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	if m.Timezone != "" {
		if !IsValidTimeZone(m.Timezone) {
			SendBadRequest(w)
			return
		}
	}
	if len(m.BookableDays) > 0 {
		if !IsValidWeekdaysList(weekdaysToString(m.BookableDays)) {
			SendBadRequest(w)
			return
		}
	}
	if err := GetLocationRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}

	err := GetLocationRepository().ReplaceAllowedBookers(e, m.AllowedBookerGroupIDs)
	if err != nil {
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
	if e.MapType == "designed" {
		var designData string
		plan, err := GetLocationFloorPlanRepository().GetDesign(e.ID)
		if err == sql.ErrNoRows {
			designData = `{"elements":[]}`
		} else if err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		} else {
			designData = plan.DesignData
		}
		svgData, width, height, err := renderFloorPlanSVG(designData)
		if err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
		res := &GetMapResponse{
			Width:    width,
			Height:   height,
			MimeType: "svg+xml",
			Scale:    1.0,
			Data:     base64.StdEncoding.EncodeToString(svgData),
		}
		SendJSON(w, res)
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
		Scale:    locationMap.Scale,
		Data:     base64.StdEncoding.EncodeToString(locationMap.Data),
	}
	SendJSON(w, res)
}

func (router *LocationRouter) getFloorPlanDesign(w http.ResponseWriter, r *http.Request) {
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
	plan, err := GetLocationFloorPlanRepository().GetDesign(e.ID)
	if err == sql.ErrNoRows {
		// No design record exists yet — return empty design
		SendJSON(w, &GetFloorPlanDesignResponse{DesignData: ""})
		return
	} else if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendJSON(w, &GetFloorPlanDesignResponse{DesignData: plan.DesignData})
}

func (router *LocationRouter) setFloorPlanDesign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetLocationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !HasPermission(user, e.OrganizationID, PermissionAreas, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	var m SetFloorPlanDesignRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	if !json.Valid([]byte(m.DesignData)) {
		SendBadRequest(w)
		return
	}
	plan := &LocationFloorPlan{
		LocationID:     e.ID,
		OrganizationID: e.OrganizationID,
		DesignData:     m.DesignData,
	}
	if err := GetLocationFloorPlanRepository().SetDesign(plan); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	// Render the SVG to compute dimensions and persist them on the location so
	// that map_width / map_height / map_mimetype are kept in sync.
	_, width, height, err := renderFloorPlanSVG(m.DesignData)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	e.MapWidth = width
	e.MapHeight = height
	e.MapMimeType = "svg+xml"
	e.MapScale = 1.0
	if err := GetLocationRepository().SetMapMeta(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
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
	if !HasPermission(user, e.OrganizationID, PermissionAreas, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		SendBadRequest(w)
		return
	}
	// Check if image is PNG, GIF of JPEG
	img, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		// On error, check is image is SVG
		parsedSvg, err := svg.ParseSvg(string(data), "", 1.0)
		if err != nil {
			log.Println(err)
			SendBadRequest(w)
			return
		}
		heightPx, err := CSSDimensionsToPixels(parsedSvg.Height)
		if err != nil {
			log.Println(err)
			SendBadRequest(w)
			return
		}
		widthPx, err := CSSDimensionsToPixels(parsedSvg.Width)
		if err != nil {
			log.Println(err)
			SendBadRequest(w)
			return
		}
		img = image.Config{
			Width:  int(widthPx),
			Height: int(heightPx),
		}
		format = "svg+xml"
	}
	locationMap := &LocationMap{
		Width:    uint(img.Width),
		Height:   uint(img.Height),
		MimeType: format,
		Scale:    1.0,
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
	if !HasPermission(user, user.OrganizationID, PermissionAreas, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	org, err := GetOrganizationRepository().GetOne(user.OrganizationID)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	GetOrganizationRepository().CreateSampleData(org)
}

func weekdaysToString(days []int) string {
	if len(days) == 0 {
		return ""
	}
	sorted := make([]int, len(days))
	copy(sorted, days)
	slices.Sort(sorted)
	parts := make([]string, len(sorted))
	for i, d := range sorted {
		parts[i] = strconv.Itoa(d)
	}
	return strings.Join(parts, ",")
}

func (router *LocationRouter) copyFromRestModel(m *CreateLocationRequest) *Location {
	e := &Location{}
	e.Name = m.Name
	e.Description = m.Description
	e.MaxConcurrentBookings = m.MaxConcurrentBookings
	e.Timezone = m.Timezone
	e.Enabled = m.Enabled
	e.MapScale = m.MapScale
	e.MapType = m.MapType
	e.BookableDays = weekdaysToString(m.BookableDays)
	return e
}

func (router *LocationRouter) copyToRestModel(e *Location, allowedBookers []*LocationGroup) *GetLocationResponse {
	m := &GetLocationResponse{}
	m.ID = e.ID
	m.OrganizationID = e.OrganizationID
	m.Name = e.Name
	m.MapMimeType = e.MapMimeType
	m.MapWidth = e.MapWidth
	m.MapHeight = e.MapHeight
	m.MapScale = e.MapScale
	m.MapType = e.MapType
	m.Description = e.Description
	m.MaxConcurrentBookings = e.MaxConcurrentBookings
	m.Timezone = e.Timezone
	m.Enabled = e.Enabled
	m.BookableDays = service.WeekdaysFromString(e.BookableDays)

	if allowedBookers != nil {
		m.AllowedBookerGroupIDs = []string{}
		for _, allowedBooker := range allowedBookers {
			if allowedBooker.LocationID == e.ID {
				m.AllowedBookerGroupIDs = append(m.AllowedBookerGroupIDs, allowedBooker.GroupID)
			}
		}
	}

	return m
}
