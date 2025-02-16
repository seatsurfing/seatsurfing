package repository

import (
	"strconv"
	"sync"
)

type SpaceAttributeRepository struct {
}

type SpaceAttribute struct {
	ID                 string
	OrganizationID     string
	Label              string
	Type               SettingType
	SpaceApplicable    bool
	LocationApplicable bool
}

var spaceAttributeRepository *SpaceAttributeRepository
var spaceAttributeRepositoryOnce sync.Once

func GetSpaceAttributeRepository() *SpaceAttributeRepository {
	spaceAttributeRepositoryOnce.Do(func() {
		spaceAttributeRepository = &SpaceAttributeRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS space_attributes (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"organization_id uuid NOT NULL, " +
			"label VARCHAR NOT NULL, " +
			"type INTEGER DEFAULT " + strconv.Itoa(int(SettingTypeString)) + "," +
			"space_applicable boolean NOT NULL DEFAULT FALSE, " +
			"location_applicable boolean NOT NULL DEFAULT FALSE, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
	})
	return spaceAttributeRepository
}

func (r *SpaceAttributeRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// nothing yet
}

func (r *SpaceAttributeRepository) Create(e *SpaceAttribute) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO space_attributes "+
		"(organization_id, label, type, space_applicable, location_applicable) "+
		"VALUES ($1, $2, $3, $4, $5) "+
		"RETURNING id",
		e.OrganizationID, e.Label, e.Type, e.SpaceApplicable, e.LocationApplicable).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *SpaceAttributeRepository) GetOne(id string) (*SpaceAttribute, error) {
	e := &SpaceAttribute{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, label, type, space_applicable, location_applicable "+
		"FROM space_attributes "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.OrganizationID, &e.Label, &e.Type, &e.SpaceApplicable, &e.LocationApplicable)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *SpaceAttributeRepository) GetAll(organizationID string) ([]*SpaceAttribute, error) {
	var result []*SpaceAttribute
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, label, type, space_applicable, location_applicable "+
		"FROM space_attributes "+
		"WHERE organization_id = $1 "+
		"ORDER BY label", organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &SpaceAttribute{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Label, &e.Type, &e.SpaceApplicable, &e.LocationApplicable)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *SpaceAttributeRepository) Update(e *SpaceAttribute) error {
	_, err := GetDatabase().DB().Exec("UPDATE space_attributes SET "+
		"organization_id = $1, "+
		"label = $2, "+
		"type = $3, "+
		"space_applicable = $4, "+
		"location_applicable = $5 "+
		"WHERE id = $6",
		e.OrganizationID, e.Label, e.Type, e.SpaceApplicable, e.LocationApplicable, e.ID)
	return err
}

func (r *SpaceAttributeRepository) Delete(e *SpaceAttribute) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM space_attribute_values WHERE attribute_id = $1", e.ID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM space_attributes WHERE id = $1", e.ID)
	return err
}
