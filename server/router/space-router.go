package router

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type SpaceRouter struct {
}

type SpaceAttributeValueRequest struct {
	AttributeID string `json:"attributeId"`
	Value       string `json:"value"`
}

type CreateSpaceRequest struct {
	Name                  string                       `json:"name" validate:"required"`
	X                     uint                         `json:"x"`
	Y                     uint                         `json:"y"`
	Width                 uint                         `json:"width"`
	Height                uint                         `json:"height"`
	Rotation              uint                         `json:"rotation"`
	RequireSubject        bool                         `json:"requireSubject"`
	Attributes            []SpaceAttributeValueRequest `json:"attributes"`
	ApproverGroupIDs      []string                     `json:"approverGroupIds"`
	AllowedBookerGroupIDs []string                     `json:"allowedBookerGroupIds"`
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
	ID         string               `json:"id"`
	Available  bool                 `json:"available"`
	LocationID string               `json:"locationId"`
	Location   *GetLocationResponse `json:"location,omitempty"`
	CreateSpaceRequest
}

type GetSpaceAvailabilityBookingsResponse struct {
	BookingID string    `json:"id"`
	UserID    string    `json:"userId"`
	UserEmail string    `json:"userEmail"`
	Enter     time.Time `json:"enter"`
	Leave     time.Time `json:"leave"`
	Subject   string    `json:"subject"`
}

type GetSpaceAvailabilityResponse struct {
	GetSpaceResponse
	Bookings           []*GetSpaceAvailabilityBookingsResponse `json:"bookings"`
	IsAllowed          bool                                    `json:"allowed"`
	IsApprovalRequired bool                                    `json:"approvalRequired"`
}

type GetSpaceAvailabilityRequest struct {
	Enter      time.Time         `json:"enter"`
	Leave      time.Time         `json:"leave"`
	SpaceID    string            `json:"spaceId"`
	Attributes []SearchAttribute `json:"attributes"`
}

func (router *SpaceRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/availability", router.getAvailability).Methods("GET")
	s.HandleFunc("/bulk", router.bulkUpdate).Methods("POST")
	s.HandleFunc("/{id}/approver/remove", router.removeApprovers).Methods("POST")
	s.HandleFunc("/{id}/approver", router.getApprovers).Methods("GET")
	s.HandleFunc("/{id}/approver", router.addApprovers).Methods("PUT")
	s.HandleFunc("/{id}/allowedbooker/remove", router.removeAllowedBookers).Methods("POST")
	s.HandleFunc("/{id}/allowedbooker", router.getAllowedBookers).Methods("GET")
	s.HandleFunc("/{id}/allowedbooker", router.addAllowedBookers).Methods("PUT")
	s.HandleFunc("/{id}/availability", router.getSingleSpaceAvailability).Methods("GET")
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
	attributes, err := GetSpaceAttributeValueRepository().GetAllForEntity(e.ID, SpaceAttributeValueEntityTypeSpace)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	approvers, err := GetSpaceRepository().GetAllApproversForSpaceList([]string{e.ID})
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	allowedBookers, err := GetSpaceRepository().GetAllAllowedBookersForSpaceList([]string{e.ID})
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := router.copyToRestModel(e, attributes, approvers, allowedBookers)
	SendJSON(w, res)
}

func (router *SpaceRouter) getSingleSpaceAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	router._getAvailability(vars["id"], w, r)
}

func (router *SpaceRouter) getAvailability(w http.ResponseWriter, r *http.Request) {
	router._getAvailability("", w, r)
}

func (router *SpaceRouter) _getAvailability(spaceID string, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationId := vars["locationId"]
	location, err := GetLocationRepository().GetOne(locationId)
	if err != nil {
		SendBadRequest(w)
		return
	}
	var enter, leave time.Time
	if !r.URL.Query().Has("enter") && !r.URL.Query().Has("leave") {
		tz := location.Timezone
		if tz == "" {
			tz, _ = GetSettingsRepository().Get(location.OrganizationID, SettingDefaultTimezone.Name)
		}
		tzLocation, err := time.LoadLocation(tz)
		if err != nil || tzLocation == nil {
			log.Println("Error loading timezone:", tz, "Error:", err)
			SendInternalServerError(w)
			return
		}
		enter = time.Now().In(tzLocation).Add(time.Minute * -1)
		leave = time.Now().In(tzLocation).Add(time.Minute * +1)
	} else if r.URL.Query().Has("enter") && r.URL.Query().Has("leave") {
		var err error
		if enter, err = time.Parse(time.RFC3339Nano, r.URL.Query().Get("enter")); err != nil {
			SendBadRequest(w)
			return
		}
		if leave, err = time.Parse(time.RFC3339Nano, r.URL.Query().Get("leave")); err != nil {
			SendBadRequest(w)
			return
		}
	} else {
		SendBadRequest(w)
		return
	}
	enterNew, err := GetLocationRepository().AttachTimezoneInformation(enter, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	leaveNew, err := GetLocationRepository().AttachTimezoneInformation(leave, location)
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
	spaceIds := []string{}
	for _, e := range list {
		spaceIds = append(spaceIds, e.Space.ID)
	}
	attributeValues, err := GetSpaceAttributeValueRepository().GetAllForEntityList(spaceIds, SpaceAttributeValueEntityTypeSpace)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	userGroups, err := GetGroupRepository().GetAllWhereUserIsMember(user.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	allowedBookers, err := GetSpaceRepository().GetAllAllowedBookersForSpaceList(spaceIds)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	approvers, err := GetSpaceRepository().GetAllApproversForSpaceList(spaceIds)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	attributes := []SearchAttribute{}
	if r.URL.Query().Has("attributes") {
		json.Unmarshal([]byte(r.URL.Query().Get("attributes")), &attributes)
	}
	res := []*GetSpaceAvailabilityResponse{}
	for _, e := range list {
		if spaceID != "" && e.ID != spaceID {
			continue
		}
		if MatchesSearchAttributes(e.ID, &attributes, attributeValues) {
			m := &GetSpaceAvailabilityResponse{}
			m.ID = e.ID
			m.LocationID = e.LocationID
			m.Name = e.Name
			m.X = e.X
			m.Y = e.Y
			m.Width = e.Width
			m.Height = e.Height
			m.Rotation = e.Rotation
			m.RequireSubject = e.RequireSubject
			m.Available = e.Available
			m.IsAllowed = router.IsUserAllowedToBook(&e.Space, allowedBookers, userGroups)
			m.IsApprovalRequired = router.IsApprovalRequired(&e.Space, approvers)
			router.appendAttributesToRestModel(&m.GetSpaceResponse, attributeValues)
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
					Subject:   booking.Subject,
				}
				m.Bookings = append(m.Bookings, entry)
			}
			res = append(res, m)
		}
	}
	SendJSON(w, res)
}

func (router *SpaceRouter) IsApprovalRequired(e *Space, approvers []*SpaceGroup) bool {
	for _, approver := range approvers {
		if approver.SpaceID == e.ID {
			return true
		}
	}
	return false
}

func (router *SpaceRouter) IsUserAllowedToBook(e *Space, allowedBookers []*SpaceGroup, userGroups []*Group) bool {
	restricted := false
	for _, allowedBooker := range allowedBookers {
		if allowedBooker.SpaceID == e.ID {
			restricted = true
			for _, userGroup := range userGroups {
				if allowedBooker.GroupID == userGroup.ID {
					return true
				}
			}
		}
	}
	return !restricted
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
	availableAttributes, err := GetSpaceAttributeRepository().GetAll(location.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
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
				if err := router.applySpaceAttributes(availableAttributes, e, &mSpace); err != nil {
					log.Println("Could not apply space attributes:", err)
				}
				if err := router.applyApprovers(e, &mSpace); err != nil {
					log.Println("Could not apply approvers:", err)
				}
				if err := router.applyAllowBookers(e, &mSpace); err != nil {
					log.Println("Could not apply allow bookers:", err)
				}
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
				if err := router.applySpaceAttributes(availableAttributes, e, &mSpace.CreateSpaceRequest); err != nil {
					log.Println("Could not apply space attributes:", err)
				}
				if err := router.applyApprovers(e, &mSpace.CreateSpaceRequest); err != nil {
					log.Println("Could not apply approvers:", err)
				}
				if err := router.applyAllowBookers(e, &mSpace.CreateSpaceRequest); err != nil {
					log.Println("Could not apply allow bookers:", err)
				}
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
	spaceIds := []string{}
	for _, e := range list {
		spaceIds = append(spaceIds, e.ID)
	}
	attributes, err := GetSpaceAttributeValueRepository().GetAllForEntityList(spaceIds, SpaceAttributeValueEntityTypeSpace)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	approvers, err := GetSpaceRepository().GetAllApproversForSpaceList(spaceIds)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	allowedBookers, err := GetSpaceRepository().GetAllAllowedBookersForSpaceList(spaceIds)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetSpaceResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e, attributes, approvers, allowedBookers)
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
	availableAttributes, err := GetSpaceAttributeRepository().GetAll(location.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	router.applySpaceAttributes(availableAttributes, e, &m)
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
	availableAttributes, err := GetSpaceAttributeRepository().GetAll(location.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	router.applySpaceAttributes(availableAttributes, e, &m)
	SendCreated(w, e.ID)
}

func (router *SpaceRouter) applySpaceAttributes(availableAttributes []*SpaceAttribute, space *Space, m *CreateSpaceRequest) error {
	existingSpaceAttributes, err := GetSpaceAttributeValueRepository().GetAllForEntity(space.ID, SpaceAttributeValueEntityTypeSpace)
	if err != nil {
		return err
	}
	// Check deletes
	for _, attribute := range existingSpaceAttributes {
		found := false
		for _, mAttribute := range m.Attributes {
			if attribute.AttributeID == mAttribute.AttributeID {
				found = true
				break
			}
		}
		if !found {
			if err := GetSpaceAttributeValueRepository().Delete(attribute.AttributeID, space.ID, SpaceAttributeValueEntityTypeSpace); err != nil {
				return err
			}
		}
	}
	// Check creates / updates
	for _, mAttribute := range m.Attributes {
		// Check if attribute is valid
		found := false
		for _, availableAttribute := range availableAttributes {
			if availableAttribute.ID == mAttribute.AttributeID {
				found = true
				break
			}
		}
		if found {
			if err := GetSpaceAttributeValueRepository().Set(mAttribute.AttributeID, space.ID, SpaceAttributeValueEntityTypeSpace, mAttribute.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (router *SpaceRouter) applyApprovers(space *Space, m *CreateSpaceRequest) error {
	existingApprovers, err := GetSpaceRepository().GetApproverGroupIDs(space.ID)
	if err != nil {
		return err
	}
	// Check deletes
	removes := []string{}
	for _, approver := range existingApprovers {
		found := false
		for _, mApprover := range m.ApproverGroupIDs {
			if approver == mApprover {
				found = true
				break
			}
		}
		if !found {
			removes = append(removes, approver)
		}
	}
	// Check creates
	adds := []string{}
	for _, mApprover := range m.ApproverGroupIDs {
		found := false
		for _, approver := range existingApprovers {
			if approver == mApprover {
				found = true
				break
			}
		}
		if !found {
			adds = append(adds, mApprover)
		}
	}
	if err := GetSpaceRepository().AddApprovers(space, adds); err != nil {
		return err
	}
	if err := GetSpaceRepository().RemoveApprovers(space, removes); err != nil {
		return err
	}
	return nil
}

func (router *SpaceRouter) applyAllowBookers(space *Space, m *CreateSpaceRequest) error {
	existingAllowBookers, err := GetSpaceRepository().GetAllowedBookersGroupIDs(space)
	if err != nil {
		return err
	}
	// Check deletes
	removes := []string{}
	for _, allowBooker := range existingAllowBookers {
		found := false
		for _, mAllowBooker := range m.AllowedBookerGroupIDs {
			if allowBooker == mAllowBooker {
				found = true
				break
			}
		}
		if !found {
			removes = append(removes, allowBooker)
		}
	}
	// Check creates
	adds := []string{}
	for _, mAllowBooker := range m.AllowedBookerGroupIDs {
		found := false
		for _, allowBooker := range existingAllowBookers {
			if allowBooker == mAllowBooker {
				found = true
				break
			}
		}
		if !found {
			adds = append(adds, mAllowBooker)
		}
	}
	if err := GetSpaceRepository().AddAllowedBookers(space, adds); err != nil {
		return err
	}
	if err := GetSpaceRepository().RemoveAllowedBookers(space, removes); err != nil {
		return err
	}
	return nil
}

func (router *SpaceRouter) addApprovers(w http.ResponseWriter, r *http.Request) {
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
	e, err := GetSpaceRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	var approvers []string
	if UnmarshalBody(r, &approvers) != nil {
		SendBadRequest(w)
		return
	}
	ok, err := GetGroupRepository().GroupsExistAndBelongToOrg(location.OrganizationID, approvers)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if !ok {
		SendBadRequest(w)
		return
	}
	if err := GetSpaceRepository().AddApprovers(e, approvers); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *SpaceRouter) removeApprovers(w http.ResponseWriter, r *http.Request) {
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
	e, err := GetSpaceRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	var approvers []string
	if UnmarshalBody(r, &approvers) != nil {
		SendBadRequest(w)
		return
	}
	if err := GetSpaceRepository().RemoveApprovers(e, approvers); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *SpaceRouter) getApprovers(w http.ResponseWriter, r *http.Request) {
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
	e, err := GetSpaceRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	approvers, err := GetSpaceRepository().GetApproverGroupIDs(e.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	groups, err := GetGroupRepository().GetAllByIDs(approvers)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	gr := &GroupRouter{}
	res := []*GetGroupResponse{}
	for _, e := range groups {
		m := gr.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *SpaceRouter) addAllowedBookers(w http.ResponseWriter, r *http.Request) {
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
	e, err := GetSpaceRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	var approvers []string
	if UnmarshalBody(r, &approvers) != nil {
		SendBadRequest(w)
		return
	}
	ok, err := GetGroupRepository().GroupsExistAndBelongToOrg(location.OrganizationID, approvers)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if !ok {
		SendBadRequest(w)
		return
	}
	if err := GetSpaceRepository().AddAllowedBookers(e, approvers); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *SpaceRouter) removeAllowedBookers(w http.ResponseWriter, r *http.Request) {
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
	e, err := GetSpaceRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	var approvers []string
	if UnmarshalBody(r, &approvers) != nil {
		SendBadRequest(w)
		return
	}
	if err := GetSpaceRepository().RemoveAllowedBookers(e, approvers); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *SpaceRouter) getAllowedBookers(w http.ResponseWriter, r *http.Request) {
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
	e, err := GetSpaceRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	approvers, err := GetSpaceRepository().GetAllowedBookersGroupIDs(e)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	groups, err := GetGroupRepository().GetAllByIDs(approvers)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	gr := &GroupRouter{}
	res := []*GetGroupResponse{}
	for _, e := range groups {
		m := gr.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *SpaceRouter) copyFromRestModel(m *CreateSpaceRequest) *Space {
	e := &Space{}
	e.Name = m.Name
	e.X = m.X
	e.Y = m.Y
	e.Width = m.Width
	e.Height = m.Height
	e.Rotation = m.Rotation
	e.RequireSubject = m.RequireSubject
	return e
}

func (router *SpaceRouter) copyToRestModel(e *Space, attributes []*SpaceAttributeValue, approvers, allowedBookers []*SpaceGroup) *GetSpaceResponse {
	m := &GetSpaceResponse{}
	m.ID = e.ID
	m.LocationID = e.LocationID
	m.Name = e.Name
	m.X = e.X
	m.Y = e.Y
	m.Width = e.Width
	m.Height = e.Height
	m.Rotation = e.Rotation
	m.RequireSubject = e.RequireSubject
	if attributes != nil {
		m.Attributes = []SpaceAttributeValueRequest{}
		for _, attribute := range attributes {
			if attribute.EntityType == SpaceAttributeValueEntityTypeSpace {
				if attribute.EntityID == e.ID {
					m.Attributes = append(m.Attributes, SpaceAttributeValueRequest{AttributeID: attribute.AttributeID, Value: attribute.Value})
				}
			}
		}
	}
	if approvers != nil {
		m.ApproverGroupIDs = []string{}
		for _, approver := range approvers {
			if approver.SpaceID == e.ID {
				m.ApproverGroupIDs = append(m.ApproverGroupIDs, approver.GroupID)
			}
		}
	}
	if allowedBookers != nil {
		m.AllowedBookerGroupIDs = []string{}
		for _, allowedBooker := range allowedBookers {
			if allowedBooker.SpaceID == e.ID {
				m.AllowedBookerGroupIDs = append(m.AllowedBookerGroupIDs, allowedBooker.GroupID)
			}
		}
	}
	return m
}

func (router *SpaceRouter) appendAttributesToRestModel(m *GetSpaceResponse, attributes []*SpaceAttributeValue) {
	if attributes != nil {
		m.Attributes = []SpaceAttributeValueRequest{}
		for _, attribute := range attributes {
			if attribute.EntityType == SpaceAttributeValueEntityTypeSpace {
				if attribute.EntityID == m.ID {
					m.Attributes = append(m.Attributes, SpaceAttributeValueRequest{AttributeID: attribute.AttributeID, Value: attribute.Value})
				}
			}
		}
	}
}
