package router

import (
	"encoding/json"
	"log"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

func enqueueExchangeSync(booking *Booking, operation string, exchangeEventID string) {
	// Resolve space and location
	space, err := GetSpaceRepository().GetOne(booking.SpaceID)
	if err != nil {
		return
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		return
	}

	// Check org Exchange settings
	settings, err := GetSettingsRepository().GetExchangeSettings(location.OrganizationID)
	if err != nil || !settings.Enabled {
		return
	}

	// Check space mapping
	mapping, err := GetExchangeSpaceMappingRepository().GetBySpaceID(space.ID)
	if err != nil || mapping == nil || mapping.RoomEmail == "" {
		return
	}

	// Resolve timezone at enqueue time
	tz := GetLocationRepository().GetTimezone(location)

	// Resolve user info
	user, err := GetUserRepository().GetOne(booking.UserID)
	if err != nil {
		log.Println("enqueueExchangeSync: error getting user:", err)
		return
	}

	payload := ExchangeSyncPayload{
		OrgID:           location.OrganizationID,
		BookingID:       booking.ID,
		Operation:       operation,
		RoomEmail:       mapping.RoomEmail,
		ExchangeEventID: exchangeEventID,
		Enter:           booking.Enter.UTC(),
		Leave:           booking.Leave.UTC(),
		LocationTZ:      tz,
		UserFirstname:   user.Firstname,
		UserLastname:    user.Lastname,
		SpaceName:       space.Name,
		LocationName:    location.Name,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Println("enqueueExchangeSync: error marshalling payload:", err)
		return
	}

	if err := GetExchangeSyncQueueRepository().Enqueue(booking.ID, operation, string(payloadJSON)); err != nil {
		log.Println("enqueueExchangeSync: error enqueueing job:", err)
	}
}

func enqueueExchangeCreate(booking *Booking) {
	enqueueExchangeSync(booking, ExchangeOpCreate, "")
}

func enqueueExchangeUpdate(booking *Booking) {
	enqueueExchangeSync(booking, ExchangeOpUpdate, "")
}

func enqueueExchangeDelete(booking *Booking) {
	// Look up exchange event ID before the booking row is gone
	var exchangeEventID string
	bm, err := GetExchangeBookingMappingRepository().GetByBookingID(booking.ID)
	if err == nil && bm != nil {
		exchangeEventID = bm.ExchangeEventID
	}
	enqueueExchangeSync(booking, ExchangeOpDelete, exchangeEventID)
}
