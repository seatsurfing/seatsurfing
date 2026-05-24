package repository

import (
	"log"
	"sync"
)

type LocationFloorPlanRepository struct {
}

type LocationFloorPlan struct {
	LocationID     string
	OrganizationID string
	DesignData     string
}

var locationFloorPlanRepository *LocationFloorPlanRepository
var locationFloorPlanRepositoryOnce sync.Once

func GetLocationFloorPlanRepository() *LocationFloorPlanRepository {
	locationFloorPlanRepositoryOnce.Do(func() {
		locationFloorPlanRepository = &LocationFloorPlanRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS location_floor_plans (" +
			"location_id uuid PRIMARY KEY REFERENCES locations(id) ON DELETE CASCADE, " +
			"organization_id uuid NOT NULL, " +
			"design_data TEXT NOT NULL DEFAULT '{}')")
		if err != nil {
			log.Println(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_location_floor_plans_org ON location_floor_plans(organization_id)")
		if err != nil {
			log.Println(err)
		}
	})
	return locationFloorPlanRepository
}

func (r *LocationFloorPlanRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
}

func (r *LocationFloorPlanRepository) GetDesign(locationID string) (*LocationFloorPlan, error) {
	e := &LocationFloorPlan{}
	err := GetDatabase().DB().QueryRow(
		"SELECT location_id, organization_id, design_data FROM location_floor_plans WHERE location_id = $1",
		locationID).Scan(&e.LocationID, &e.OrganizationID, &e.DesignData)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *LocationFloorPlanRepository) SetDesign(e *LocationFloorPlan) error {
	_, err := GetDatabase().DB().Exec(
		"INSERT INTO location_floor_plans (location_id, organization_id, design_data) "+
			"VALUES ($1, $2, $3) "+
			"ON CONFLICT (location_id) DO UPDATE SET design_data = EXCLUDED.design_data, organization_id = EXCLUDED.organization_id",
		e.LocationID, e.OrganizationID, e.DesignData)
	return err
}

func (r *LocationFloorPlanRepository) Delete(locationID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM location_floor_plans WHERE location_id = $1", locationID)
	return err
}
