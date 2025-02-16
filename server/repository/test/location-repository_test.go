package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestLocationsCount(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	l1 := &Location{
		OrganizationID: org.ID,
		Name:           "L1",
	}
	GetLocationRepository().Create(l1)

	res, err := GetLocationRepository().GetCount(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, res)
}
