package api

import (
	"testing"
	"time"
)

func TestUserRoundTrip(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)
	u := &User{
		ID:                     "u1",
		OrganizationID:         "o1",
		Email:                  "test@example.com",
		Firstname:              "Test",
		Lastname:               "User",
		AtlassianID:            "atl-1",
		HashedPassword:         "hash",
		AuthProviderID:         "auth-1",
		PasswordPending:        true,
		PasswordUpdateRequired: true,
		Role:                   UserRoleOrgAdmin,
		Disabled:               true,
		BanExpiry:              &now,
		LastActivityAtUTC:      &now,
		TotpSecret:             "secret",
		ApiToken:               "token",
	}
	got := userFromProto(userToProto(u))
	if got.BanExpiry == nil || !got.BanExpiry.Equal(*u.BanExpiry) {
		t.Errorf("BanExpiry mismatch: got=%v want=%v", got.BanExpiry, u.BanExpiry)
	}
	if got.LastActivityAtUTC == nil || !got.LastActivityAtUTC.Equal(*u.LastActivityAtUTC) {
		t.Errorf("LastActivityAtUTC mismatch: got=%v want=%v", got.LastActivityAtUTC, u.LastActivityAtUTC)
	}
	got.BanExpiry, got.LastActivityAtUTC = u.BanExpiry, u.LastActivityAtUTC
	if *got != *u {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, u)
	}
}

func TestUserRoundTripNilOptionalFields(t *testing.T) {
	u := &User{ID: "u1", Email: "e@example.com"}
	got := userFromProto(userToProto(u))
	if got.BanExpiry != nil {
		t.Errorf("expected nil BanExpiry, got %v", got.BanExpiry)
	}
	if got.LastActivityAtUTC != nil {
		t.Errorf("expected nil LastActivityAtUTC, got %v", got.LastActivityAtUTC)
	}
	if got.AtlassianID != "" || got.HashedPassword != "" || got.AuthProviderID != "" {
		t.Errorf("expected empty NullString/NullUUID fields, got %+v", got)
	}
}

func TestUserToProtoFromProtoNil(t *testing.T) {
	if userToProto(nil) != nil {
		t.Error("userToProto(nil) should be nil")
	}
	if userFromProto(nil) != nil {
		t.Error("userFromProto(nil) should be nil")
	}
}

func TestOrgRoundTrip(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)
	o := &Organization{
		ID:               "o1",
		Name:             "Acme",
		ContactFirstname: "A",
		ContactLastname:  "B",
		ContactEmail:     "a@example.com",
		Language:         "en",
		SignupDate:       now,
	}
	got := orgFromProto(orgToProto(o))
	if !got.SignupDate.Equal(o.SignupDate) {
		t.Errorf("SignupDate mismatch: got=%v want=%v", got.SignupDate, o.SignupDate)
	}
	got.SignupDate = o.SignupDate
	if *got != *o {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, o)
	}
}

func TestDomainRoundTripNilAccessCheck(t *testing.T) {
	d := &Domain{DomainName: "example.com", OrganizationID: "o1", Active: true}
	got := domainFromProto(domainToProto(d))
	if got.AccessCheck != nil {
		t.Errorf("expected nil AccessCheck, got %v", got.AccessCheck)
	}
	got.AccessCheck = nil
	if *got != *d {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, d)
	}
}

func TestGroupRoundTrip(t *testing.T) {
	g := &Group{ID: "g1", OrganizationID: "o1", Name: "Engineering"}
	got := groupFromProto(groupToProto(g))
	if *got != *g {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, g)
	}
}

func TestLocationRoundTrip(t *testing.T) {
	l := &Location{
		ID: "l1", OrganizationID: "o1", Name: "HQ",
		MapWidth: 100, MapHeight: 200, MapScale: 1.5,
		MapMimeType: "image/png", MapType: "svg",
		Description: "desc", MaxConcurrentBookings: 5,
		Timezone: "Europe/Berlin", Enabled: true,
	}
	got := locationFromProto(locationToProto(l))
	if *got != *l {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, l)
	}
}

func TestSpaceRoundTrip(t *testing.T) {
	s := &Space{
		ID: "s1", LocationID: "l1", Name: "Desk 1",
		X: 1, Y: 2, Width: 3, Height: 4, Rotation: 90,
		RequireSubject: true, Enabled: true, KioskEnabled: true,
		Shape: "rect", FontSize: "12",
	}
	got := spaceFromProto(spaceToProto(s))
	if *got != *s {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, s)
	}
}

func TestBookingDetailsRoundTrip(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)
	bd := &BookingDetails{
		Space: SpaceDetails{
			Location: Location{ID: "l1", Name: "HQ"},
			Space:    Space{ID: "s1", Name: "Desk 1"},
		},
		UserEmail:     "u@example.com",
		UserFirstname: "First",
		UserLastname:  "Last",
		Booking: Booking{
			ID: "b1", UserID: "u1", SpaceID: "s1",
			Enter: now, Leave: now.Add(time.Hour),
			CalDavID: "cal1", Approved: true, Subject: "Meeting",
			RecurringID:  "rec1",
			CreatedAtUTC: &now,
		},
	}
	got := bookingDetailsFromProto(bookingDetailsToProto(bd))
	if got.UserEmail != bd.UserEmail || got.Space.Location.ID != bd.Space.Location.ID || got.Booking.ID != bd.Booking.ID {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, bd)
	}
	if got.Booking.LastInfoMailSentAtUTC != nil || got.Booking.ReminderSentAtUTC != nil {
		t.Errorf("expected nil optional timestamps, got %+v", got.Booking)
	}
	if !got.Booking.Enter.Equal(bd.Booking.Enter) || !got.Booking.Leave.Equal(bd.Booking.Leave) {
		t.Errorf("Enter/Leave mismatch: got=%+v want=%+v", got.Booking, bd.Booking)
	}
}

func TestAuthProviderRoundTrip(t *testing.T) {
	a := &AuthProvider{
		ID: "a1", OrganizationID: "o1", Name: "OAuth",
		ProviderType: int(OAuth2), AuthURL: "https://auth", TokenURL: "https://token",
		AuthStyle: 1, Scopes: "openid", UserInfoURL: "https://userinfo",
		UserInfoEmailField: "email", UserInfoFirstnameField: "first",
		UserInfoLastnameField: "last", ClientID: "cid", ClientSecret: "secret",
		LogoutURL: "https://logout", ProfilePageURL: "https://profile", ReadOnly: true,
	}
	got := authProviderFromProto(authProviderToProto(a))
	if *got != *a {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, a)
	}
}

func TestAuthStateRoundTrip(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)
	a := &AuthState{ID: "a1", AuthProviderID: "p1", Expiry: now, AuthStateType: AuthRequestState, Payload: "payload"}
	got := authStateFromProto(authStateToProto(a))
	if !got.Expiry.Equal(a.Expiry) {
		t.Errorf("Expiry mismatch: got=%v want=%v", got.Expiry, a.Expiry)
	}
	got.Expiry = a.Expiry
	if *got != *a {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, a)
	}
}
