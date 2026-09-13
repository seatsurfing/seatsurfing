package repository

import (
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

type AnonymousBookingStore struct {
}

var anonymousBookingRepository *AnonymousBookingStore
var anonymousBookingRepositoryOnce sync.Once

func GetAnonymousBookingRepository() *AnonymousBookingStore {
	anonymousBookingRepositoryOnce.Do(func() {
		anonymousBookingRepository = &AnonymousBookingStore{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS anonymous_bookings (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"booking_id uuid NOT NULL UNIQUE REFERENCES bookings(id) ON DELETE CASCADE, " +
			"name VARCHAR NOT NULL, " +
			"email VARCHAR NOT NULL, " +
			"created_at_utc TIMESTAMP NOT NULL DEFAULT NOW(), " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
	})
	return anonymousBookingRepository
}

func (r *AnonymousBookingStore) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *AnonymousBookingStore) Create(e *AnonymousBooking) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO anonymous_bookings "+
		"(booking_id, name, email, created_at_utc) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		e.BookingID, e.Name, e.Email, time.Now().UTC()).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *AnonymousBookingStore) GetByBookingID(bookingID string) (*AnonymousBooking, error) {
	e := &AnonymousBooking{}
	err := GetDatabase().DB().QueryRow("SELECT id, booking_id, name, email, created_at_utc "+
		"FROM anonymous_bookings "+
		"WHERE booking_id = $1",
		bookingID).Scan(&e.ID, &e.BookingID, &e.Name, &e.Email, &e.CreatedAtUTC)
	if err != nil {
		return nil, err
	}
	return e, nil
}
