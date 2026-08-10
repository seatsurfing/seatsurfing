package api

import (
	"time"

	"github.com/seatsurfing/seatsurfing/server/api/commonpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds pure conversion functions between the hand-written domain
// types in entities.go and their generated protobuf wire counterparts in
// commonpb. Kept isolated from the RPC plumbing (plugin_grpc_client.go,
// hostapi_grpc_server.go, etc.) so the mapping is easy to review and unit
// test on its own.

func timeToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func timeFromProto(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}

func nullTimeToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func nullTimeFromProto(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	v := t.AsTime()
	return &v
}

func userToProto(u *User) *commonpb.User {
	if u == nil {
		return nil
	}
	return &commonpb.User{
		Id:                     u.ID,
		OrganizationId:         u.OrganizationID,
		Email:                  u.Email,
		Firstname:              u.Firstname,
		Lastname:               u.Lastname,
		AtlassianId:            string(u.AtlassianID),
		HashedPassword:         string(u.HashedPassword),
		AuthProviderId:         string(u.AuthProviderID),
		PasswordPending:        u.PasswordPending,
		PasswordUpdateRequired: u.PasswordUpdateRequired,
		Role:                   int32(u.Role),
		Disabled:               u.Disabled,
		BanExpiry:              nullTimeToProto(u.BanExpiry),
		LastActivityAtUtc:      nullTimeToProto(u.LastActivityAtUTC),
		TotpSecret:             string(u.TotpSecret),
		ApiToken:               string(u.ApiToken),
	}
}

func userFromProto(p *commonpb.User) *User {
	if p == nil {
		return nil
	}
	return &User{
		ID:                     p.Id,
		OrganizationID:         p.OrganizationId,
		Email:                  p.Email,
		Firstname:              p.Firstname,
		Lastname:               p.Lastname,
		AtlassianID:            NullString(p.AtlassianId),
		HashedPassword:         NullString(p.HashedPassword),
		AuthProviderID:         NullUUID(p.AuthProviderId),
		PasswordPending:        p.PasswordPending,
		PasswordUpdateRequired: p.PasswordUpdateRequired,
		Role:                   UserRole(p.Role),
		Disabled:               p.Disabled,
		BanExpiry:              nullTimeFromProto(p.BanExpiry),
		LastActivityAtUTC:      nullTimeFromProto(p.LastActivityAtUtc),
		TotpSecret:             NullString(p.TotpSecret),
		ApiToken:               NullString(p.ApiToken),
	}
}

func usersToProto(users []*User) []*commonpb.User {
	out := make([]*commonpb.User, 0, len(users))
	for _, u := range users {
		out = append(out, userToProto(u))
	}
	return out
}

func usersFromProto(users []*commonpb.User) []*User {
	out := make([]*User, 0, len(users))
	for _, u := range users {
		out = append(out, userFromProto(u))
	}
	return out
}

func orgToProto(o *Organization) *commonpb.Organization {
	if o == nil {
		return nil
	}
	return &commonpb.Organization{
		Id:               o.ID,
		Name:             o.Name,
		ContactFirstname: o.ContactFirstname,
		ContactLastname:  o.ContactLastname,
		ContactEmail:     o.ContactEmail,
		Language:         o.Language,
		SignupDate:       timeToProto(o.SignupDate),
	}
}

func orgFromProto(p *commonpb.Organization) *Organization {
	if p == nil {
		return nil
	}
	return &Organization{
		ID:               p.Id,
		Name:             p.Name,
		ContactFirstname: p.ContactFirstname,
		ContactLastname:  p.ContactLastname,
		ContactEmail:     p.ContactEmail,
		Language:         p.Language,
		SignupDate:       timeFromProto(p.SignupDate),
	}
}

func orgsToProto(orgs []*Organization) []*commonpb.Organization {
	out := make([]*commonpb.Organization, 0, len(orgs))
	for _, o := range orgs {
		out = append(out, orgToProto(o))
	}
	return out
}

func orgsFromProto(orgs []*commonpb.Organization) []*Organization {
	out := make([]*Organization, 0, len(orgs))
	for _, o := range orgs {
		out = append(out, orgFromProto(o))
	}
	return out
}

func domainToProto(d *Domain) *commonpb.Domain {
	if d == nil {
		return nil
	}
	return &commonpb.Domain{
		DomainName:     d.DomainName,
		OrganizationId: d.OrganizationID,
		Active:         d.Active,
		VerifyToken:    d.VerifyToken,
		Primary:        d.Primary,
		Accessible:     d.Accessible,
		AccessCheck:    nullTimeToProto(d.AccessCheck),
	}
}

func domainFromProto(p *commonpb.Domain) *Domain {
	if p == nil {
		return nil
	}
	return &Domain{
		DomainName:     p.DomainName,
		OrganizationID: p.OrganizationId,
		Active:         p.Active,
		VerifyToken:    p.VerifyToken,
		Primary:        p.Primary,
		Accessible:     p.Accessible,
		AccessCheck:    nullTimeFromProto(p.AccessCheck),
	}
}

func groupToProto(g *Group) *commonpb.Group {
	if g == nil {
		return nil
	}
	return &commonpb.Group{
		Id:             g.ID,
		OrganizationId: g.OrganizationID,
		Name:           g.Name,
	}
}

func groupFromProto(p *commonpb.Group) *Group {
	if p == nil {
		return nil
	}
	return &Group{
		ID:             p.Id,
		OrganizationID: p.OrganizationId,
		Name:           p.Name,
	}
}

func groupsToProto(groups []*Group) []*commonpb.Group {
	out := make([]*commonpb.Group, 0, len(groups))
	for _, g := range groups {
		out = append(out, groupToProto(g))
	}
	return out
}

func groupsFromProto(groups []*commonpb.Group) []*Group {
	out := make([]*Group, 0, len(groups))
	for _, g := range groups {
		out = append(out, groupFromProto(g))
	}
	return out
}

func locationToProto(l *Location) *commonpb.Location {
	if l == nil {
		return nil
	}
	return &commonpb.Location{
		Id:                    l.ID,
		OrganizationId:        l.OrganizationID,
		Name:                  l.Name,
		MapWidth:              uint32(l.MapWidth),
		MapHeight:             uint32(l.MapHeight),
		MapScale:              l.MapScale,
		MapMimeType:           l.MapMimeType,
		MapType:               l.MapType,
		Description:           l.Description,
		MaxConcurrentBookings: uint32(l.MaxConcurrentBookings),
		Timezone:              l.Timezone,
		Enabled:               l.Enabled,
	}
}

func locationFromProto(p *commonpb.Location) *Location {
	if p == nil {
		return nil
	}
	return &Location{
		ID:                    p.Id,
		OrganizationID:        p.OrganizationId,
		Name:                  p.Name,
		MapWidth:              uint(p.MapWidth),
		MapHeight:             uint(p.MapHeight),
		MapScale:              p.MapScale,
		MapMimeType:           p.MapMimeType,
		MapType:               p.MapType,
		Description:           p.Description,
		MaxConcurrentBookings: uint(p.MaxConcurrentBookings),
		Timezone:              p.Timezone,
		Enabled:               p.Enabled,
	}
}

func spaceToProto(s *Space) *commonpb.Space {
	if s == nil {
		return nil
	}
	return &commonpb.Space{
		Id:             s.ID,
		LocationId:     s.LocationID,
		Name:           s.Name,
		X:              uint32(s.X),
		Y:              uint32(s.Y),
		Width:          uint32(s.Width),
		Height:         uint32(s.Height),
		Rotation:       uint32(s.Rotation),
		RequireSubject: s.RequireSubject,
		Enabled:        s.Enabled,
		KioskEnabled:   s.KioskEnabled,
		Shape:          s.Shape,
		FontSize:       s.FontSize,
	}
}

func spaceFromProto(p *commonpb.Space) *Space {
	if p == nil {
		return nil
	}
	return &Space{
		ID:             p.Id,
		LocationID:     p.LocationId,
		Name:           p.Name,
		X:              uint(p.X),
		Y:              uint(p.Y),
		Width:          uint(p.Width),
		Height:         uint(p.Height),
		Rotation:       uint(p.Rotation),
		RequireSubject: p.RequireSubject,
		Enabled:        p.Enabled,
		KioskEnabled:   p.KioskEnabled,
		Shape:          p.Shape,
		FontSize:       p.FontSize,
	}
}

func spaceDetailsToProto(s *SpaceDetails) *commonpb.SpaceDetails {
	if s == nil {
		return nil
	}
	return &commonpb.SpaceDetails{
		Location: locationToProto(&s.Location),
		Space:    spaceToProto(&s.Space),
	}
}

func spaceDetailsFromProto(p *commonpb.SpaceDetails) *SpaceDetails {
	if p == nil {
		return nil
	}
	sd := &SpaceDetails{}
	if l := locationFromProto(p.Location); l != nil {
		sd.Location = *l
	}
	if s := spaceFromProto(p.Space); s != nil {
		sd.Space = *s
	}
	return sd
}

func bookingToProto(b *Booking) *commonpb.Booking {
	if b == nil {
		return nil
	}
	return &commonpb.Booking{
		Id:                    b.ID,
		UserId:                b.UserID,
		SpaceId:               b.SpaceID,
		Enter:                 timeToProto(b.Enter),
		Leave:                 timeToProto(b.Leave),
		CalDavId:              b.CalDavID,
		Approved:              b.Approved,
		Subject:               b.Subject,
		RecurringId:           string(b.RecurringID),
		CreatedAtUtc:          nullTimeToProto(b.CreatedAtUTC),
		LastInfoMailSentAtUtc: nullTimeToProto(b.LastInfoMailSentAtUTC),
		ReminderSentAtUtc:     nullTimeToProto(b.ReminderSentAtUTC),
	}
}

func bookingFromProto(p *commonpb.Booking) *Booking {
	if p == nil {
		return nil
	}
	return &Booking{
		ID:                    p.Id,
		UserID:                p.UserId,
		SpaceID:               p.SpaceId,
		Enter:                 timeFromProto(p.Enter),
		Leave:                 timeFromProto(p.Leave),
		CalDavID:              p.CalDavId,
		Approved:              p.Approved,
		Subject:               p.Subject,
		RecurringID:           NullUUID(p.RecurringId),
		CreatedAtUTC:          nullTimeFromProto(p.CreatedAtUtc),
		LastInfoMailSentAtUTC: nullTimeFromProto(p.LastInfoMailSentAtUtc),
		ReminderSentAtUTC:     nullTimeFromProto(p.ReminderSentAtUtc),
	}
}

func bookingDetailsToProto(b *BookingDetails) *commonpb.BookingDetails {
	if b == nil {
		return nil
	}
	return &commonpb.BookingDetails{
		Space:         spaceDetailsToProto(&b.Space),
		UserEmail:     b.UserEmail,
		UserFirstname: b.UserFirstname,
		UserLastname:  b.UserLastname,
		Booking:       bookingToProto(&b.Booking),
	}
}

func bookingDetailsFromProto(p *commonpb.BookingDetails) *BookingDetails {
	if p == nil {
		return nil
	}
	bd := &BookingDetails{
		UserEmail:     p.UserEmail,
		UserFirstname: p.UserFirstname,
		UserLastname:  p.UserLastname,
	}
	if sd := spaceDetailsFromProto(p.Space); sd != nil {
		bd.Space = *sd
	}
	if b := bookingFromProto(p.Booking); b != nil {
		bd.Booking = *b
	}
	return bd
}

func authProviderToProto(a *AuthProvider) *commonpb.AuthProvider {
	if a == nil {
		return nil
	}
	return &commonpb.AuthProvider{
		Id:                     a.ID,
		OrganizationId:         a.OrganizationID,
		Name:                   a.Name,
		ProviderType:           int32(a.ProviderType),
		AuthUrl:                a.AuthURL,
		TokenUrl:               a.TokenURL,
		AuthStyle:              int32(a.AuthStyle),
		Scopes:                 a.Scopes,
		UserInfoUrl:            a.UserInfoURL,
		UserInfoEmailField:     a.UserInfoEmailField,
		UserInfoFirstnameField: a.UserInfoFirstnameField,
		UserInfoLastnameField:  a.UserInfoLastnameField,
		ClientId:               a.ClientID,
		ClientSecret:           a.ClientSecret,
		LogoutUrl:              a.LogoutURL,
		ProfilePageUrl:         a.ProfilePageURL,
		ReadOnly:               a.ReadOnly,
	}
}

func authProviderFromProto(p *commonpb.AuthProvider) *AuthProvider {
	if p == nil {
		return nil
	}
	return &AuthProvider{
		ID:                     p.Id,
		OrganizationID:         p.OrganizationId,
		Name:                   p.Name,
		ProviderType:           int(p.ProviderType),
		AuthURL:                p.AuthUrl,
		TokenURL:               p.TokenUrl,
		AuthStyle:              int(p.AuthStyle),
		Scopes:                 p.Scopes,
		UserInfoURL:            p.UserInfoUrl,
		UserInfoEmailField:     p.UserInfoEmailField,
		UserInfoFirstnameField: p.UserInfoFirstnameField,
		UserInfoLastnameField:  p.UserInfoLastnameField,
		ClientID:               p.ClientId,
		ClientSecret:           p.ClientSecret,
		LogoutURL:              p.LogoutUrl,
		ProfilePageURL:         p.ProfilePageUrl,
		ReadOnly:               p.ReadOnly,
	}
}

func authStateToProto(a *AuthState) *commonpb.AuthState {
	if a == nil {
		return nil
	}
	return &commonpb.AuthState{
		Id:             a.ID,
		AuthProviderId: a.AuthProviderID,
		Expiry:         timeToProto(a.Expiry),
		AuthStateType:  int32(a.AuthStateType),
		Payload:        a.Payload,
	}
}

func authStateFromProto(p *commonpb.AuthState) *AuthState {
	if p == nil {
		return nil
	}
	return &AuthState{
		ID:             p.Id,
		AuthProviderID: p.AuthProviderId,
		Expiry:         timeFromProto(p.Expiry),
		AuthStateType:  AuthStateType(p.AuthStateType),
		Payload:        p.Payload,
	}
}
