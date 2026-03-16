package router

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type StatsRouter struct {
}

type GetLoadResponse struct {
	SpaceLoadToday     int `json:"spaceLoadToday"`
	SpaceLoadYesterday int `json:"spaceLoadYesterday"`
	SpaceLoadThisWeek  int `json:"spaceLoadThisWeek"`
	SpaceLoadLastWeek  int `json:"spaceLoadLastWeek"`
}

type GetStatsResponse struct {
	NumUsers             int `json:"numUsers"`
	NumBookings          int `json:"numBookings"`
	NumLocations         int `json:"numLocations"`
	NumSpaces            int `json:"numSpaces"`
	NumBookingsCurrent   int `json:"numBookingsCurrent"`
	NumBookingsToday     int `json:"numBookingsToday"`
	NumBookingsYesterday int `json:"numBookingsYesterday"`
	NumBookingsThisWeek  int `json:"numBookingsThisWeek"`
	GetLoadResponse
}

func (router *StatsRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/", router.getStats).Methods("GET")
	s.HandleFunc("/load", router.getLoad).Methods("GET")
}

func getDateRanges() (todayEnter, todayLeave, yesterdayEnter, yesterdayLeave, thisWeekEnter, thisWeekLeave, lastWeekEnter, lastWeekLeave time.Time) {
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	todayEnter = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayLeave = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	yesterdayEnter = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	yesterdayLeave = time.Date(now.Year(), now.Month(), now.Day()-1, 23, 59, 59, 0, now.Location())
	thisWeekEnter = time.Date(now.Year(), now.Month(), now.Day()-int(weekday-1), 0, 0, 0, 0, now.Location())
	thisWeekLeave = time.Date(now.Year(), now.Month(), now.Day()+int(7-weekday), 23, 59, 59, 0, now.Location())
	lastWeekEnter = time.Date(now.Year(), now.Month(), now.Day()-int(weekday-1)-7, 0, 0, 0, 0, now.Location())
	lastWeekLeave = time.Date(now.Year(), now.Month(), now.Day()+int(7-weekday)-7, 23, 59, 59, 0, now.Location())
	return
}

func (router *StatsRouter) getLoad(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}

	locationId := r.URL.Query().Get("location")
	var location *Location = nil
	if locationId != "" {
		var err error
		location, err = GetLocationRepository().GetOne(locationId)
		if err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
		if location == nil || location.OrganizationID != user.OrganizationID {
			SendBadRequest(w)
			return
		}
	}

	todayEnter, todayLeave, yesterdayEnter, yesterdayLeave, thisWeekEnter, thisWeekLeave, lastWeekEnter, lastWeekLeave := getDateRanges()

	m := &GetLoadResponse{}
	m.SpaceLoadToday, _ = GetBookingRepository().GetLoad(user.OrganizationID, todayEnter, todayLeave, location)
	m.SpaceLoadYesterday, _ = GetBookingRepository().GetLoad(user.OrganizationID, yesterdayEnter, yesterdayLeave, location)
	m.SpaceLoadThisWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, thisWeekEnter, thisWeekLeave, location)
	m.SpaceLoadLastWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, lastWeekEnter, lastWeekLeave, location)

	SendJSON(w, m)
}

func (router *StatsRouter) getStats(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}

	todayEnter, todayLeave, yesterdayEnter, yesterdayLeave, thisWeekEnter, thisWeekLeave, lastWeekEnter, lastWeekLeave := getDateRanges()

	m := &GetStatsResponse{}
	m.NumUsers, _ = GetUserRepository().GetCount(user.OrganizationID)
	m.NumBookings, _ = GetBookingRepository().GetCount(user.OrganizationID)
	m.NumLocations, _ = GetLocationRepository().GetCount(user.OrganizationID)
	m.NumSpaces, _ = GetSpaceRepository().GetCount(user.OrganizationID)
	m.NumBookingsCurrent, _ = GetBookingRepository().GetCountCurrent(user.OrganizationID)
	m.NumBookingsToday, _ = GetBookingRepository().GetCountDateRange(user.OrganizationID, todayEnter, todayLeave)
	m.NumBookingsYesterday, _ = GetBookingRepository().GetCountDateRange(user.OrganizationID, yesterdayEnter, yesterdayLeave)
	m.NumBookingsThisWeek, _ = GetBookingRepository().GetCountDateRange(user.OrganizationID, thisWeekEnter, thisWeekLeave)
	m.SpaceLoadToday, _ = GetBookingRepository().GetLoad(user.OrganizationID, todayEnter, todayLeave, nil)
	m.SpaceLoadYesterday, _ = GetBookingRepository().GetLoad(user.OrganizationID, yesterdayEnter, yesterdayLeave, nil)
	m.SpaceLoadThisWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, thisWeekEnter, thisWeekLeave, nil)
	m.SpaceLoadLastWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, lastWeekEnter, lastWeekLeave, nil)

	SendJSON(w, m)
}
