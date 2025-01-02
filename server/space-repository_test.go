package main

import (
	"testing"
)

func TestSpacesCount(t *testing.T) {
	clearTestDB()
	org := createTestOrg("test.com")

	l1 := &Location{
		OrganizationID: org.ID,
		Name:           "L1",
	}
	GetLocationRepository().Create(l1)

	s1 := &Space{
		LocationID: l1.ID,
		Name:       "S1",
	}
	GetSpaceRepository().Create(s1)
	s2 := &Space{
		LocationID: l1.ID,
		Name:       "S2",
	}
	GetSpaceRepository().Create(s2)

	res, err := GetSpaceRepository().GetCount(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	checkTestInt(t, 2, res)
}

func TestSpacesCountMap(t *testing.T) {
	clearTestDB()
	org := createTestOrg("test.com")

	l1 := &Location{OrganizationID: org.ID, Name: "L1"}
	l2 := &Location{OrganizationID: org.ID, Name: "L2"}
	GetLocationRepository().Create(l1)
	GetLocationRepository().Create(l2)

	GetSpaceRepository().Create(&Space{LocationID: l1.ID, Name: "S1.1"})
	GetSpaceRepository().Create(&Space{LocationID: l1.ID, Name: "S1.2"})
	GetSpaceRepository().Create(&Space{LocationID: l1.ID, Name: "S1.3"})
	GetSpaceRepository().Create(&Space{LocationID: l2.ID, Name: "S2.1"})
	GetSpaceRepository().Create(&Space{LocationID: l2.ID, Name: "S2.2"})

	res, err := GetSpaceRepository().GetTotalCountMap(org.ID)
	checkTestBool(t, true, err == nil)
	checkTestInt(t, 2, len(res))
	checkTestInt(t, 3, res[l1.ID])
	checkTestInt(t, 2, res[l2.ID])
}
