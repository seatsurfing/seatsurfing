package main

import (
	"sync"
)

type SpaceAttributeValueRepository struct {
}

type SpaceAttributeValueEntityType int

const (
	SpaceAttributeValueEntityTypeLocation SpaceAttributeValueEntityType = 1
	SpaceAttributeValueEntityTypeSpace    SpaceAttributeValueEntityType = 2
)

type SpaceAttributeValue struct {
	AttributeID string
	EntityID    string
	EntityType  SpaceAttributeValueEntityType
	Value       string
}

var spaceAttributeValueRepository *SpaceAttributeValueRepository
var spaceAttributeValueRepositoryOnce sync.Once

func GetSpaceAttributeValueRepository() *SpaceAttributeValueRepository {
	spaceAttributeValueRepositoryOnce.Do(func() {
		spaceAttributeValueRepository = &SpaceAttributeValueRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS space_attribute_values (" +
			"attribute_id uuid NOT NULL, " +
			"entity_id uuid NOT NULL, " +
			"entity_type INTEGER NOT NULL, " +
			"value VARCHAR NOT NULL DEFAULT '', " +
			"PRIMARY KEY (attribute_id, entity_id, entity_type))")
		if err != nil {
			panic(err)
		}
	})
	return spaceAttributeValueRepository
}

func (r *SpaceAttributeValueRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// nothing yet
}

func (r *SpaceAttributeValueRepository) Set(attributeID string, entityID string, entityType SpaceAttributeValueEntityType, value string) error {
	_, err := GetDatabase().DB().Exec("INSERT INTO space_attribute_values (attribute_id, entity_id, entity_type, value) "+
		"VALUES ($1, $2, $3, $4) "+
		"ON CONFLICT (attribute_id, entity_id, entity_type) DO UPDATE SET value = $4",
		attributeID, entityID, entityType, value)
	return err
}

func (r *SpaceAttributeValueRepository) Get(attributeID string, entityID string, entityType SpaceAttributeValueEntityType) (string, error) {
	var res string
	err := GetDatabase().DB().QueryRow("SELECT value FROM space_attribute_values "+
		"FROM space_attribute_values "+
		"WHERE attribute_id = $1 AND entity_id = $2 AND entity_type = $3",
		attributeID, entityID, entityType).Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (r *SpaceAttributeValueRepository) GetAllForEntity(entityID string, entityType SpaceAttributeValueEntityType) ([]*SpaceAttributeValue, error) {
	var result []*SpaceAttributeValue
	rows, err := GetDatabase().DB().Query("SELECT attribute_id, entity_id, entity_type, value "+
		"FROM space_attribute_values "+
		"WHERE entity_id = $1 AND entity_type = $2",
		entityID, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &SpaceAttributeValue{}
		err = rows.Scan(&e.AttributeID, &e.EntityID, &e.EntityType, &e.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *SpaceAttributeValueRepository) GetAll(OrganizationID string, entityType SpaceAttributeValueEntityType) ([]*SpaceAttributeValue, error) {
	join := "LEFT JOIN locations ON space_attribute_values.entity_id = locations.id"
	if entityType == SpaceAttributeValueEntityTypeSpace {
		join = "LEFT JOIN spaces ON space_attribute_values.entity_id = spaces.id " + join
	}
	var result []*SpaceAttributeValue
	rows, err := GetDatabase().DB().Query("SELECT attribute_id, entity_id, entity_type, value "+
		"FROM space_attribute_values "+
		join+" "+
		"WHERE organization_id = $1 AND entity_type = $2",
		OrganizationID, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &SpaceAttributeValue{}
		err = rows.Scan(&e.AttributeID, &e.EntityID, &e.EntityType, &e.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *SpaceAttributeValueRepository) Delete(attributeID string, entityID string, entityType SpaceAttributeValueEntityType) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM space_attribute_values WHERE attribute_id = $1 AND entity_id = $2 AND entity_type = $3",
		attributeID, entityID, entityType)
	return err
}
