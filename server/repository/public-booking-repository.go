package repository

import (
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

type PublicBookingStore struct {
}

var publicBookingRepository *PublicBookingStore
var publicBookingRepositoryOnce sync.Once

func GetPublicBookingRepository() *PublicBookingStore {
	publicBookingRepositoryOnce.Do(func() {
		publicBookingRepository = &PublicBookingStore{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS public_bookings (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"external_id uuid NOT NULL UNIQUE DEFAULT uuid_generate_v4(), " +
			"name VARCHAR NOT NULL, " +
			"email VARCHAR NOT NULL, " +
			"language VARCHAR NULL, " +
			"created_at_utc TIMESTAMP NOT NULL DEFAULT NOW(), " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
	})
	return publicBookingRepository
}

func (r *PublicBookingStore) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet - the table (including the language column) is created
	// fresh above, so there's nothing for existing installs to migrate.
}

func (r *PublicBookingStore) Create(e *PublicBooking) error {
	var id, externalID string
	err := GetDatabase().DB().QueryRow("INSERT INTO public_bookings "+
		"(name, email, language, created_at_utc) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id, external_id",
		e.Name, e.Email, e.Language, time.Now().UTC()).Scan(&id, &externalID)
	if err != nil {
		return err
	}
	e.ID = id
	e.ExternalID = externalID
	return nil
}

func (r *PublicBookingStore) GetOne(id string) (*PublicBooking, error) {
	e := &PublicBooking{}
	err := GetDatabase().DB().QueryRow("SELECT id, external_id, name, email, COALESCE(language, ''), created_at_utc "+
		"FROM public_bookings "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.ExternalID, &e.Name, &e.Email, &e.Language, &e.CreatedAtUTC)
	if err != nil {
		return nil, err
	}
	return e, nil
}
