package router

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/api"
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

type GetWeekdayResponse struct {
	BookingsByWeekday [7]int `json:"bookingsByWeekday"`
}

type GetStatsResponse struct {
	NumUsers             int    `json:"numUsers"`
	NumBookings          int    `json:"numBookings"`
	NumLocations         int    `json:"numLocations"`
	NumSpaces            int    `json:"numSpaces"`
	NumBookingsCurrent   int    `json:"numBookingsCurrent"`
	NumBookingsToday     int    `json:"numBookingsToday"`
	NumBookingsYesterday int    `json:"numBookingsYesterday"`
	NumBookingsThisWeek  int    `json:"numBookingsThisWeek"`
	BookingsByWeekday    [7]int `json:"bookingsByWeekday"`
	GetLoadResponse
}

func (router *StatsRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/", router.getStats).Methods("GET")
	s.HandleFunc("/load", router.getLoad).Methods("GET")
	s.HandleFunc("/weekday", router.getWeekday).Methods("GET")
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
	if !HasPermission(user, user.OrganizationID, PermissionAnalytics, PermissionLevelRead) {
		SendForbidden(w)
		return
	}
	hideStats, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingHideStats.Name)
	if hideStats {
		SendNotFound(w)
		return
	}

	locationId := r.URL.Query().Get("location")
	var location *Location = nil
	if locationId != "" {
		if uuid.Validate(locationId) != nil {
			SendBadRequest(w)
			return
		}
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
	load, _ := GetBookingRepository().GetLoadMulti(user.OrganizationID, []DateRange{
		{Enter: nextWeekEnter, Leave: nextWeekLeave},
		{Enter: thisWeekEnter, Leave: thisWeekLeave},
		{Enter: lastWeekEnter, Leave: lastWeekLeave},
		{Enter: lastMonthEnter, Leave: lastMonthLeave},
	}, location)
	if load != nil {
		m.SpaceLoadNextWeek, m.SpaceLoadThisWeek, m.SpaceLoadLastWeek, m.SpaceLoadLastMonth = load[0], load[1], load[2], load[3]
	}

	SendJSON(w, m)
}

func (router *StatsRouter) getStats(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionAnalytics, PermissionLevelRead) {
		SendForbidden(w)
		return
	}
	hideStats, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingHideStats.Name)
	if hideStats {
		SendNotFound(w)
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
	m.NumLocations, _ = GetLocationRepository().GetCount(user.OrganizationID)
	m.NumSpaces, _ = GetSpaceRepository().GetCount(user.OrganizationID)

	counts, _ := GetBookingRepository().GetCountsSummary(user.OrganizationID,
		DateRange{Enter: todayEnter, Leave: todayLeave},
		DateRange{Enter: yesterdayEnter, Leave: yesterdayLeave},
		DateRange{Enter: thisWeekEnter, Leave: thisWeekLeave})
	m.NumBookings = counts.Total
	m.NumBookingsCurrent = counts.Current
	m.NumBookingsToday = counts.Today
	m.NumBookingsYesterday = counts.Yesterday
	m.NumBookingsThisWeek = counts.ThisWeek

	load, _ := GetBookingRepository().GetLoadMulti(user.OrganizationID, []DateRange{
		{Enter: nextWeekEnter, Leave: nextWeekLeave},
		{Enter: thisWeekEnter, Leave: thisWeekLeave},
		{Enter: lastWeekEnter, Leave: lastWeekLeave},
		{Enter: lastMonthEnter, Leave: lastMonthLeave},
	}, nil)
	if load != nil {
		m.SpaceLoadNextWeek, m.SpaceLoadThisWeek, m.SpaceLoadLastWeek, m.SpaceLoadLastMonth = load[0], load[1], load[2], load[3]
	}

	m.BookingsByWeekday, _ = GetBookingRepository().GetCountByWeekday(user.OrganizationID, nil, nil, nil)
	SendJSON(w, m)
}

func (router *StatsRouter) getWeekday(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionAnalytics, PermissionLevelRead) {
		SendForbidden(w)
		return
	}
	hideStats, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingHideStats.Name)
	if hideStats {
		SendNotFound(w)
		return
	}

	locationId := r.URL.Query().Get("location")
	var location *Location = nil
	if locationId != "" {
		if uuid.Validate(locationId) != nil {
			SendBadRequest(w)
			return
		}
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

	var enter, leave *time.Time
	period := r.URL.Query().Get("period")
	if period != "" {
		thisWeekEnter, thisWeekLeave, lastWeekEnter, lastWeekLeave, nextWeekEnter, nextWeekLeave, lastMonthEnter, lastMonthLeave := getDateRanges()
		switch period {
		case "thisWeek":
			enter, leave = &thisWeekEnter, &thisWeekLeave
		case "lastWeek":
			enter, leave = &lastWeekEnter, &lastWeekLeave
		case "nextWeek":
			enter, leave = &nextWeekEnter, &nextWeekLeave
		case "lastMonth":
			enter, leave = &lastMonthEnter, &lastMonthLeave
		default:
			SendBadRequest(w)
			return
		}
	}

	m := &GetWeekdayResponse{}
	m.BookingsByWeekday, _ = GetBookingRepository().GetCountByWeekday(user.OrganizationID, location, enter, leave)
	SendJSON(w, m)
}
