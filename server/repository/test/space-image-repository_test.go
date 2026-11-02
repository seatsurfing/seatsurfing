package test

import (
	"database/sql"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestSpaceImageCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	_, space2 := CreateTestLocationAndSpace(org)
	repo := GetSpaceImageRepository()

	_, err := repo.Get(space.ID, org.ID)
	CheckTestBool(t, true, err == sql.ErrNoRows)

	data := CreateTestPNG(4, 3)
	err = repo.Set(&SpaceImage{SpaceID: space.ID, OrganizationID: org.ID, MimeType: "image/png", Width: 4, Height: 3, Data: data})
	CheckTestIsNil(t, err)
	e, err := repo.Get(space.ID, org.ID)
	CheckTestIsNil(t, err)
	CheckTestString(t, "image/png", e.MimeType)
	CheckTestUint(t, 4, e.Width)
	CheckTestUint(t, 3, e.Height)
	CheckTestInt(t, len(data), len(e.Data))

	// Upsert replaces the existing image
	err = repo.Set(&SpaceImage{SpaceID: space.ID, OrganizationID: org.ID, MimeType: "image/jpeg", Width: 10, Height: 20, Data: []byte{1, 2, 3}})
	CheckTestIsNil(t, err)
	e, _ = repo.Get(space.ID, org.ID)
	CheckTestString(t, "image/jpeg", e.MimeType)
	CheckTestInt(t, 3, len(e.Data))

	ids, err := repo.GetSpaceIDsWithImage([]string{space.ID, space2.ID}, org.ID)
	CheckTestIsNil(t, err)
	CheckTestBool(t, true, ids[space.ID])
	CheckTestBool(t, false, ids[space2.ID])

	CheckTestIsNil(t, repo.Delete(space.ID, org.ID))
	_, err = repo.Get(space.ID, org.ID)
	CheckTestBool(t, true, err == sql.ErrNoRows)
}

func TestSpaceImageOrgIsolation(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")
	_, space := CreateTestLocationAndSpace(org)
	repo := GetSpaceImageRepository()
	repo.Set(&SpaceImage{SpaceID: space.ID, OrganizationID: org.ID, MimeType: "image/png", Width: 1, Height: 1, Data: CreateTestPNG(1, 1)})

	_, err := repo.Get(space.ID, org2.ID)
	CheckTestBool(t, true, err == sql.ErrNoRows)
	ids, _ := repo.GetSpaceIDsWithImage([]string{space.ID}, org2.ID)
	CheckTestBool(t, false, ids[space.ID])
	repo.Delete(space.ID, org2.ID)
	_, err = repo.Get(space.ID, org.ID)
	CheckTestIsNil(t, err)
}

func TestSpaceImageDeletedWithSpace(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	repo := GetSpaceImageRepository()
	repo.Set(&SpaceImage{SpaceID: space.ID, OrganizationID: org.ID, MimeType: "image/png", Width: 1, Height: 1, Data: CreateTestPNG(1, 1)})

	CheckTestIsNil(t, GetSpaceRepository().Delete(space))
	_, err := repo.Get(space.ID, org.ID)
	CheckTestBool(t, true, err == sql.ErrNoRows)
}
