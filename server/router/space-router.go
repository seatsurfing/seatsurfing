package router

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type SpaceRouter struct {
}

type CreateSpaceRequest struct {
	Name     string `json:"name" validate:"required"`
	X        uint   `json:"x"`
	Y        uint   `json:"y"`
	Width    uint   `json:"width"`
	Height   uint   `json:"height"`
	Rotation uint   `json:"rotation"`
}

type UpdateSpaceRequest struct {
	CreateSpaceRequest
	ID string `json:"id"`
}

type SpaceBulkUpdateRequest struct {
	Creates   []CreateSpaceRequest `json:"creates"`
	Updates   []UpdateSpaceRequest `json:"updates"`
	DeleteIDs []string             `json:"deleteIds"`
}

type BulkUpdateItemResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
}

type BulkUpdateResponse struct {
	Creates []BulkUpdateItemResponse `json:"creates"`
	Updates []BulkUpdateItemResponse `json:"updates"`
	Deletes []BulkUpdateItemResponse `json:"deletes"`
}

type GetSpaceResponse struct {
	ID         string              `json:"id"`
	Available  bool                `json:"available"`
	LocationID string              `json:"locationId"`
	Location   GetLocationResponse `json:"location"`
	CreateSpaceRequest
}

type GetSpaceAvailabilityBookingsResponse struct {
	BookingID string    `json:"id"`
	UserID    string    `json:"userId"`
	UserEmail string    `json:"userEmail"`
	Enter     time.Time `json:"enter"`
	Leave     time.Time `json:"leave"`
}

type GetSpaceAvailabilityResponse struct {
	GetSpaceResponse
	Bookings []*GetSpaceAvailabilityBookingsResponse `json:"bookings"`
}

type GetSpaceAvailabilityRequest struct {
	Enter time.Time `json:"enter" validate:"required"`
	Leave time.Time `json:"leave" validate:"required"`
}

func (router *SpaceRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/availability", router.getAvailability).Methods("POST")
	s.HandleFunc("/bulk", router.bulkUpdate).Methods("POST")
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *SpaceRouter) getOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetSpaceRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	location, err := GetLocationRepository().GetOne(e.LocationID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAccessOrg(user, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	res := router.copyToRestModel(e)
	SendJSON(w, res)
}

func (router *SpaceRouter) getAvailability(w http.ResponseWriter, r *http.Request) {
	var m GetSpaceAvailabilityRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	location, err := GetLocationRepository().GetOne(vars["locationId"])
	if err != nil {
		SendBadRequest(w)
		return
	}
	enterNew, err := GetLocationRepository().AttachTimezoneInformation(m.Enter, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	leaveNew, err := GetLocationRepository().AttachTimezoneInformation(m.Leave, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAccessOrg(user, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	var showNames bool = false
	if CanSpaceAdminOrg(user, location.OrganizationID) {
		showNames = true
	} else {
		showNames, _ = GetSettingsRepository().GetBool(location.OrganizationID, SettingShowNames.Name)
	}
	list, err := GetSpaceRepository().GetAllInTime(location.ID, enterNew, leaveNew)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetSpaceAvailabilityResponse{}
	for _, e := range list {
		m := &GetSpaceAvailabilityResponse{}
		m.ID = e.ID
		m.LocationID = e.LocationID
		m.Name = e.Name
		m.X = e.X
		m.Y = e.Y
		m.Width = e.Width
		m.Height = e.Height
		m.Rotation = e.Rotation
		m.Available = e.Available
		m.Bookings = []*GetSpaceAvailabilityBookingsResponse{}
		for _, booking := range e.Bookings {
			var showName bool = showNames
			enter, _ := GetLocationRepository().AttachTimezoneInformation(booking.Enter, location)
			leave, _ := GetLocationRepository().AttachTimezoneInformation(booking.Leave, location)
			outUserId := ""
			outUserEmail := ""
			if showName || user.Email == booking.UserEmail {
				outUserId = booking.UserID
				outUserEmail = booking.UserEmail
			}
			entry := &GetSpaceAvailabilityBookingsResponse{
				BookingID: booking.BookingID,
				UserID:    outUserId,
				UserEmail: outUserEmail,
				Enter:     enter,
				Leave:     leave,
			}
			m.Bookings = append(m.Bookings, entry)
		}
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *SpaceRouter) bulkUpdate(w http.ResponseWriter, r *http.Request) {
	var m SpaceBulkUpdateRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	location, err := GetLocationRepository().GetOne(vars["locationId"])
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, location.OrganizationID) {
		SendForbidden(w)
		return
	}

	res := BulkUpdateResponse{
		Creates: []BulkUpdateItemResponse{},
		Updates: []BulkUpdateItemResponse{},
		Deletes: []BulkUpdateItemResponse{},
	}

	// Process deletes
	if m.DeleteIDs != nil {
		for _, deleteID := range m.DeleteIDs {
			e, err := GetSpaceRepository().GetOne(deleteID)
			if err != nil {
				res.Deletes = append(res.Deletes, BulkUpdateItemResponse{ID: deleteID, Success: false})
			} else {
				if err := GetSpaceRepository().Delete(e); err != nil {
					res.Deletes = append(res.Deletes, BulkUpdateItemResponse{ID: deleteID, Success: false})
				} else {
					res.Deletes = append(res.Deletes, BulkUpdateItemResponse{ID: deleteID, Success: true})
				}
			}
		}
	}

	// Process creates
	if m.Creates != nil {
		for _, mSpace := range m.Creates {
			e := router.copyFromRestModel(&mSpace)
			e.LocationID = vars["locationId"]
			if err := GetSpaceRepository().Create(e); err != nil {
				log.Println(err)
				res.Creates = append(res.Creates, BulkUpdateItemResponse{ID: "", Success: false})
			} else {
				res.Creates = append(res.Creates, BulkUpdateItemResponse{ID: e.ID, Success: true})
			}
		}
	}

	// Process updates
	if m.Updates != nil {
		for _, mSpace := range m.Updates {
			e := router.copyFromRestModel(&mSpace.CreateSpaceRequest)
			e.ID = mSpace.ID
			e.LocationID = vars["locationId"]
			if err := GetSpaceRepository().Update(e); err != nil {
				log.Println(err)
				res.Updates = append(res.Updates, BulkUpdateItemResponse{ID: "", Success: false})
			} else {
				res.Updates = append(res.Updates, BulkUpdateItemResponse{ID: e.ID, Success: true})
			}
		}
	}
	SendJSON(w, res)
}

func (router *SpaceRouter) getAll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	location, err := GetLocationRepository().GetOne(vars["locationId"])
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAccessOrg(user, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	list, err := GetSpaceRepository().GetAll(location.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetSpaceResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *SpaceRouter) update(w http.ResponseWriter, r *http.Request) {
	var m CreateSpaceRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	e := router.copyFromRestModel(&m)
	e.ID = vars["id"]
	e.LocationID = vars["locationId"]
	location, err := GetLocationRepository().GetOne(e.LocationID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	if err := GetSpaceRepository().Update(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *SpaceRouter) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetSpaceRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	location, err := GetLocationRepository().GetOne(e.LocationID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	if err := GetSpaceRepository().Delete(e); err != nil {
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *SpaceRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateSpaceRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	e := router.copyFromRestModel(&m)
	e.LocationID = vars["locationId"]
	location, err := GetLocationRepository().GetOne(e.LocationID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	if err := GetSpaceRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}

func (router *SpaceRouter) copyFromRestModel(m *CreateSpaceRequest) *Space {
	e := &Space{}
	e.Name = m.Name
	e.X = m.X
	e.Y = m.Y
	e.Width = m.Width
	e.Height = m.Height
	e.Rotation = m.Rotation
	return e
}

func (router *SpaceRouter) copyToRestModel(e *Space) *GetSpaceResponse {
	m := &GetSpaceResponse{}
	m.ID = e.ID
	m.LocationID = e.LocationID
	m.Name = e.Name
	m.X = e.X
	m.Y = e.Y
	m.Width = e.Width
	m.Height = e.Height
	m.Rotation = e.Rotation
	return m
}
