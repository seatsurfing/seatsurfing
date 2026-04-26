package repository

import (
	"sync"
)

type ExchangeSpaceMappingRepository struct {
}

type ExchangeSpaceMapping struct {
	SpaceID   string
	RoomEmail string
}

var exchangeSpaceMappingRepository *ExchangeSpaceMappingRepository
var exchangeSpaceMappingRepositoryOnce sync.Once

func GetExchangeSpaceMappingRepository() *ExchangeSpaceMappingRepository {
	exchangeSpaceMappingRepositoryOnce.Do(func() {
		exchangeSpaceMappingRepository = &ExchangeSpaceMappingRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS exchange_space_mapping (" +
			"space_id UUID NOT NULL REFERENCES spaces(id) ON DELETE CASCADE, " +
			"room_email VARCHAR NOT NULL DEFAULT '', " +
			"PRIMARY KEY (space_id))")
		if err != nil {
			panic(err)
		}
	})
	return exchangeSpaceMappingRepository
}

func (r *ExchangeSpaceMappingRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
}

func (r *ExchangeSpaceMappingRepository) Upsert(e *ExchangeSpaceMapping) error {
	_, err := GetDatabase().DB().Exec("INSERT INTO exchange_space_mapping "+
		"(space_id, room_email) "+
		"VALUES ($1, $2) "+
		"ON CONFLICT (space_id) DO UPDATE SET room_email = EXCLUDED.room_email",
		e.SpaceID, e.RoomEmail)
	return err
}

func (r *ExchangeSpaceMappingRepository) GetBySpaceID(spaceID string) (*ExchangeSpaceMapping, error) {
	e := &ExchangeSpaceMapping{}
	err := GetDatabase().DB().QueryRow("SELECT space_id, room_email "+
		"FROM exchange_space_mapping "+
		"WHERE space_id = $1",
		spaceID).Scan(&e.SpaceID, &e.RoomEmail)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *ExchangeSpaceMappingRepository) Delete(spaceID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM exchange_space_mapping WHERE space_id = $1", spaceID)
	return err
}
