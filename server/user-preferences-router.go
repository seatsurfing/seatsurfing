package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type UserPreferencesRouter struct {
}

type ListCaldavCalendarsRequest struct {
	URL      string `json:"url" validate:"required,url"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type ListCaldavCalendarsResponse struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func (router *UserPreferencesRouter) setupRoutes(s *mux.Router) {
	s.HandleFunc("/caldav/listCalendars", router.caldavListCalendars).Methods("POST")
	s.HandleFunc("/{name}", router.getPreference).Methods("GET")
	s.HandleFunc("/{name}", router.setPreference).Methods("PUT")
	s.HandleFunc("/", router.getAll).Methods("GET")
	s.HandleFunc("/", router.setAll).Methods("PUT")
}

func (router *UserPreferencesRouter) caldavListCalendars(w http.ResponseWriter, r *http.Request) {
	if !canCrypt() {
		log.Println("Error: CalDAV integration requires a valid crypt key (CRYPT_KEY).")
		SendInternalServerError(w)
		return
	}
	var m ListCaldavCalendarsRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	caldavClient := &CalDAVClient{}
	if err := caldavClient.Connect(m.URL, m.Username, m.Password); err != nil {
		SendNotFound(w)
		return
	}
	calendars, err := caldavClient.ListCalendars()
	if err != nil {
		SendNotFound(w)
		return
	}
	res := make([]*ListCaldavCalendarsResponse, 0)
	for _, calendar := range calendars {
		res = append(res, &ListCaldavCalendarsResponse{Path: calendar.Path, Name: calendar.Name})
	}
	SendJSON(w, res)
}

func (router *UserPreferencesRouter) getPreference(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	vars := mux.Vars(r)
	if !router.isValidPreferenceName(vars["name"]) {
		SendNotFound(w)
		return
	}
	value, err := GetUserPreferencesRepository().Get(user.ID, vars["name"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	SendJSON(w, value)
}

func (router *UserPreferencesRouter) setPreference(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	var value SetSettingsRequest
	if UnmarshalValidateBody(r, &value) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	if !router.isValidPreferenceName(vars["name"]) {
		SendNotFound(w)
		return
	}
	if !router.isValidPreferenceType(vars["name"], value.Value) {
		SendBadRequest(w)
		return
	}
	if !router.isValidPreferenceValue(vars["name"], value.Value) {
		SendBadRequest(w)
		return
	}
	err := router.doSetOne(user.ID, vars["name"], value.Value)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *UserPreferencesRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	list, err := GetUserPreferencesRepository().GetAll(user.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetSettingsResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *UserPreferencesRouter) setAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	var list []GetSettingsResponse
	if err := UnmarshalBody(r, &list); err != nil {
		log.Println(err)
		SendBadRequest(w)
		return
	}
	for _, e := range list {
		if !router.isValidPreferenceName(e.Name) {
			SendNotFound(w)
			return
		}
		if !router.isValidPreferenceType(e.Name, e.Value) {
			SendBadRequest(w)
			return
		}
		if !router.isValidPreferenceValue(e.Name, e.Value) {
			SendBadRequest(w)
			return
		}
		err := router.doSetOne(user.ID, e.Name, e.Value)
		if err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
	}
	SendUpdated(w)
}

func (router *UserPreferencesRouter) doSetOne(userID, name, value string) error {
	if router.getPreferenceType(name) == SettingTypeEncryptedString {
		value = encryptString(value)
	}
	err := GetUserPreferencesRepository().Set(userID, name, value)
	return err
}

func (router *UserPreferencesRouter) isValidPreferenceName(name string) bool {
	if name == PreferenceEnterTime.Name ||
		name == PreferenceWorkdayStart.Name ||
		name == PreferenceWorkdayEnd.Name ||
		name == PreferenceWorkdays.Name ||
		name == PreferenceBookedColor.Name ||
		name == PreferenceBuddyBookedColor.Name ||
		name == PreferenceSelfBookedColor.Name ||
		name == PreferencePartiallyBookedColor.Name ||
		name == PreferenceNotBookedColor.Name ||
		name == PreferenceLocation.Name ||
		name == PreferenceCalDAVURL.Name ||
		name == PreferenceCalDAVUser.Name ||
		name == PreferenceCalDAVPass.Name ||
		name == PreferenceCalDAVPath.Name {
		return true
	}
	return false
}

func (router *UserPreferencesRouter) getPreferenceType(name string) SettingType {
	if name == PreferenceEnterTime.Name {
		return PreferenceEnterTime.Type
	}
	if name == PreferenceWorkdayStart.Name {
		return PreferenceWorkdayStart.Type
	}
	if name == PreferenceWorkdayEnd.Name {
		return PreferenceWorkdayEnd.Type
	}
	if name == PreferenceBookedColor.Name {
		return PreferenceBookedColor.Type
	}
	if name == PreferenceBuddyBookedColor.Name {
		return PreferenceBuddyBookedColor.Type
	}
	if name == PreferenceSelfBookedColor.Name {
		return PreferenceSelfBookedColor.Type
	}
	if name == PreferencePartiallyBookedColor.Name {
		return PreferencePartiallyBookedColor.Type
	}
	if name == PreferenceNotBookedColor.Name {
		return PreferenceNotBookedColor.Type
	}
	if name == PreferenceWorkdays.Name {
		return PreferenceWorkdays.Type
	}
	if name == PreferenceLocation.Name {
		return PreferenceLocation.Type
	}
	if name == PreferenceCalDAVURL.Name {
		return PreferenceCalDAVURL.Type
	}
	if name == PreferenceCalDAVUser.Name {
		return PreferenceCalDAVUser.Type
	}
	if name == PreferenceCalDAVPass.Name {
		return PreferenceCalDAVPass.Type
	}
	if name == PreferenceCalDAVPath.Name {
		return PreferenceCalDAVPath.Type
	}
	return 0
}

func (router *UserPreferencesRouter) isValidPreferenceType(name string, value string) bool {
	settingType := router.getPreferenceType(name)
	if settingType == 0 {
		return false
	}
	if settingType == SettingTypeString || settingType == SettingTypeEncryptedString {
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
	if settingType == SettingTypeIntArray {
		tokens := strings.Split(value, ",")
		ok := true
		for _, token := range tokens {
			if _, err := strconv.Atoi(token); err != nil {
				ok = false
			}
		}
		return ok
	}
	return false
}

func (router *UserPreferencesRouter) isValidPreferenceValue(name string, value string) bool {
	if name == PreferenceEnterTime.Name {
		i, _ := strconv.Atoi(value)
		if !(i == PreferenceEnterTimeNow || i == PreferenceEnterTimeNextDay || i == PreferenceEnterTimeNextWorkday) {
			return false
		}
	}
	if name == PreferenceWorkdayStart.Name {
		i, _ := strconv.Atoi(value)
		if i < 0 || i > 24 {
			return false
		}
	}
	if name == PreferenceWorkdayEnd.Name {
		i, _ := strconv.Atoi(value)
		if i < 0 || i > 24 {
			return false
		}
	}
	if name == PreferenceWorkdays.Name {
		tokens := strings.Split(value, ",")
		ok := true
		for _, token := range tokens {
			if workday, err := strconv.Atoi(token); err != nil || workday < 0 || workday > 6 {
				ok = false
			}
		}
		return ok
	}
	return true
}

func (router *UserPreferencesRouter) copyToRestModel(e *UserPreference) *GetSettingsResponse {
	m := &GetSettingsResponse{}
	m.Name = e.Name
	if router.getPreferenceType(e.Name) == SettingTypeEncryptedString {
		m.Value = decryptString(e.Value)
	} else {
		m.Value = e.Value
	}
	return m
}
