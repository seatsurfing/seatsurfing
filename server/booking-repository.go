package main

import (
	"database/sql"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
)

type BookingRepository struct {
}

type Booking struct {
	ID       string
	UserID   string
	SpaceID  string
	Enter    time.Time
	Leave    time.Time
	CalDavID string
}

type BookingDetails struct {
	Space     SpaceDetails
	UserEmail string
	Booking
}

type BookingPresenceItem struct {
	User     *User
	Presence map[string]int
}

var bookingRepository *BookingRepository
var bookingRepositoryOnce sync.Once

func GetBookingRepository() *BookingRepository {
	bookingRepositoryOnce.Do(func() {
		bookingRepository = &BookingRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS bookings (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"user_id uuid NOT NULL, " +
			"space_id uuid NOT NULL, " +
			"enter_time TIMESTAMP NOT NULL, " +
			"leave_time TIMESTAMP NOT NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id)")
		if err != nil {
			panic(err)
		}
	})
	return bookingRepository
}

func (r *BookingRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 18 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE bookings " +
			"ADD COLUMN caldav_id VARCHAR NOT NULL DEFAULT ''"); err != nil {
			panic(err)
		}
	}
}

func (r *BookingRepository) Create(e *Booking) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO bookings "+
		"(user_id, space_id, enter_time, leave_time, caldav_id) "+
		"VALUES ($1, $2, $3, $4, $5) "+
		"RETURNING id",
		e.UserID, e.SpaceID, e.Enter, e.Leave, e.CalDavID).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *BookingRepository) GetOne(id string) (*BookingDetails, error) {
	e := &BookingDetails{}
	err := GetDatabase().DB().QueryRow("SELECT bookings.id, bookings.user_id, bookings.space_id, bookings.enter_time, bookings.leave_time, bookings.caldav_id, "+
		"spaces.id, spaces.location_id, spaces.name, "+
		"locations.id, locations.organization_id, locations.name, locations.description, locations.tz, "+
		"users.email "+
		"FROM bookings "+
		"INNER JOIN spaces ON bookings.space_id = spaces.id "+
		"INNER JOIN locations ON spaces.location_id = locations.id "+
		"INNER JOIN users ON bookings.user_id = users.id "+
		"WHERE bookings.id = $1",
		id).Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.CalDavID, &e.Space.ID, &e.Space.LocationID, &e.Space.Name, &e.Space.Location.ID, &e.Space.Location.OrganizationID, &e.Space.Location.Name, &e.Space.Location.Description, &e.Space.Location.Timezone, &e.UserEmail)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// Get first upcoming booking by user
func (r *BookingRepository) GetFirstUpcomingBookingByUserID(userID string) (*BookingDetails, error) {
	e := &BookingDetails{}
	err := GetDatabase().DB().QueryRow("SELECT bookings.id, bookings.user_id, bookings.space_id, bookings.enter_time, bookings.leave_time, bookings.caldav_id, "+
		"spaces.id, spaces.location_id, spaces.name, "+
		"locations.id, locations.organization_id, locations.name, locations.description, locations.tz, "+
		"users.email "+
		"FROM bookings "+
		"INNER JOIN spaces ON bookings.space_id = spaces.id "+
		"INNER JOIN locations ON spaces.location_id = locations.id "+
		"INNER JOIN users ON bookings.user_id = users.id "+
		"WHERE bookings.user_id = $1 AND bookings.enter_time > $2 "+
		"ORDER BY bookings.enter_time ASC LIMIT 1",
		userID, time.Now()).Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.CalDavID, &e.Space.ID, &e.Space.LocationID, &e.Space.Name, &e.Space.Location.ID, &e.Space.Location.OrganizationID, &e.Space.Location.Name, &e.Space.Location.Description, &e.Space.Location.Timezone, &e.UserEmail)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *BookingRepository) GetAllByOrg(organizationID string, startTime, endTime time.Time) ([]*BookingDetails, error) {
	var result []*BookingDetails
	rows, err := GetDatabase().DB().Query("SELECT bookings.id, bookings.user_id, bookings.space_id, bookings.enter_time, bookings.leave_time, bookings.caldav_id, "+
		"spaces.id, spaces.location_id, spaces.name, "+
		"locations.id, locations.organization_id, locations.name, locations.description, locations.tz, "+
		"users.email "+
		"FROM bookings "+
		"INNER JOIN spaces ON bookings.space_id = spaces.id "+
		"INNER JOIN locations ON spaces.location_id = locations.id "+
		"INNER JOIN users ON bookings.user_id = users.id "+
		"WHERE locations.organization_id = $1 AND leave_time >= $2 AND enter_time <= $3 "+
		"ORDER BY enter_time", organizationID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &BookingDetails{}
		err = rows.Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.CalDavID, &e.Space.ID, &e.Space.LocationID, &e.Space.Name, &e.Space.Location.ID, &e.Space.Location.OrganizationID, &e.Space.Location.Name, &e.Space.Location.Description, &e.Space.Location.Timezone, &e.UserEmail)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *BookingRepository) GetAllByUser(userID string, startTime time.Time) ([]*BookingDetails, error) {
	var result []*BookingDetails
	rows, err := GetDatabase().DB().Query("SELECT bookings.id, bookings.user_id, bookings.space_id, bookings.enter_time, bookings.leave_time, bookings.caldav_id, "+
		"spaces.id, spaces.location_id, spaces.name, "+
		"locations.id, locations.organization_id, locations.name, locations.description, locations.tz, "+
		"users.email "+
		"FROM bookings "+
		"INNER JOIN spaces ON bookings.space_id = spaces.id "+
		"INNER JOIN locations ON spaces.location_id = locations.id "+
		"INNER JOIN users ON bookings.user_id = users.id "+
		"WHERE user_id = $1 AND leave_time >= $2 "+
		"ORDER BY enter_time", userID, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &BookingDetails{}
		err = rows.Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.CalDavID, &e.Space.ID, &e.Space.LocationID, &e.Space.Name, &e.Space.Location.ID, &e.Space.Location.OrganizationID, &e.Space.Location.Name, &e.Space.Location.Description, &e.Space.Location.Timezone, &e.UserEmail)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
func (r *BookingRepository) Update(e *Booking) error {
	_, err := GetDatabase().DB().Exec("UPDATE bookings SET "+
		"user_id = $1, "+
		"space_id = $2, "+
		"enter_time = $3, "+
		"leave_time = $4, "+
		"caldav_id = $5 "+
		"WHERE id = $6",
		e.UserID, e.SpaceID, e.Enter, e.Leave, e.CalDavID, e.ID)
	return err
}

func (r *BookingRepository) Delete(e *BookingDetails) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM bookings WHERE id = $1", e.ID)
	return err
}

func (r *BookingRepository) GetCount(organizationID string) (int, error) {
	var res int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(bookings.id) "+
		"FROM bookings "+
		"INNER JOIN spaces ON spaces.id = bookings.space_id "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE locations.organization_id = $1",
		organizationID).Scan(&res)
	return res, err
}

func (r *BookingRepository) GetCountDateRange(organizationID string, enter, leave time.Time) (int, error) {
	var res int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(bookings.id) "+
		"FROM bookings "+
		"INNER JOIN spaces ON spaces.id = bookings.space_id "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE locations.organization_id = $1 AND ("+
		"($2 BETWEEN enter_time AND leave_time) OR "+
		"($3 BETWEEN enter_time AND leave_time) OR "+
		"(enter_time BETWEEN $2 AND $3) OR "+
		"(leave_time BETWEEN $2 AND $3)"+
		")",
		organizationID, enter, leave).Scan(&res)
	return res, err
}

func (r *BookingRepository) GetTotalBookedMinutes(organizationID string, enter, leave time.Time) (int, error) {
	var totalBookedMinutes float64
	err := GetDatabase().DB().QueryRow("SELECT SUM(EXTRACT(EPOCH FROM (LEAST(leave_time, $3) - GREATEST(enter_time, $2)))/60) "+
		"FROM bookings "+
		"INNER JOIN spaces ON spaces.id = bookings.space_id "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE locations.organization_id = $1 AND ("+
		"($2 BETWEEN enter_time AND leave_time) OR "+
		"($3 BETWEEN enter_time AND leave_time) OR "+
		"(enter_time BETWEEN $2 AND $3) OR "+
		"(leave_time BETWEEN $2 AND $3)"+
		")",
		organizationID, enter, leave).Scan(&totalBookedMinutes)
	return int(math.RoundToEven(totalBookedMinutes)), err
}

func (r *BookingRepository) GetLoad(organizationID string, enter, leave time.Time) (int, error) {
	totalBookedMinutes, err := r.GetTotalBookedMinutes(organizationID, enter, leave)
	if err != nil {
		return 0, err
	}
	numSpaces, err := GetSpaceRepository().GetCount(organizationID)
	if err != nil {
		return 0, err
	}
	totalTimeMinutes := leave.Sub(enter).Minutes() * float64(numSpaces)
	res := float64(totalBookedMinutes) / float64(totalTimeMinutes) * float64(100)
	if res > 100.0 {
		res = 100.0
	}
	return int(math.RoundToEven(res)), nil
}

// various scenarios
//
//	   |-----------|    (base)
//	|------|            (overlap start)
//	           |------| (overlap end)
//	 |----------------| (bigger than)
//	       |---|        (within)
//
// bigger than should be covered by overlap start / end checks
//
// get all bookings by a specific user which overlap with the provided time range
func (r *BookingRepository) GetTimeRangeByUser(userID string, enter time.Time, leave time.Time, excludeBookingID string) ([]*Booking, error) {
	var result []*Booking
	rows, err := GetDatabase().DB().Query("SELECT id, user_id, space_id, enter_time, leave_time, caldav_id "+
		"FROM bookings "+
		"WHERE id::text != $4 AND user_id = $1 AND ("+
		"($2 <= enter_time AND $3 > enter_time) OR "+ // (overlap start, can end at same time as next start)
		"($2 < leave_time AND $3 >= leave_time) OR "+ // (overlap end, start can equal previous leave time)
		"($2 >= enter_time AND $3 <= leave_time)"+ // (within)
		") "+
		"ORDER BY enter_time", userID, enter, leave, excludeBookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Booking{}
		err = rows.Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.CalDavID)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

// GetConflicts returns bookings for a specific space which overlap
// with the specified enter and leave times.
func (r *BookingRepository) GetConflicts(spaceID string, enter time.Time, leave time.Time, excludeBookingID string) ([]*Booking, error) {
	var result []*Booking
	rows, err := GetDatabase().DB().Query("SELECT id, user_id, space_id, enter_time, leave_time, caldav_id "+
		"FROM bookings "+
		"WHERE id::text != $1 AND space_id = $2 AND ("+
		"($3 >= enter_time AND $3 <= leave_time) OR "+
		"($4 >= enter_time AND $4 <= leave_time) OR "+
		"(enter_time >= $3 AND enter_time <= $4) OR "+
		"(leave_time >= $3 AND leave_time <= $4)"+
		") "+
		"ORDER BY enter_time", excludeBookingID, spaceID, enter, leave)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Booking{}
		err = rows.Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.CalDavID)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

// GetConcurrent returns concurrent bookings for a specific location
// within the specified enter and leave times.
func (r *BookingRepository) GetConcurrent(location *Location, enter time.Time, leave time.Time, excludeBookingID string) (int, error) {
	var getNumActive = func(bookings []*Booking, timestamp time.Time) int {
		res := 0
		for _, b := range bookings {
			if b.Enter.Before(timestamp) && b.Leave.After(timestamp) && !b.Enter.Equal(timestamp) && !b.Leave.Equal(timestamp) {
				res++
			}
		}
		return res
	}

	var result []*Booking
	tz := GetLocationRepository().GetTimezone(location)
	targetTz, err := time.LoadLocation(tz)
	if err != nil {
		return 0, err
	}
	rows, err := GetDatabase().DB().Query("SELECT id, user_id, space_id, enter_time, leave_time "+
		"FROM bookings "+
		"WHERE id::text != $1 AND space_id IN (SELECT id FROM spaces WHERE location_id = $2) AND ("+
		"($3 >= enter_time AND $3 <= leave_time) OR "+
		"($4 >= enter_time AND $4 <= leave_time) OR "+
		"(enter_time >= $3 AND enter_time <= $4) OR "+
		"(leave_time >= $3 AND leave_time <= $4)"+
		") "+
		"ORDER BY enter_time", excludeBookingID, location.ID, enter, leave)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Booking{}
		err = rows.Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave)
		e.Enter, _ = time.ParseInLocation(JsDateTimeFormat, e.Enter.Format(JsDateTimeFormat), targetTz)
		e.Leave, _ = time.ParseInLocation(JsDateTimeFormat, e.Leave.Format(JsDateTimeFormat), targetTz)
		if err != nil {
			return 0, err
		}
		result = append(result, e)
	}

	max := 0
	timestamp := enter
	for timestamp.Before(leave) || timestamp.Equal(leave) {
		numActive := getNumActive(result, timestamp)
		if numActive > max {
			max = numActive
		}
		timestamp = timestamp.Add(time.Minute * 1)
	}

	return max, nil
}

func (r *BookingRepository) GetPresenceReport(organizationID string, location *Location, start time.Time, end time.Time, maxResults, offset int) ([]*BookingPresenceItem, error) {
	// Build list of users to include in report
	users, err := GetUserRepository().GetAll(organizationID, maxResults, offset)
	if err != nil {
		return nil, err
	}
	userIds := make([]string, len(users))
	for i, user := range users {
		userIds[i] = user.ID
	}

	// Prepare array of days to report
	var times []time.Time
	curTime := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	var cols strings.Builder
	const DateFormat string = "2006-01-02"
	for curTime.Before(end) {
		times = append(times, curTime)
		cols.WriteString(", ")
		cols.WriteString("(SELECT COUNT(*) FROM bookings b2 WHERE b2.user_id = b.user_id AND DATE(b2.enter_time) = '" + curTime.Format(DateFormat) + "'::DATE)")
		curTime = curTime.AddDate(0, 0, 1)
	}

	// Prepare result
	res := make([]*BookingPresenceItem, len(users))
	for i, user := range users {
		presence := make(map[string]int)
		for _, time := range times {
			presence[time.Format(DateFormat)] = 0
		}
		item := &BookingPresenceItem{
			User:     user,
			Presence: presence,
		}
		res[i] = item
	}

	// Build query
	conditions := ""
	if location != nil {
		conditions = "AND b.space_id IN (SELECT id FROM spaces WHERE location_id = $2) "
	}
	stm := "SELECT b.user_id" + cols.String() + " " +
		"FROM bookings b " +
		"WHERE b.user_id = ANY($1) " + conditions +
		"GROUP BY b.user_id"
	var rows *sql.Rows
	if location != nil {
		rows, err = GetDatabase().DB().Query(stm, pq.Array(userIds), location.ID)
	} else {
		rows, err = GetDatabase().DB().Query(stm, pq.Array(userIds))
	}
	if err == sql.ErrNoRows {
		return res, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		// Scan row
		dest := make([]interface{}, len(times)+1)
		dest[0] = new(string)
		for i := 1; i < len(times)+1; i++ {
			dest[i] = new(int)
		}
		if err := rows.Scan(dest...); err != nil {
			continue
		}

		// Get
		var item *BookingPresenceItem = nil
		for _, i := range res {
			if i.User.ID == *(dest[0].(*string)) {
				item = i
			}
		}
		if item == nil {
			continue
		}
		for i, time := range times {
			item.Presence[time.Format(DateFormat)] = *(dest[i+1].(*int))
		}
	}
	return res, nil
}
