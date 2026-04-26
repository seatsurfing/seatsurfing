package repository

import (
	"sync"
)

type ExchangeBookingMappingRepository struct {
}

type ExchangeBookingMapping struct {
	BookingID       string
	ExchangeEventID string
	RoomEmail       string
}

var exchangeBookingMappingRepository *ExchangeBookingMappingRepository
var exchangeBookingMappingRepositoryOnce sync.Once

func GetExchangeBookingMappingRepository() *ExchangeBookingMappingRepository {
	exchangeBookingMappingRepositoryOnce.Do(func() {
		exchangeBookingMappingRepository = &ExchangeBookingMappingRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS exchange_booking_mapping (" +
			"booking_id UUID NOT NULL, " +
			"exchange_event_id VARCHAR NOT NULL, " +
			"room_email VARCHAR NOT NULL, " +
			"PRIMARY KEY (booking_id))")
		if err != nil {
			panic(err)
		}
	})
	return exchangeBookingMappingRepository
}

func (r *ExchangeBookingMappingRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
}

func (r *ExchangeBookingMappingRepository) Create(bookingID, exchangeEventID, roomEmail string) error {
	_, err := GetDatabase().DB().Exec("INSERT INTO exchange_booking_mapping "+
		"(booking_id, exchange_event_id, room_email) "+
		"VALUES ($1, $2, $3) "+
		"ON CONFLICT (booking_id) DO UPDATE SET exchange_event_id = EXCLUDED.exchange_event_id, room_email = EXCLUDED.room_email",
		bookingID, exchangeEventID, roomEmail)
	return err
}

func (r *ExchangeBookingMappingRepository) GetByBookingID(bookingID string) (*ExchangeBookingMapping, error) {
	e := &ExchangeBookingMapping{}
	err := GetDatabase().DB().QueryRow("SELECT booking_id, exchange_event_id, room_email "+
		"FROM exchange_booking_mapping "+
		"WHERE booking_id = $1",
		bookingID).Scan(&e.BookingID, &e.ExchangeEventID, &e.RoomEmail)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *ExchangeBookingMappingRepository) Delete(bookingID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM exchange_booking_mapping WHERE booking_id = $1", bookingID)
	return err
}
