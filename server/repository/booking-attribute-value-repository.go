package repository

import (
	"sync"

	"github.com/lib/pq"
)

type BookingAttributeValueRepository struct {
}

type BookingAttributeValue struct {
	BookingID   string
	AttributeID string
	Label       string
	Type        BookingAttributeType
	Value       string
}

var bookingAttributeValueRepository *BookingAttributeValueRepository
var bookingAttributeValueRepositoryOnce sync.Once

func GetBookingAttributeValueRepository() *BookingAttributeValueRepository {
	bookingAttributeValueRepositoryOnce.Do(func() {
		bookingAttributeValueRepository = &BookingAttributeValueRepository{}
		// Requires the bookings and booking_attributes tables to exist.
		GetBookingRepository()
		GetBookingAttributeRepository()
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS booking_attribute_values (" +
			"booking_id uuid NOT NULL REFERENCES bookings(id) ON DELETE CASCADE, " +
			"attribute_id uuid NOT NULL REFERENCES booking_attributes(id) ON DELETE CASCADE, " +
			"type INTEGER NOT NULL, " +
			"value VARCHAR NOT NULL DEFAULT '', " +
			"PRIMARY KEY (booking_id, attribute_id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_booking_attribute_values_attribute ON booking_attribute_values(attribute_id)")
		if err != nil {
			panic(err)
		}
	})
	return bookingAttributeValueRepository
}

func (r *BookingAttributeValueRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// nothing yet
}

// Set stores a value together with the attribute's current type, but only if
// the attribute belongs to the booking's organization.
func (r *BookingAttributeValueRepository) Set(bookingID, attributeID, value string) error {
	_, err := GetDatabase().DB().Exec("INSERT INTO booking_attribute_values (booking_id, attribute_id, type, value) "+
		"SELECT bookings.id, booking_attributes.id, booking_attributes.type, $3 "+
		"FROM bookings "+
		"INNER JOIN booking_attributes ON booking_attributes.organization_id = bookings.organization_id "+
		"WHERE bookings.id = $1 AND booking_attributes.id = $2 "+
		"ON CONFLICT (booking_id, attribute_id) DO UPDATE SET type = EXCLUDED.type, value = EXCLUDED.value",
		bookingID, attributeID, value)
	return err
}

func (r *BookingAttributeValueRepository) GetAllForBooking(bookingID string) ([]*BookingAttributeValue, error) {
	return r.GetAllForBookingList([]string{bookingID})
}

func (r *BookingAttributeValueRepository) GetAllForBookingList(bookingIDs []string) ([]*BookingAttributeValue, error) {
	result := []*BookingAttributeValue{}
	if len(bookingIDs) == 0 {
		return result, nil
	}
	rows, err := GetDatabase().DB().Query("SELECT booking_attribute_values.booking_id, booking_attributes.id, booking_attributes.label, booking_attribute_values.type, booking_attribute_values.value "+
		"FROM booking_attribute_values "+
		"INNER JOIN booking_attributes ON booking_attributes.id = booking_attribute_values.attribute_id "+
		"INNER JOIN bookings ON bookings.id = booking_attribute_values.booking_id AND bookings.organization_id = booking_attributes.organization_id "+
		"WHERE booking_attribute_values.booking_id = ANY($1::uuid[]) "+
		"ORDER BY booking_attributes.position, booking_attributes.label",
		pq.StringArray(bookingIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &BookingAttributeValue{}
		err = rows.Scan(&e.BookingID, &e.AttributeID, &e.Label, &e.Type, &e.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
