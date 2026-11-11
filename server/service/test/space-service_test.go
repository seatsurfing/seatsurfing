package test

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestSpaceServiceAvailabilityRedactsOtherBookers(t *testing.T) {
	org, user, location, space := setupServiceTest(t)
	other := CreateTestUserInOrg(org)
	space2 := &Space{LocationID: location.ID, Enabled: true}
	GetSpaceRepository().Create(space2)
	enter, leave := serviceTestSlot(0, 8, 17)
	if _, bErr := GetBookingService().CreateBooking(other, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	if _, bErr := GetBookingService().CreateBooking(user, &BookingInput{SpaceID: space2.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	location, _ = GetLocationRepository().GetOne(location.ID)
	enter, _ = GetLocationRepository().AttachTimezoneInformation(enter, location)
	leave, _ = GetLocationRepository().AttachTimezoneInformation(leave, location)

	list, err := GetSpaceService().GetAvailabilityForUser(user, location, "", enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(list))
	for _, s := range list {
		CheckTestBool(t, false, s.Available)
		CheckTestInt(t, 1, len(s.Bookings))
		if s.Space.ID == space.ID {
			CheckTestString(t, "", s.Bookings[0].UserEmail)
		} else {
			CheckTestString(t, user.Email, s.Bookings[0].UserEmail)
		}
	}

	GetSettingsRepository().Set(org.ID, SettingShowNames.Name, "1")
	list, _ = GetSpaceService().GetAvailabilityForUser(user, location, space.ID, enter, leave, nil)
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, other.Email, list[0].Bookings[0].UserEmail)
}

func TestSpaceServiceRequiresApproval(t *testing.T) {
	org, user, _, space := setupServiceTest(t)
	group := CreateTestGroup(org, user)
	GetSpaceRepository().AddApprovers(space, []string{group.ID})
	CheckTestBool(t, false, GetSpaceService().RequiresApproval(org.ID, space))
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")
	CheckTestBool(t, true, GetSpaceService().RequiresApproval(org.ID, space))
}

func TestSpaceServiceAvailabilityApprovalRequired(t *testing.T) {
	org, user, location, space := setupServiceTest(t)
	group := CreateTestGroup(org, user)
	GetSpaceRepository().AddApprovers(space, []string{group.ID})
	enter, leave := serviceTestSlot(0, 8, 17)
	location, _ = GetLocationRepository().GetOne(location.ID)
	enter, _ = GetLocationRepository().AttachTimezoneInformation(enter, location)
	leave, _ = GetLocationRepository().AttachTimezoneInformation(leave, location)

	// Groups feature disabled: approver groups have no effect
	list, err := GetSpaceService().GetAvailabilityForUser(user, location, space.ID, enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, false, list[0].ApprovalRequired)

	// Groups feature enabled
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")
	list, err = GetSpaceService().GetAvailabilityForUser(user, location, space.ID, enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, true, list[0].ApprovalRequired)

	// Deleting the approver group removes the approval requirement
	if err := GetGroupRepository().Delete(group); err != nil {
		t.Fatal(err)
	}
	list, err = GetSpaceService().GetAvailabilityForUser(user, location, space.ID, enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, false, list[0].ApprovalRequired)
}

func TestSpaceServicePublicAvailability(t *testing.T) {
	org, _, location, space := setupServiceTest(t)
	space.PublicBookingEnabled = true
	GetSpaceRepository().Update(space)
	space2 := &Space{LocationID: location.ID, Enabled: true, PublicBookingEnabled: true}
	GetSpaceRepository().Create(space2)
	disabledSpace := &Space{LocationID: location.ID, Enabled: false, PublicBookingEnabled: true}
	GetSpaceRepository().Create(disabledSpace)
	GetBookingRepository().Create(&Booking{
		SpaceID: space.ID,
		Enter:   time.Date(2030, 1, 2, 9, 0, 0, 0, time.UTC),
		Leave:   time.Date(2030, 1, 2, 17, 0, 0, 0, time.UTC),
	})

	res, err := GetSpaceService().GetPublicAvailability(org.ID, time.Date(2030, 1, 2, 16, 0, 0, 0, time.UTC), time.Date(2030, 1, 2, 18, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(res))
	CheckTestBool(t, false, res[space.ID])
	CheckTestBool(t, true, res[space2.ID])

	res, _ = GetSpaceService().GetPublicAvailability(org.ID, time.Date(2030, 1, 2, 17, 0, 0, 0, time.UTC), time.Date(2030, 1, 2, 18, 0, 0, 0, time.UTC))
	CheckTestBool(t, true, res[space.ID])
}

func TestSpaceServicePublicAvailabilityShortBookingAtCapacity(t *testing.T) {
	org, _, location, space := setupServiceTest(t)
	location.MaxConcurrentBookings = 1
	GetLocationRepository().Update(location)
	space.PublicBookingEnabled = true
	GetSpaceRepository().Update(space)
	space2 := &Space{LocationID: location.ID, Enabled: true, PublicBookingEnabled: true}
	GetSpaceRepository().Create(space2)
	GetBookingRepository().Create(&Booking{
		SpaceID: space.ID,
		Enter:   time.Date(2030, 1, 2, 9, 0, 0, 0, time.UTC),
		Leave:   time.Date(2030, 1, 2, 9, 1, 0, 0, time.UTC),
	})

	res, err := GetSpaceService().GetPublicAvailability(org.ID, time.Date(2030, 1, 2, 9, 0, 0, 0, time.UTC), time.Date(2030, 1, 2, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, false, res[space.ID])
	CheckTestBool(t, false, res[space2.ID])

	res, _ = GetSpaceService().GetPublicAvailability(org.ID, time.Date(2030, 1, 2, 9, 1, 0, 0, time.UTC), time.Date(2030, 1, 2, 10, 0, 0, 0, time.UTC))
	CheckTestBool(t, true, res[space.ID])
	CheckTestBool(t, true, res[space2.ID])
}

func TestSpaceServiceIsPublicSlotAvailable(t *testing.T) {
	_, _, location, space := setupServiceTest(t)
	location.Timezone = "UTC"
	GetLocationRepository().Update(location)
	space2 := &Space{LocationID: location.ID, Enabled: true, PublicBookingEnabled: true}
	GetSpaceRepository().Create(space2)
	at := func(day, h, m int) time.Time {
		return time.Date(2030, 1, day, h, m, 0, 0, time.UTC)
	}
	GetBookingRepository().Create(&Booking{SpaceID: space.ID, Enter: at(2, 9, 0), Leave: at(2, 9, 1)})

	// Free slot
	available, err := GetSpaceService().IsPublicSlotAvailable(space, location, at(2, 10, 0), at(2, 11, 0))
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, true, available)

	// Conflicting booking on the same space
	available, _ = GetSpaceService().IsPublicSlotAvailable(space, location, at(2, 9, 0), at(2, 10, 0))
	CheckTestBool(t, false, available)
	available, _ = GetSpaceService().IsPublicSlotAvailable(space2, location, at(2, 9, 0), at(2, 10, 0))
	CheckTestBool(t, true, available)

	// Concurrent booking limit reached by a short booking on another space
	location.MaxConcurrentBookings = 1
	GetLocationRepository().Update(location)
	available, _ = GetSpaceService().IsPublicSlotAvailable(space2, location, at(2, 9, 0), at(2, 10, 0))
	CheckTestBool(t, false, available)
	available, _ = GetSpaceService().IsPublicSlotAvailable(space2, location, at(2, 9, 1), at(2, 10, 0))
	CheckTestBool(t, true, available)

	// Weekday not bookable (2030-01-02 is a Wednesday, 2030-01-01 a Tuesday)
	location.BookableDays = "1,2"
	GetLocationRepository().Update(location)
	available, _ = GetSpaceService().IsPublicSlotAvailable(space2, location, at(2, 10, 0), at(2, 11, 0))
	CheckTestBool(t, false, available)
	available, _ = GetSpaceService().IsPublicSlotAvailable(space2, location, at(1, 10, 0), at(1, 11, 0))
	CheckTestBool(t, true, available)
}

func TestSpaceServiceImageCRUD(t *testing.T) {
	org, user, location, space := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingFeatureSpaceImages.Name, "1")
	admin := CreateTestUserOrgAdmin(org)
	svc := GetSpaceService()

	_, err := svc.GetImage(user, location, space, false)
	CheckTestBool(t, true, err == ErrSpaceImageNotFound)

	CheckTestIsNil(t, svc.SetImage(admin, location, space, CreateTestPNG(40, 30)))
	e, err := svc.GetImage(user, location, space, false)
	CheckTestIsNil(t, err)
	CheckTestString(t, "image/png", e.MimeType)
	CheckTestUint(t, 40, e.Width)
	CheckTestUint(t, 30, e.Height)

	enter, leave := serviceTestSlot(0, 8, 17)
	location, _ = GetLocationRepository().GetOne(location.ID)
	enter, _ = GetLocationRepository().AttachTimezoneInformation(enter, location)
	leave, _ = GetLocationRepository().AttachTimezoneInformation(leave, location)
	list, err := svc.GetAvailabilityForUser(user, location, space.ID, enter, leave, nil)
	CheckTestIsNil(t, err)
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, true, list[0].HasImage)

	CheckTestIsNil(t, svc.DeleteImage(admin, location, space))
	_, err = svc.GetImage(user, location, space, false)
	CheckTestBool(t, true, err == ErrSpaceImageNotFound)
}

func TestSpaceServiceImageRequiresAreaAdmin(t *testing.T) {
	org, user, location, space := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingFeatureSpaceImages.Name, "1")
	admin := CreateTestUserOrgAdmin(org)
	svc := GetSpaceService()

	CheckTestBool(t, true, svc.SetImage(user, location, space, CreateTestPNG(1, 1)) == ErrSpaceImageForbidden)
	CheckTestIsNil(t, svc.SetImage(admin, location, space, CreateTestPNG(1, 1)))
	CheckTestBool(t, true, svc.DeleteImage(user, location, space) == ErrSpaceImageForbidden)
	_, err := svc.GetImage(user, location, space, false)
	CheckTestIsNil(t, err)
}

func TestSpaceServiceImageForeignOrg(t *testing.T) {
	org, _, location, space := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingFeatureSpaceImages.Name, "1")
	admin := CreateTestUserOrgAdmin(org)
	org2 := CreateTestOrg("test2.com")
	admin2 := CreateTestUserOrgAdmin(org2)
	location2, _ := CreateTestLocationAndSpace(org2)
	svc := GetSpaceService()
	CheckTestIsNil(t, svc.SetImage(admin, location, space, CreateTestPNG(1, 1)))

	CheckTestBool(t, true, svc.SetImage(admin2, location, space, CreateTestPNG(1, 1)) == ErrSpaceImageForbidden)
	CheckTestBool(t, true, svc.DeleteImage(admin2, location, space) == ErrSpaceImageForbidden)
	_, err := svc.GetImage(admin2, location, space, false)
	CheckTestBool(t, true, err == ErrSpaceImageForbidden)

	// A space cannot be addressed through a location it does not belong to.
	CheckTestBool(t, true, svc.SetImage(admin2, location2, space, CreateTestPNG(1, 1)) == ErrSpaceImageNotFound)
	_, err = svc.GetImage(admin2, location2, space, false)
	CheckTestBool(t, true, err == ErrSpaceImageNotFound)
}

func TestSpaceServiceImageFeatureDisabled(t *testing.T) {
	org, user, location, space := setupServiceTest(t)
	admin := CreateTestUserOrgAdmin(org)
	svc := GetSpaceService()

	CheckTestBool(t, true, svc.SetImage(admin, location, space, CreateTestPNG(1, 1)) == ErrSpaceImageNotEnabled)

	GetSettingsRepository().Set(org.ID, SettingFeatureSpaceImages.Name, "1")
	CheckTestIsNil(t, svc.SetImage(admin, location, space, CreateTestPNG(1, 1)))
	GetSettingsRepository().Set(org.ID, SettingFeatureSpaceImages.Name, "0")
	_, err := svc.GetImage(user, location, space, false)
	CheckTestBool(t, true, err == ErrSpaceImageNotFound)
	CheckTestIsNil(t, svc.DeleteImage(admin, location, space))
}

func TestSpaceServiceImageValidation(t *testing.T) {
	org, _, location, space := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingFeatureSpaceImages.Name, "1")
	admin := CreateTestUserOrgAdmin(org)
	svc := GetSpaceService()

	CheckTestBool(t, true, svc.SetImage(admin, location, space, []byte("not an image")) == ErrSpaceImageInvalidFormat)
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"></svg>`)
	CheckTestBool(t, true, svc.SetImage(admin, location, space, svg) == ErrSpaceImageInvalidFormat)
	var gifData bytes.Buffer
	CheckTestIsNil(t, gif.Encode(&gifData, image.NewPaletted(image.Rect(0, 0, 10, 10), color.Palette{color.Black}), nil))
	CheckTestBool(t, true, svc.SetImage(admin, location, space, gifData.Bytes()) == ErrSpaceImageInvalidFormat)
	CheckTestBool(t, true, svc.SetImage(admin, location, space, CreateTestPNG(SpaceImageMaxDimension+1, 1)) == ErrSpaceImageInvalidFormat)
	tooLarge := make([]byte, SpaceImageMaxBytes+1)
	CheckTestBool(t, true, svc.SetImage(admin, location, space, tooLarge) == ErrSpaceImageTooLarge)
	_, err := svc.GetImage(admin, location, space, false)
	CheckTestBool(t, true, err == ErrSpaceImageNotFound)
}

func TestSpaceServiceImageRejectsTooManyPixels(t *testing.T) {
	org, _, location, space := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingFeatureSpaceImages.Name, "1")
	admin := CreateTestUserOrgAdmin(org)
	svc := GetSpaceService()

	data := setTestPNGSize(CreateTestPNG(1, 1), SpaceImageMaxDimension, SpaceImageMaxDimension)
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	CheckTestIsNil(t, err)
	CheckTestInt(t, SpaceImageMaxDimension, cfg.Width)
	CheckTestBool(t, true, svc.SetImage(admin, location, space, data) == ErrSpaceImageInvalidFormat)
	_, err = svc.GetImage(admin, location, space, false)
	CheckTestBool(t, true, err == ErrSpaceImageNotFound)
}

func TestSpaceServiceImageRejectsTruncated(t *testing.T) {
	org, _, location, space := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingFeatureSpaceImages.Name, "1")
	admin := CreateTestUserOrgAdmin(org)
	svc := GetSpaceService()
	CheckTestIsNil(t, svc.SetImage(admin, location, space, CreateTestPNG(40, 30)))

	var jpg bytes.Buffer
	jpeg.Encode(&jpg, image.NewRGBA(image.Rect(0, 0, 64, 64)), nil)
	for _, data := range [][]byte{CreateTestPNG(40, 30), jpg.Bytes()} {
		truncated := data[:len(data)-20]
		_, _, err := image.DecodeConfig(bytes.NewReader(truncated))
		CheckTestIsNil(t, err)
		CheckTestBool(t, true, svc.SetImage(admin, location, space, truncated) == ErrSpaceImageInvalidFormat)
	}

	e, err := svc.GetImage(admin, location, space, false)
	CheckTestIsNil(t, err)
	CheckTestUint(t, 40, e.Width)
	CheckTestUint(t, 30, e.Height)
}

func setTestPNGSize(data []byte, width, height uint32) []byte {
	res := bytes.Clone(data)
	// IHDR chunk type at 12, width at 16, height at 20, CRC at 29
	binary.BigEndian.PutUint32(res[16:], width)
	binary.BigEndian.PutUint32(res[20:], height)
	binary.BigEndian.PutUint32(res[29:], crc32.ChecksumIEEE(res[12:29]))
	return res
}
