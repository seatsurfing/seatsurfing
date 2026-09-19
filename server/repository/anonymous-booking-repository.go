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
			"name VARCHAR NOT NULL, " +
			"email VARCHAR NOT NULL, " +
			"language VARCHAR NULL, " +
			"created_at_utc TIMESTAMP NOT NULL DEFAULT NOW(), " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
	})
	return anonymousBookingRepository
}

func (r *AnonymousBookingStore) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet - the table (including the language column) is created
	// fresh above, so there's nothing for existing installs to migrate.
}

func (r *AnonymousBookingStore) Create(e *AnonymousBooking) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO anonymous_bookings "+
		"(name, email, language, created_at_utc) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		e.Name, e.Email, e.Language, time.Now().UTC()).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *AnonymousBookingStore) GetOne(id string) (*AnonymousBooking, error) {
	e := &AnonymousBooking{}
	err := GetDatabase().DB().QueryRow("SELECT id, name, email, COALESCE(language, ''), created_at_utc "+
		"FROM anonymous_bookings "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.Name, &e.Email, &e.Language, &e.CreatedAtUTC)
	if err != nil {
		return nil, err
	}
	return e, nil
}
