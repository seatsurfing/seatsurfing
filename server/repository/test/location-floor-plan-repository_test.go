package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestLocationFloorPlanCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	location := &Location{
		OrganizationID: org.ID,
		Name:           "TestLocation",
	}
	if err := GetLocationRepository().Create(location); err != nil {
		t.Fatal(err)
	}

	// GetDesign on non-existent record should return error
	_, err := GetLocationFloorPlanRepository().GetDesign(location.ID)
	CheckTestBool(t, false, err == nil)

	// SetDesign (insert)
	designData := `{"version":1,"elements":[]}`
	plan := &LocationFloorPlan{
		LocationID:     location.ID,
		OrganizationID: org.ID,
		DesignData:     designData,
	}
	err = GetLocationFloorPlanRepository().SetDesign(plan)
	CheckTestBool(t, true, err == nil)

	// GetDesign should now return the stored data
	got, err := GetLocationFloorPlanRepository().GetDesign(location.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, designData, got.DesignData)
	CheckTestString(t, location.ID, got.LocationID)
	CheckTestString(t, org.ID, got.OrganizationID)

	// SetDesign again (upsert)
	updatedDesignData := `{"version":1,"elements":[{"id":"abc","type":"wall","x1":0,"y1":0,"x2":100,"y2":0,"thickness":8}]}`
	plan.DesignData = updatedDesignData
	err = GetLocationFloorPlanRepository().SetDesign(plan)
	CheckTestBool(t, true, err == nil)

	got, err = GetLocationFloorPlanRepository().GetDesign(location.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, updatedDesignData, got.DesignData)

	// Delete
	err = GetLocationFloorPlanRepository().Delete(location.ID)
	CheckTestBool(t, true, err == nil)

	_, err = GetLocationFloorPlanRepository().GetDesign(location.ID)
	CheckTestBool(t, false, err == nil)
}

func TestLocationFloorPlanDeletedWithLocation(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	location := &Location{
		OrganizationID: org.ID,
		Name:           "TestLocation",
	}
	if err := GetLocationRepository().Create(location); err != nil {
		t.Fatal(err)
	}

	plan := &LocationFloorPlan{
		LocationID:     location.ID,
		OrganizationID: org.ID,
		DesignData:     `{"version":1,"elements":[]}`,
	}
	if err := GetLocationFloorPlanRepository().SetDesign(plan); err != nil {
		t.Fatal(err)
	}

	// Deleting the location should cascade-delete the floor plan
	if err := GetLocationRepository().Delete(location); err != nil {
		t.Fatal(err)
	}

	_, err := GetLocationFloorPlanRepository().GetDesign(location.ID)
	CheckTestBool(t, false, err == nil)
}

func TestLocationMapTypeField(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	location := &Location{
		OrganizationID: org.ID,
		Name:           "TestLocation",
	}
	if err := GetLocationRepository().Create(location); err != nil {
		t.Fatal(err)
	}

	// Default map_type should be empty
	got, err := GetLocationRepository().GetOne(location.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, "", got.MapType)

	// Update with map_type = "designed"
	got.MapType = "designed"
	if err := GetLocationRepository().Update(got); err != nil {
		t.Fatal(err)
	}

	updated, err := GetLocationRepository().GetOne(location.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, "designed", updated.MapType)

	// Update back to empty
	updated.MapType = ""
	if err := GetLocationRepository().Update(updated); err != nil {
		t.Fatal(err)
	}

	final, err := GetLocationRepository().GetOne(location.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, "", final.MapType)
}
