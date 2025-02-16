package router

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type SettingsRouter struct {
}

type SetSettingsRequest struct {
	Value string `json:"value"`
}

type GetSettingsResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var (
	ErrAlreadyExists          = errors.New("resource already exists")
	SysSettingOrgSignupDelete = "_sys_org_signup_delete"
	SysSettingVersion         = "_sys_version"
)

func (router *SettingsRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/timezones", router.getTimezones).Methods("GET")
	s.HandleFunc("/{name}", router.getSetting).Methods("GET")
	s.HandleFunc("/{name}", router.setSetting).Methods("PUT")
	s.HandleFunc("/", router.getAll).Methods("GET")
	s.HandleFunc("/", router.setAll).Methods("PUT")
}

func (router *SettingsRouter) getTimezones(w http.ResponseWriter, r *http.Request) {
	SendJSON(w, TimeZones)
}

func (router *SettingsRouter) getSetting(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	vars := mux.Vars(r)
	orgAdmin := CanAdminOrg(user, user.OrganizationID)
	if !((orgAdmin && router.isValidSettingNameReadAdmin(vars["name"])) || (router.isValidSettingNameReadPublic(vars["name"]))) {
		SendForbidden(w)
		return
	}
	if (vars["name"] == SysSettingOrgSignupDelete) && orgAdmin {
		SendJSON(w, router.getSysSettingOrgSignupDelete())
		return
	}
	if vars["name"] == SysSettingVersion {
		SendJSON(w, router.getSysSettingVersion())
		return
	}
	value, err := GetSettingsRepository().Get(user.OrganizationID, vars["name"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	SendJSON(w, value)
}

func (router *SettingsRouter) setSetting(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	var value SetSettingsRequest
	if UnmarshalValidateBody(r, &value) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	if !router.isValidSettingNameWrite(vars["name"]) {
		SendNotFound(w)
		return
	}
	if !router.isValidSettingType(vars["name"], value.Value) {
		SendBadRequest(w)
		return
	}
	if !router.isValidSettingValue(vars["name"], value.Value) {
		SendBadRequest(w)
		return
	}
	err := router.doSetOne(user.OrganizationID, vars["name"], value.Value)
	if err != nil {
		log.Println(err)
		if errors.Is(err, ErrAlreadyExists) {
			SendAleadyExists(w)
		} else {
			SendInternalServerError(w)
		}
		return
	}
	SendUpdated(w)
}

func (router *SettingsRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAccessOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	orgAdmin := CanAdminOrg(user, user.OrganizationID)
	list, err := GetSettingsRepository().GetAll(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetSettingsResponse{}
	for _, e := range list {
		if (orgAdmin && router.isValidSettingNameReadAdmin(e.Name)) || (router.isValidSettingNameReadPublic(e.Name)) {
			m := router.copyToRestModel(e)
			res = append(res, m)
		}
	}
	if orgAdmin {
		res = append(res, router.getSysSettingOrgSignupDelete())
	}
	res = append(res, router.getSysSettingVersion())
	SendJSON(w, res)
}

func (router *SettingsRouter) setAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	var list []GetSettingsResponse
	if err := UnmarshalBody(r, &list); err != nil {
		log.Println(err)
		SendBadRequest(w)
		return
	}
	for _, e := range list {
		if !router.isValidSettingNameWrite(e.Name) {
			SendNotFound(w)
			return
		}
		if !router.isValidSettingType(e.Name, e.Value) {
			SendBadRequest(w)
			return
		}
		if !router.isValidSettingValue(e.Name, e.Value) {
			SendBadRequest(w)
			return
		}
		err := router.doSetOne(user.OrganizationID, e.Name, e.Value)
		if err != nil {
			log.Println(err)
			if errors.Is(err, ErrAlreadyExists) {
				SendAleadyExists(w)
			} else {
				SendInternalServerError(w)
			}
			return
		}
	}
	SendUpdated(w)
}

func (router *SettingsRouter) doSetOne(organizationID, name, value string) error {
	err := GetSettingsRepository().Set(organizationID, name, value)
	return err
}

func (router *SettingsRouter) copyToRestModel(e *OrgSetting) *GetSettingsResponse {
	m := &GetSettingsResponse{}
	m.Name = e.Name
	m.Value = e.Value
	return m
}

func (router *SettingsRouter) isValidSettingNameReadPublic(name string) bool {
	if name == SettingMaxBookingsPerUser.Name ||
		name == SettingMaxConcurrentBookingsPerUser.Name ||
		name == SettingMaxDaysInAdvance.Name ||
		name == SettingMaxBookingDurationHours.Name ||
		name == SettingMaxHoursBeforeDelete.Name ||
		name == SettingMinBookingDurationHours.Name ||
		name == SettingShowNames.Name ||
		name == SettingEnableMaxHourBeforeDelete.Name ||
		name == SettingAllowBookingsNonExistingUsers.Name ||
		name == SettingDailyBasisBooking.Name ||
		name == SettingNoAdminRestrictions.Name ||
		name == SettingCustomLogoUrl.Name ||
		name == SettingMaxHoursPartiallyBooked.Name ||
		name == SettingMaxHoursPartiallyBookedEnabled.Name ||
		name == SettingDefaultTimezone.Name ||
		name == SettingDisableBuddies.Name ||
		name == SysSettingVersion {
		return true
	}
	return false
}

func (router *SettingsRouter) isValidSettingNameReadAdmin(name string) bool {
	if router.isValidSettingNameReadPublic(name) ||
		name == SettingAllowAnyUser.Name ||
		name == SettingMaxHoursBeforeDelete.Name ||
		name == SettingEnableMaxHourBeforeDelete.Name ||
		name == SettingMinBookingDurationHours.Name ||
		name == SettingActiveSubscription.Name ||
		name == SettingMaxHoursPartiallyBooked.Name ||
		name == SettingMaxHoursPartiallyBookedEnabled.Name ||
		name == SettingSubscriptionMaxUsers.Name ||
		name == SettingConfluenceServerSharedSecret.Name ||
		name == SettingConfluenceAnonymous.Name ||
		name == SysSettingOrgSignupDelete {
		return true
	}
	return false
}

func (router *SettingsRouter) isValidSettingNameWrite(name string) bool {
	if name == SettingAllowAnyUser.Name ||
		name == SettingConfluenceServerSharedSecret.Name ||
		name == SettingConfluenceAnonymous.Name ||
		name == SettingEnableMaxHourBeforeDelete.Name ||
		name == SettingMaxBookingsPerUser.Name ||
		name == SettingMaxConcurrentBookingsPerUser.Name ||
		name == SettingMaxDaysInAdvance.Name ||
		name == SettingMaxHoursBeforeDelete.Name ||
		name == SettingMinBookingDurationHours.Name ||
		name == SettingDailyBasisBooking.Name ||
		name == SettingNoAdminRestrictions.Name ||
		name == SettingCustomLogoUrl.Name ||
		name == SettingShowNames.Name ||
		name == SettingMaxHoursPartiallyBooked.Name ||
		name == SettingMaxHoursPartiallyBookedEnabled.Name ||
		name == SettingAllowBookingsNonExistingUsers.Name ||
		name == SettingMaxBookingDurationHours.Name ||
		name == SettingDisableBuddies.Name ||
		name == SettingDefaultTimezone.Name {
		return true
	}
	return false
}

func (router *SettingsRouter) getSettingType(name string) SettingType {
	if name == SettingAllowAnyUser.Name {
		return SettingAllowAnyUser.Type
	}
	if name == SettingConfluenceServerSharedSecret.Name {
		return SettingConfluenceServerSharedSecret.Type
	}
	if name == SettingConfluenceAnonymous.Name {
		return SettingConfluenceAnonymous.Type
	}
	if name == SettingMaxBookingsPerUser.Name {
		return SettingMaxBookingsPerUser.Type
	}
	if name == SettingMaxConcurrentBookingsPerUser.Name {
		return SettingMaxConcurrentBookingsPerUser.Type
	}
	if name == SettingMaxDaysInAdvance.Name {
		return SettingMaxDaysInAdvance.Type
	}
	if name == SettingMaxBookingDurationHours.Name {
		return SettingMaxBookingDurationHours.Type
	}
	if name == SettingDailyBasisBooking.Name {
		return SettingDailyBasisBooking.Type
	}
	if name == SettingNoAdminRestrictions.Name {
		return SettingNoAdminRestrictions.Type
	}
	if name == SettingShowNames.Name {
		return SettingShowNames.Type
	}
	if name == SettingAllowBookingsNonExistingUsers.Name {
		return SettingAllowBookingsNonExistingUsers.Type
	}
	if name == SettingDefaultTimezone.Name {
		return SettingDefaultTimezone.Type
	}
	if name == SettingMaxHoursBeforeDelete.Name {
		return SettingMaxHoursBeforeDelete.Type
	}
	if name == SettingEnableMaxHourBeforeDelete.Name {
		return SettingEnableMaxHourBeforeDelete.Type
	}
	if name == SettingDisableBuddies.Name {
		return SettingDisableBuddies.Type
	}
	if name == SettingCustomLogoUrl.Name {
		return SettingCustomLogoUrl.Type
	}
	if name == SettingMaxHoursPartiallyBooked.Name {
		return SettingMaxHoursPartiallyBooked.Type
	}
	if name == SettingMaxHoursPartiallyBookedEnabled.Name {
		return SettingMaxHoursPartiallyBookedEnabled.Type
	}
	if name == SettingMinBookingDurationHours.Name {
		return SettingMinBookingDurationHours.Type
	}
	return 0
}

func (router *SettingsRouter) isValidSettingType(name string, value string) bool {
	settingType := router.getSettingType(name)
	if settingType == 0 {
		return false
	}
	if settingType == SettingTypeString {
		return true
	}
	if settingType == SettingTypeBool && (value == "1" || value == "0") {
		return true
	}
	if settingType == SettingTypeInt {
		if _, err := strconv.Atoi(value); err == nil {
			return true
		}
	}
	return false
}

func (router *SettingsRouter) isValidSettingValue(name string, value string) bool {
	if name == SettingDefaultTimezone.Name && !IsValidTimeZone(value) {
		return false
	}
	return true
}

func (router *SettingsRouter) getSysSettingOrgSignupDelete() *GetSettingsResponse {
	boolVal := "0"
	if GetConfig().OrgSignupDelete {
		boolVal = "1"
	}
	return &GetSettingsResponse{
		Name:  SysSettingOrgSignupDelete,
		Value: boolVal,
	}
}

func (router *SettingsRouter) getSysSettingVersion() *GetSettingsResponse {
	return &GetSettingsResponse{
		Name:  SysSettingVersion,
		Value: GetProductVersion(),
	}
}
