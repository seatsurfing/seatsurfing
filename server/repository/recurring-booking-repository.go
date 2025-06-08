package repository

import (
	"encoding/json"
	"errors"
	"slices"
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

type RecurringBookingRepository struct {
}

type Cadence int

const (
	CadenceDaily  Cadence = 1
	CadenceWeekly Cadence = 2
)

type RecurringBooking struct {
	ID      string
	UserID  string
	SpaceID string
	Enter   time.Time
	Leave   time.Time
	Subject string
	Cadence Cadence
	Details interface{}
	End     time.Time
}

type CadenceDailyDetails struct {
	Cycle int `json:"cycle"`
}

type CadenceWeeklyDetails struct {
	Cycle    int            `json:"cycle"`
	Weekdays []time.Weekday `json:"weekdays"`
}

var recurringBookingRepository *RecurringBookingRepository
var recurringBookingRepositoryOnce sync.Once

func GetRecurringBookingRepository() *RecurringBookingRepository {
	recurringBookingRepositoryOnce.Do(func() {
		recurringBookingRepository = &RecurringBookingRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS recurring_bookings (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"user_id uuid NOT NULL, " +
			"space_id uuid NOT NULL, " +
			"enter_time TIMESTAMP NOT NULL, " +
			"leave_time TIMESTAMP NOT NULL, " +
			"subject VARCHAR NOT NULL DEFAULT '', " +
			"cadence INT NOT NULL, " +
			"details VARCHAR, " +
			"end_date TIMESTAMP NOT NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_recurring_bookings_user_id ON recurring_bookings(user_id)")
		if err != nil {
			panic(err)
		}
	})
	return recurringBookingRepository
}

func (r *RecurringBookingRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// Nothing to do for now
}

func (r *RecurringBookingRepository) Create(e *RecurringBooking) error {
	var id string
	details, err := json.Marshal(&e.Details)
	if err != nil {
		return err
	}
	err = GetDatabase().DB().QueryRow("INSERT INTO recurring_bookings "+
		"(user_id, space_id, enter_time, leave_time, subject, cadence, details, end_date) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) "+
		"RETURNING id",
		e.UserID, e.SpaceID, e.Enter, e.Leave, e.Subject, e.Cadence, details, e.End).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *RecurringBookingRepository) GetOne(id string) (*RecurringBooking, error) {
	e := &RecurringBooking{}
	var details []byte
	err := GetDatabase().DB().QueryRow("SELECT id, user_id, space_id, enter_time, leave_time, subject, cadence, details, end_date "+
		"FROM recurring_bookings "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.Subject, &e.Cadence, &details, &e.End)
	if err != nil {
		return nil, err
	}
	e.Details, err = r.getCadenceDetails(e.Cadence, details)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *RecurringBookingRepository) Delete(e *RecurringBooking) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM recurring_bookings WHERE id = $1", e.ID)
	return err
}

func (r *RecurringBookingRepository) CreateBookings(e *RecurringBooking) []*Booking {
	res := make([]*Booking, 0)
	cur := e.Enter
	for cur.Before(e.End) {
		booking := &Booking{
			UserID:      e.UserID,
			SpaceID:     e.SpaceID,
			Enter:       cur,
			Leave:       cur.Add(e.Leave.Sub(e.Enter)),
			Subject:     e.Subject,
			RecurringID: NullString(e.ID),
		}
		res = append(res, booking)
		cur = r.getNextBookingTime(e, cur)
	}
	return res
}

func (r *RecurringBookingRepository) getNextBookingTime(e *RecurringBooking, current time.Time) time.Time {
	if e.Cadence == CadenceDaily {
		return current.AddDate(0, 0, e.Details.(*CadenceDailyDetails).Cycle)
	} else if e.Cadence == CadenceWeekly {
		weekdays := e.Details.(*CadenceWeeklyDetails).Weekdays
		idx := slices.Index(weekdays, current.Weekday())
		if idx == len(weekdays)-1 {
			diff := 7 - weekdays[idx] + weekdays[0]
			return current.AddDate(0, 0, int(diff)+7*(e.Details.(*CadenceWeeklyDetails).Cycle-1))
		} else {
			diff := weekdays[idx+1] - weekdays[idx]
			return current.AddDate(0, 0, int(diff))
		}
	}
	return time.Time{}
}

func (r *RecurringBookingRepository) getCadenceDetails(cadence Cadence, details []byte) (interface{}, error) {
	if cadence == CadenceDaily {
		var dailyDetails CadenceDailyDetails
		if err := json.Unmarshal(details, &dailyDetails); err != nil {
			return nil, err
		}
		return dailyDetails, nil
	} else if cadence == CadenceWeekly {
		var weeklyDetails CadenceWeeklyDetails
		if err := json.Unmarshal(details, &weeklyDetails); err != nil {
			return nil, err
		}
		return weeklyDetails, nil
	} else {
		return nil, errors.New("unknown cadence type")
	}
}
