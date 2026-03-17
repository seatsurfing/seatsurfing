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
	SpaceLoadNextWeek  int `json:"spaceLoadNextWeek"`
	SpaceLoadThisWeek  int `json:"spaceLoadThisWeek"`
	SpaceLoadLastWeek  int `json:"spaceLoadLastWeek"`
	SpaceLoadLastMonth int `json:"spaceLoadLastMonth"`
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

func getDateRanges() (thisWeekEnter, thisWeekLeave, lastWeekEnter, lastWeekLeave, nextWeekEnter, nextWeekLeave, lastMonthEnter, lastMonthLeave time.Time) {
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	// Current week: Monday to Sunday
	thisWeekEnter = time.Date(now.Year(), now.Month(), now.Day()-int(weekday-1), 0, 0, 0, 0, now.Location())
	thisWeekLeave = time.Date(now.Year(), now.Month(), now.Day()+int(7-weekday), 23, 59, 59, 0, now.Location())

	// Last week: Monday to Sunday
	lastWeekEnter = time.Date(now.Year(), now.Month(), now.Day()-int(weekday-1)-7, 0, 0, 0, 0, now.Location())
	lastWeekLeave = time.Date(now.Year(), now.Month(), now.Day()+int(7-weekday)-7, 23, 59, 59, 0, now.Location())

	// Next week: Monday to Sunday
	nextWeekEnter = time.Date(now.Year(), now.Month(), now.Day()-int(weekday-1)+7, 0, 0, 0, 0, now.Location())
	nextWeekLeave = time.Date(now.Year(), now.Month(), now.Day()+int(7-weekday)+7, 23, 59, 59, 0, now.Location())

	// Last month: 1st to last day
	lastMonthEnter = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	lastMonthLeave = time.Date(now.Year(), now.Month(), 0, 23, 59, 59, 0, now.Location())

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

	thisWeekEnter, thisWeekLeave, lastWeekEnter, lastWeekLeave, nextWeekEnter, nextWeekLeave, lastMonthEnter, lastMonthLeave := getDateRanges()

	m := &GetLoadResponse{}
	m.SpaceLoadNextWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, nextWeekEnter, nextWeekLeave, location)
	m.SpaceLoadThisWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, thisWeekEnter, thisWeekLeave, location)
	m.SpaceLoadLastWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, lastWeekEnter, lastWeekLeave, location)
	m.SpaceLoadLastMonth, _ = GetBookingRepository().GetLoad(user.OrganizationID, lastMonthEnter, lastMonthLeave, location)

	SendJSON(w, m)
}

func (router *StatsRouter) getStats(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}

	now := time.Now().UTC()
	todayEnter := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayLeave := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	yesterdayEnter := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	yesterdayLeave := time.Date(now.Year(), now.Month(), now.Day()-1, 23, 59, 59, 0, now.Location())

	thisWeekEnter, thisWeekLeave, lastWeekEnter, lastWeekLeave, nextWeekEnter, nextWeekLeave, lastMonthEnter, lastMonthLeave := getDateRanges()

	m := &GetStatsResponse{}
	m.NumUsers, _ = GetUserRepository().GetCount(user.OrganizationID)
	m.NumBookings, _ = GetBookingRepository().GetCount(user.OrganizationID)
	m.NumLocations, _ = GetLocationRepository().GetCount(user.OrganizationID)
	m.NumSpaces, _ = GetSpaceRepository().GetCount(user.OrganizationID)
	m.NumBookingsCurrent, _ = GetBookingRepository().GetCountCurrent(user.OrganizationID)
	m.NumBookingsToday, _ = GetBookingRepository().GetCountDateRange(user.OrganizationID, todayEnter, todayLeave)
	m.NumBookingsYesterday, _ = GetBookingRepository().GetCountDateRange(user.OrganizationID, yesterdayEnter, yesterdayLeave)
	m.NumBookingsThisWeek, _ = GetBookingRepository().GetCountDateRange(user.OrganizationID, thisWeekEnter, thisWeekLeave)
	m.SpaceLoadNextWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, nextWeekEnter, nextWeekLeave, nil)
	m.SpaceLoadThisWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, thisWeekEnter, thisWeekLeave, nil)
	m.SpaceLoadLastWeek, _ = GetBookingRepository().GetLoad(user.OrganizationID, lastWeekEnter, lastWeekLeave, nil)
	m.SpaceLoadLastMonth, _ = GetBookingRepository().GetLoad(user.OrganizationID, lastMonthEnter, lastMonthLeave, nil)
	SendJSON(w, m)
}
