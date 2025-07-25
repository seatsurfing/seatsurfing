package repository

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
)

type SpaceRepository struct {
}

type Space struct {
	ID             string
	LocationID     string
	Name           string
	X              uint
	Y              uint
	Width          uint
	Height         uint
	Rotation       uint
	RequireSubject bool
}

type SpaceAvailabilityBookingEntry struct {
	BookingID   string
	RecurringID string
	UserID      string
	UserEmail   string
	Enter       time.Time
	Leave       time.Time
	Subject     string
}

type SpaceAvailability struct {
	Space
	Available bool
	Bookings  []*SpaceAvailabilityBookingEntry
}

type SpaceDetails struct {
	Location Location
	Space
}

type SpaceGroup struct {
	SpaceID string
	GroupID string
}

var spaceRepository *SpaceRepository
var spaceRepositoryOnce sync.Once

func GetSpaceRepository() *SpaceRepository {
	spaceRepositoryOnce.Do(func() {
		spaceRepository = &SpaceRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS spaces (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"location_id uuid NOT NULL, " +
			"name VARCHAR NOT NULL, " +
			"x INTEGER, " +
			"y INTEGER, " +
			"width INTEGER, " +
			"height INTEGER, " +
			"rotation INTEGER, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_spaces_location_id ON spaces(location_id)")
		if err != nil {
			panic(err)
		}
	})
	return spaceRepository
}

func (r *SpaceRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 22 {
		if _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS spaces_approvers (" +
			"space_id uuid NOT NULL, " +
			"group_id uuid NOT NULL, " +
			"PRIMARY KEY (space_id, group_id))"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS spaces_allowed_bookers (" +
			"space_id uuid NOT NULL, " +
			"group_id uuid NOT NULL, " +
			"PRIMARY KEY (space_id, group_id))"); err != nil {
			panic(err)
		}
	}
	if curVersion < 23 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE spaces " +
			"ADD COLUMN IF NOT EXISTS require_subject BOOLEAN DEFAULT TRUE"); err != nil {
			panic(err)
		}
	}
}

func (r *SpaceRepository) Create(e *Space) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO spaces "+
		"(name, location_id, x, y, width, height, rotation, require_subject) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) "+
		"RETURNING id",
		e.Name, e.LocationID, e.X, e.Y, e.Width, e.Height, e.Rotation, e.RequireSubject).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *SpaceRepository) GetOne(id string) (*Space, error) {
	e := &Space{}
	err := GetDatabase().DB().QueryRow("SELECT id, location_id, name, x, y, width, height, rotation, require_subject "+
		"FROM spaces "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.LocationID, &e.Name, &e.X, &e.Y, &e.Width, &e.Height, &e.Rotation, &e.RequireSubject)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *SpaceRepository) GetAllInTime(locationID string, enter, leave time.Time) ([]*SpaceAvailability, error) {
	var result []*SpaceAvailability
	subQueryWhere := "bookings.space_id = spaces.id AND (" +
		"($1 >= bookings.enter_time AND $1 <= bookings.leave_time) OR " +
		"($2 >= bookings.enter_time AND $2 <= bookings.leave_time) OR " +
		"(bookings.enter_time >= $1 AND bookings.enter_time <= $2) OR " +
		"(bookings.leave_time >= $1 AND bookings.leave_time <= $2)" +
		")"
	rows, err := GetDatabase().DB().Query("SELECT id, location_id, name, x, y, width, height, rotation, require_subject, "+
		"NOT EXISTS(SELECT id FROM bookings WHERE "+subQueryWhere+"), "+
		"ARRAY(SELECT CONCAT(users.id, '@@@', users.email, '@@@', bookings.enter_time, '@@@', bookings.leave_time, '@@@', bookings.id, '@@@', bookings.subject, '@@@', bookings.recurring_id) FROM bookings INNER JOIN users ON users.id = bookings.user_id WHERE "+subQueryWhere+" ORDER BY bookings.enter_time ASC) "+
		"FROM spaces "+
		"WHERE location_id = $3 "+
		"ORDER BY name", enter, leave, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &SpaceAvailability{}
		var bookingUserNames []string
		err = rows.Scan(&e.ID, &e.LocationID, &e.Name, &e.X, &e.Y, &e.Width, &e.Height, &e.Rotation, &e.RequireSubject, &e.Available, pq.Array(&bookingUserNames))
		for _, bookingUserName := range bookingUserNames {
			tokens := strings.Split(bookingUserName, "@@@")
			timeFormat := "2006-01-02 15:04:05"
			enter, _ := time.Parse(timeFormat, tokens[2])
			leave, _ := time.Parse(timeFormat, tokens[3])
			entry := &SpaceAvailabilityBookingEntry{
				BookingID:   tokens[4],
				UserID:      tokens[0],
				UserEmail:   tokens[1],
				Enter:       enter,
				Leave:       leave,
				Subject:     tokens[5],
				RecurringID: tokens[6],
			}
			e.Bookings = append(e.Bookings, entry)
		}
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *SpaceRepository) GetByKeyword(organizationID string, keyword string) ([]*Space, error) {
	var result []*Space
	rows, err := GetDatabase().DB().Query("SELECT spaces.id, spaces.location_id, spaces.name, spaces.x, spaces.y, spaces.width, spaces.height, spaces.rotation, spaces.require_subject "+
		"FROM spaces "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE locations.organization_id = $1 AND LOWER(spaces.name) LIKE '%' || $2 || '%'"+
		"ORDER BY spaces.name", organizationID, strings.ToLower(keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Space{}
		err = rows.Scan(&e.ID, &e.LocationID, &e.Name, &e.X, &e.Y, &e.Width, &e.Height, &e.Rotation, &e.RequireSubject)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *SpaceRepository) GetAll(locationID string) ([]*Space, error) {
	var result []*Space
	rows, err := GetDatabase().DB().Query("SELECT id, location_id, name, x, y, width, height, rotation, require_subject "+
		"FROM spaces "+
		"WHERE location_id = $1 "+
		"ORDER BY name", locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Space{}
		err = rows.Scan(&e.ID, &e.LocationID, &e.Name, &e.X, &e.Y, &e.Width, &e.Height, &e.Rotation, &e.RequireSubject)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
func (r *SpaceRepository) Update(e *Space) error {
	_, err := GetDatabase().DB().Exec("UPDATE spaces SET "+
		"location_id = $1, "+
		"name = $2, "+
		"x = $3, "+
		"y = $4, "+
		"width = $5, "+
		"height = $6, "+
		"rotation = $7, "+
		"require_subject = $8 "+
		"WHERE id = $9",
		e.LocationID, e.Name, e.X, e.Y, e.Width, e.Height, e.Rotation, e.RequireSubject, e.ID)
	return err
}

func (r *SpaceRepository) Delete(e *Space) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM space_attribute_values WHERE entity_id = $1 AND entity_type = $2", e.ID, SpaceAttributeValueEntityTypeSpace); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM spaces_approvers WHERE space_id = $1", e.ID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM spaces_allowed_bookers WHERE space_id = $1", e.ID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM spaces WHERE id = $1", e.ID)
	return err
}

func (r *SpaceRepository) GetCount(organizationID string) (int, error) {
	var res int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(spaces.id) "+
		"FROM spaces "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE locations.organization_id = $1",
		organizationID).Scan(&res)
	return res, err
}

func (r *SpaceRepository) GetTotalCountMap(organizationID string) (map[string]int, error) {
	res := make(map[string]int)
	rows, err := GetDatabase().DB().Query("SELECT spaces.location_id, COUNT(spaces.id) "+
		"FROM spaces "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE locations.organization_id = $1 "+
		"GROUP BY spaces.location_id",
		organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var locationId string
		var count int
		err = rows.Scan(&locationId, &count)
		if err != nil {
			return nil, err
		}
		res[locationId] = count
	}
	return res, nil
}

func (r *SpaceRepository) GetFreeCountMap(organizationID string, enter, leave time.Time) (map[string]int, error) {
	res := make(map[string]int)
	locations, _ := GetLocationRepository().GetAll(organizationID)
	for _, location := range locations {
		enterNew, err := GetLocationRepository().AttachTimezoneInformation(enter, location)
		if err != nil {
			return nil, err
		}
		leaveNew, err := GetLocationRepository().AttachTimezoneInformation(leave, location)
		if err != nil {
			return nil, err
		}
		spaces, _ := r.GetAllInTime(location.ID, enterNew, leaveNew)
		res[location.ID] = 0
		for _, space := range spaces {
			if space.Available {
				res[location.ID]++
			}
		}
	}
	return res, nil
}

func (r *SpaceRepository) GetBookingUserIDMap(organizationID string, enter, leave time.Time) (map[string][]string, error) {
	res := make(map[string][]string)
	locations, _ := GetLocationRepository().GetAll(organizationID)
	for _, location := range locations {
		enterNew, err := GetLocationRepository().AttachTimezoneInformation(enter, location)
		if err != nil {
			return nil, err
		}
		leaveNew, err := GetLocationRepository().AttachTimezoneInformation(leave, location)
		if err != nil {
			return nil, err
		}
		spaces, _ := r.GetAllInTime(location.ID, enterNew, leaveNew)
		res[location.ID] = make([]string, 0)
		for _, space := range spaces {
			for _, booking := range space.Bookings {
				res[location.ID] = append(res[location.ID], booking.UserID)
			}
		}
	}
	return res, nil
}

func (r *SpaceRepository) GetApproverGroupIDs(spaceID string) ([]string, error) {
	var result []string
	rows, err := GetDatabase().DB().Query("SELECT group_id "+
		"FROM spaces_approvers "+
		"WHERE space_id = $1 "+
		"ORDER BY group_id",
		spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

func (r *SpaceRepository) AddApprovers(e *Space, groupIDs []string) error {
	if len(groupIDs) == 0 {
		return nil
	}
	sqlStr := "INSERT INTO spaces_approvers (space_id, group_id) VALUES "
	vals := []interface{}{}
	i := 1
	for _, groupID := range groupIDs {
		sqlStr += fmt.Sprintf("($%d, $%d),", i, i+1)
		i += 2
		vals = append(vals, e.ID, groupID)
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	_, err := GetDatabase().DB().Exec(sqlStr, vals...)
	return err
}

func (r *SpaceRepository) RemoveApprovers(e *Space, groupIDs []string) error {
	if len(groupIDs) == 0 {
		return nil
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM spaces_approvers WHERE space_id = $1 AND group_id = ANY($2)", e.ID, pq.Array(groupIDs))
	return err
}

func (r *SpaceRepository) GetAllApproversForSpaceList(spaceIDs []string) ([]*SpaceGroup, error) {
	var result []*SpaceGroup
	rows, err := GetDatabase().DB().Query("SELECT space_id, group_id "+
		"FROM spaces_approvers "+
		"WHERE space_id = ANY($1::uuid[])",
		pq.StringArray(spaceIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &SpaceGroup{}
		err = rows.Scan(&e.SpaceID, &e.GroupID)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *SpaceRepository) GetAllowedBookersGroupIDs(e *Space) ([]string, error) {
	var result []string
	rows, err := GetDatabase().DB().Query("SELECT group_id "+
		"FROM spaces_allowed_bookers "+
		"WHERE space_id = $1 "+
		"ORDER BY group_id",
		e.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

func (r *SpaceRepository) AddAllowedBookers(e *Space, groupIDs []string) error {
	if len(groupIDs) == 0 {
		return nil
	}
	sqlStr := "INSERT INTO spaces_allowed_bookers (space_id, group_id) VALUES "
	vals := []interface{}{}
	i := 1
	for _, groupID := range groupIDs {
		sqlStr += fmt.Sprintf("($%d, $%d),", i, i+1)
		i += 2
		vals = append(vals, e.ID, groupID)
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	_, err := GetDatabase().DB().Exec(sqlStr, vals...)
	return err
}

func (r *SpaceRepository) RemoveAllowedBookers(e *Space, groupIDs []string) error {
	if len(groupIDs) == 0 {
		return nil
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM spaces_allowed_bookers WHERE space_id = $1 AND group_id = ANY($2)", e.ID, pq.Array(groupIDs))
	return err
}

func (r *SpaceRepository) GetAllAllowedBookersForSpaceList(spaceIDs []string) ([]*SpaceGroup, error) {
	var result []*SpaceGroup
	rows, err := GetDatabase().DB().Query("SELECT space_id, group_id "+
		"FROM spaces_allowed_bookers "+
		"WHERE space_id = ANY($1::uuid[])",
		pq.StringArray(spaceIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &SpaceGroup{}
		err = rows.Scan(&e.SpaceID, &e.GroupID)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
