package repository

import (
	"database/sql"
	"sync"
)

type BookingAttributeRepository struct {
}

type BookingAttributeType int

const (
	BookingAttributeTypeText          BookingAttributeType = 1
	BookingAttributeTypeTextMultiline BookingAttributeType = 2
	BookingAttributeTypeNumber        BookingAttributeType = 3
	BookingAttributeTypeCheckbox      BookingAttributeType = 4
)

type BookingAttribute struct {
	ID             string
	OrganizationID string
	Label          string
	Type           BookingAttributeType
	Required       bool
	Position       int
}

var bookingAttributeRepository *BookingAttributeRepository
var bookingAttributeRepositoryOnce sync.Once

func GetBookingAttributeRepository() *BookingAttributeRepository {
	bookingAttributeRepositoryOnce.Do(func() {
		bookingAttributeRepository = &BookingAttributeRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS booking_attributes (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"organization_id uuid NOT NULL, " +
			"label VARCHAR NOT NULL, " +
			"type INTEGER NOT NULL, " +
			"required boolean NOT NULL DEFAULT FALSE, " +
			"position INTEGER NOT NULL DEFAULT 0, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_booking_attributes_org ON booking_attributes(organization_id)")
		if err != nil {
			panic(err)
		}
	})
	return bookingAttributeRepository
}

func (r *BookingAttributeRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// nothing yet
}

// CreateIfBelowLimit inserts the attribute unless its organization already
// has maxCount attributes. The check and the insert are a single statement.
func (r *BookingAttributeRepository) CreateIfBelowLimit(e *BookingAttribute, maxCount int) (bool, error) {
	var id string
	var position int
	err := GetDatabase().DB().QueryRow("INSERT INTO booking_attributes "+
		"(organization_id, label, type, required, position) "+
		"SELECT $1, $2, $3, $4, COALESCE((SELECT MAX(position) FROM booking_attributes WHERE organization_id = $1), 0) + 1 "+
		"WHERE (SELECT COUNT(*) FROM booking_attributes WHERE organization_id = $1) < $5 "+
		"RETURNING id, position",
		e.OrganizationID, e.Label, e.Type, e.Required, maxCount).Scan(&id, &position)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	e.ID = id
	e.Position = position
	return true, nil
}

func (r *BookingAttributeRepository) GetOne(id string) (*BookingAttribute, error) {
	e := &BookingAttribute{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, label, type, required, position "+
		"FROM booking_attributes "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.OrganizationID, &e.Label, &e.Type, &e.Required, &e.Position)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *BookingAttributeRepository) GetAll(organizationID string) ([]*BookingAttribute, error) {
	result := []*BookingAttribute{}
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, label, type, required, position "+
		"FROM booking_attributes "+
		"WHERE organization_id = $1 "+
		"ORDER BY position, label", organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &BookingAttribute{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Label, &e.Type, &e.Required, &e.Position)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *BookingAttributeRepository) Update(e *BookingAttribute) error {
	_, err := GetDatabase().DB().Exec("UPDATE booking_attributes SET "+
		"label = $1, "+
		"type = $2, "+
		"required = $3 "+
		"WHERE id = $4 AND organization_id = $5",
		e.Label, e.Type, e.Required, e.ID, e.OrganizationID)
	return err
}

// SetPositions renumbers the given attributes of the organization in the
// order of the ID list.
func (r *BookingAttributeRepository) SetPositions(organizationID string, ids []string) error {
	tx, err := GetDatabase().DB().Begin()
	if err != nil {
		return err
	}
	for i, id := range ids {
		if _, err := tx.Exec("UPDATE booking_attributes SET position = $1 WHERE id = $2 AND organization_id = $3", i+1, id, organizationID); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *BookingAttributeRepository) Delete(e *BookingAttribute) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM booking_attributes WHERE id = $1 AND organization_id = $2", e.ID, e.OrganizationID)
	return err
}

func (r *BookingAttributeRepository) DeleteAll(organizationID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM booking_attributes WHERE organization_id = $1", organizationID)
	return err
}
