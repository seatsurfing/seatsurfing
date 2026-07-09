package api

import (
	"context"

	"github.com/seatsurfing/seatsurfing/server/api/commonpb"
	"github.com/seatsurfing/seatsurfing/server/api/hostapipb"
)

// HostAPIGRPCServer runs in the HOST process and exposes the existing HostAPI
// interface over gRPC so that any number of plugin processes can dial in.
// It wraps the same HostAPI implementation the net/rpc HostAPIRPCServer used
// to wrap - hostAPIImpl in app.go needs no changes.
type HostAPIGRPCServer struct {
	hostapipb.UnimplementedHostAPIServiceServer
	impl HostAPI
}

func NewHostAPIGRPCServer(h HostAPI) *HostAPIGRPCServer {
	return &HostAPIGRPCServer{impl: h}
}

// ─── Settings ─────────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) SettingsGet(ctx context.Context, a *hostapipb.SettingsGetArgs) (*hostapipb.SettingsGetReply, error) {
	v, err := s.impl.GetSettingsRepository().Get(a.OrgId, a.Name)
	return &hostapipb.SettingsGetReply{Value: v, Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) SettingsGetBool(ctx context.Context, a *hostapipb.SettingsGetArgs) (*hostapipb.SettingsGetBoolReply, error) {
	v, err := s.impl.GetSettingsRepository().GetBool(a.OrgId, a.Name)
	return &hostapipb.SettingsGetBoolReply{Value: v, Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) SettingsGetInt(ctx context.Context, a *hostapipb.SettingsGetArgs) (*hostapipb.IntReply, error) {
	v, err := s.impl.GetSettingsRepository().GetInt(a.OrgId, a.Name)
	return &hostapipb.IntReply{V: int32(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) SettingsGetNullUUID(ctx context.Context, _ *commonpb.Empty) (*hostapipb.StringReply, error) {
	return &hostapipb.StringReply{V: s.impl.GetSettingsRepository().GetNullUUID()}, nil
}
func (s *HostAPIGRPCServer) SettingsGetOrgIDsByValue(ctx context.Context, a *hostapipb.SettingsGetOrgIDsArgs) (*hostapipb.StringSliceReply, error) {
	v, err := s.impl.GetSettingsRepository().GetOrgIDsByValue(a.Name, a.Value)
	return &hostapipb.StringSliceReply{V: v, Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) SettingsSet(ctx context.Context, a *hostapipb.SettingsSetArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetSettingsRepository().Set(a.OrgId, a.Name, a.Value)
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) SettingsDelete(ctx context.Context, a *hostapipb.SettingsDeleteArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetSettingsRepository().Delete(a.OrgId, a.Name)
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}

// ─── Users ────────────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) UserGetOne(ctx context.Context, a *hostapipb.UserGetOneArgs) (*hostapipb.UserGetOneReply, error) {
	v, err := s.impl.GetUserRepository().GetOne(a.Id)
	return &hostapipb.UserGetOneReply{User: userToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) UserGetAll(ctx context.Context, a *hostapipb.UserGetAllArgs) (*hostapipb.UserGetAllReply, error) {
	v, err := s.impl.GetUserRepository().GetAll(a.OrgId, int(a.MaxResults), int(a.Offset))
	return &hostapipb.UserGetAllReply{Users: usersToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) UserGetByEmail(ctx context.Context, a *hostapipb.UserGetByEmailArgs) (*hostapipb.UserGetOneReply, error) {
	v, err := s.impl.GetUserRepository().GetByEmail(a.OrgId, a.Email)
	return &hostapipb.UserGetOneReply{User: userToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) UserGetCount(ctx context.Context, a *hostapipb.UserGetCountArgs) (*hostapipb.IntReply, error) {
	v, err := s.impl.GetUserRepository().GetCount(a.OrgId)
	return &hostapipb.IntReply{V: int32(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) UserGetHashedPassword(ctx context.Context, a *hostapipb.UserHashPasswordArgs) (*hostapipb.StringReply, error) {
	return &hostapipb.StringReply{V: s.impl.GetUserRepository().GetHashedPassword(a.Password)}, nil
}
func (s *HostAPIGRPCServer) UserGetUsersWithEmail(ctx context.Context, a *hostapipb.UserGetUsersWithEmailArgs) (*hostapipb.UserGetUsersWithEmailReply, error) {
	v, err := s.impl.GetUserRepository().GetUsersWithEmail(a.Email)
	return &hostapipb.UserGetUsersWithEmailReply{Users: usersToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) UserIsOrgAdmin(ctx context.Context, a *hostapipb.UserIsAdminArgs) (*hostapipb.BoolReply, error) {
	return &hostapipb.BoolReply{V: s.impl.GetUserRepository().IsOrgAdmin(userFromProto(a.User))}, nil
}
func (s *HostAPIGRPCServer) UserIsSuperAdmin(ctx context.Context, a *hostapipb.UserIsAdminArgs) (*hostapipb.BoolReply, error) {
	return &hostapipb.BoolReply{V: s.impl.GetUserRepository().IsSuperAdmin(userFromProto(a.User))}, nil
}
func (s *HostAPIGRPCServer) UserCreate(ctx context.Context, a *hostapipb.UserMutateArgs) (*hostapipb.UserGetOneReply, error) {
	u := userFromProto(a.User)
	err := s.impl.GetUserRepository().Create(u)
	return &hostapipb.UserGetOneReply{User: userToProto(u), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) UserUpdate(ctx context.Context, a *hostapipb.UserMutateArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetUserRepository().Update(userFromProto(a.User))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) UserDelete(ctx context.Context, a *hostapipb.UserMutateArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetUserRepository().Delete(userFromProto(a.User))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}

// ─── Organizations ────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) OrgGetOne(ctx context.Context, a *hostapipb.OrgGetOneArgs) (*hostapipb.OrgGetOneReply, error) {
	v, err := s.impl.GetOrganizationRepository().GetOne(a.Id)
	return &hostapipb.OrgGetOneReply{Org: orgToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgGetAll(ctx context.Context, _ *commonpb.Empty) (*hostapipb.OrgGetAllReply, error) {
	v, err := s.impl.GetOrganizationRepository().GetAll()
	return &hostapipb.OrgGetAllReply{Orgs: orgsToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgGetOneByDomain(ctx context.Context, a *hostapipb.OrgGetByDomainArgs) (*hostapipb.OrgGetOneReply, error) {
	v, err := s.impl.GetOrganizationRepository().GetOneByDomain(a.Domain)
	return &hostapipb.OrgGetOneReply{Org: orgToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgGetByEmail(ctx context.Context, a *hostapipb.OrgGetByEmailArgs) (*hostapipb.OrgGetOneReply, error) {
	v, err := s.impl.GetOrganizationRepository().GetByEmail(a.Email)
	return &hostapipb.OrgGetOneReply{Org: orgToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgGetAllDaysPassedSinceSignup(ctx context.Context, a *hostapipb.OrgGetDaysPassedArgs) (*hostapipb.OrgGetAllReply, error) {
	v, err := s.impl.GetOrganizationRepository().GetAllDaysPassedSinceSignup(int(a.DaysPassed), a.SettingExists)
	return &hostapipb.OrgGetAllReply{Orgs: orgsToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgGetPrimaryDomain(ctx context.Context, a *hostapipb.OrgGetPrimaryDomainArgs) (*hostapipb.OrgGetPrimaryDomainReply, error) {
	v, err := s.impl.GetOrganizationRepository().GetPrimaryDomain(orgFromProto(a.Org))
	return &hostapipb.OrgGetPrimaryDomainReply{Domain: domainToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgCreate(ctx context.Context, a *hostapipb.OrgMutateArgs) (*hostapipb.OrgGetOneReply, error) {
	o := orgFromProto(a.Org)
	err := s.impl.GetOrganizationRepository().Create(o)
	return &hostapipb.OrgGetOneReply{Org: orgToProto(o), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgUpdate(ctx context.Context, a *hostapipb.OrgMutateArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetOrganizationRepository().Update(orgFromProto(a.Org))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgDelete(ctx context.Context, a *hostapipb.OrgMutateArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetOrganizationRepository().Delete(orgFromProto(a.Org))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgAddDomain(ctx context.Context, a *hostapipb.OrgAddDomainArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetOrganizationRepository().AddDomain(orgFromProto(a.Org), a.Domain, a.Active)
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgSetPrimaryDomain(ctx context.Context, a *hostapipb.OrgSetPrimaryDomainArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetOrganizationRepository().SetPrimaryDomain(orgFromProto(a.Org), a.Domain)
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) OrgCreateSampleData(ctx context.Context, a *hostapipb.OrgCreateSampleDataArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetOrganizationRepository().CreateSampleData(orgFromProto(a.Org))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}

// ─── Groups ───────────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) GroupGetOne(ctx context.Context, a *hostapipb.GroupGetOneArgs) (*hostapipb.GroupGetOneReply, error) {
	v, err := s.impl.GetGroupRepository().GetOne(a.Id)
	return &hostapipb.GroupGetOneReply{Group: groupToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) GroupGetAll(ctx context.Context, a *hostapipb.GroupGetAllArgs) (*hostapipb.GroupGetAllReply, error) {
	v, err := s.impl.GetGroupRepository().GetAll(a.OrgId)
	return &hostapipb.GroupGetAllReply{Groups: groupsToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) GroupGetByName(ctx context.Context, a *hostapipb.GroupGetByNameArgs) (*hostapipb.GroupGetOneReply, error) {
	v, err := s.impl.GetGroupRepository().GetByName(a.OrgId, a.Name)
	return &hostapipb.GroupGetOneReply{Group: groupToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) GroupGetMemberUserIDs(ctx context.Context, a *hostapipb.GroupGetMemberIDsArgs) (*hostapipb.GroupGetMemberIDsReply, error) {
	v, err := s.impl.GetGroupRepository().GetMemberUserIDs(groupFromProto(a.Group))
	return &hostapipb.GroupGetMemberIDsReply{Ids: v, Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) GroupAddMembers(ctx context.Context, a *hostapipb.GroupMembersArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetGroupRepository().AddMembers(groupFromProto(a.Group), a.UserIds)
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) GroupRemoveMembers(ctx context.Context, a *hostapipb.GroupMembersArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetGroupRepository().RemoveMembers(groupFromProto(a.Group), a.UserIds)
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) GroupCreate(ctx context.Context, a *hostapipb.GroupMutateArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetGroupRepository().Create(groupFromProto(a.Group))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) GroupUpdate(ctx context.Context, a *hostapipb.GroupMutateArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetGroupRepository().Update(groupFromProto(a.Group))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) GroupDelete(ctx context.Context, a *hostapipb.GroupMutateArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetGroupRepository().Delete(groupFromProto(a.Group))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}

// ─── Bookings ─────────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) BookingGetOne(ctx context.Context, a *hostapipb.BookingGetOneArgs) (*hostapipb.BookingGetOneReply, error) {
	v, err := s.impl.GetBookingRepository().GetOne(a.Id)
	return &hostapipb.BookingGetOneReply{Booking: bookingDetailsToProto(v), Err: errStr(err)}, nil
}

// ─── Spaces ───────────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) SpaceGetOne(ctx context.Context, a *hostapipb.SpaceGetOneArgs) (*hostapipb.SpaceGetOneReply, error) {
	v, err := s.impl.GetSpaceRepository().GetOne(a.Id)
	return &hostapipb.SpaceGetOneReply{Space: spaceToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) SpaceGetCount(ctx context.Context, a *hostapipb.SpaceGetCountArgs) (*hostapipb.IntReply, error) {
	v, err := s.impl.GetSpaceRepository().GetCount(a.OrgId)
	return &hostapipb.IntReply{V: int32(v), Err: errStr(err)}, nil
}

// ─── Locations ────────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) LocationGetOne(ctx context.Context, a *hostapipb.LocationGetOneArgs) (*hostapipb.LocationGetOneReply, error) {
	v, err := s.impl.GetLocationRepository().GetOne(a.Id)
	return &hostapipb.LocationGetOneReply{Location: locationToProto(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) LocationGetCount(ctx context.Context, a *hostapipb.LocationGetCountArgs) (*hostapipb.IntReply, error) {
	v, err := s.impl.GetLocationRepository().GetCount(a.OrgId)
	return &hostapipb.IntReply{V: int32(v), Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) LocationGetTimezone(ctx context.Context, a *hostapipb.LocationGetTimezoneArgs) (*hostapipb.StringReply, error) {
	return &hostapipb.StringReply{V: s.impl.GetLocationRepository().GetTimezone(locationFromProto(a.Location))}, nil
}

// ─── AuthProvider ─────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) AuthProviderCreate(ctx context.Context, a *hostapipb.AuthProviderMutateArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetAuthProviderRepository().Create(authProviderFromProto(a.AuthProvider))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) AuthProviderUpdate(ctx context.Context, a *hostapipb.AuthProviderMutateArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.GetAuthProviderRepository().Update(authProviderFromProto(a.AuthProvider))
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}

// ─── AuthState ────────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) AuthStateCreate(ctx context.Context, a *hostapipb.AuthStateMutateArgs) (*hostapipb.AuthStateGetOneReply, error) {
	as := authStateFromProto(a.AuthState)
	err := s.impl.GetAuthStateRepository().Create(as)
	return &hostapipb.AuthStateGetOneReply{AuthState: authStateToProto(as), Err: errStr(err)}, nil
}

// ─── General ──────────────────────────────────────────────────────────────────

func (s *HostAPIGRPCServer) SendEmail(ctx context.Context, a *hostapipb.SendEmailArgs) (*hostapipb.ErrorReply, error) {
	err := s.impl.SendEmail(a.Recipient, a.Subject, a.Body, a.Language, a.OrgId)
	return &hostapipb.ErrorReply{Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) Encrypt(ctx context.Context, a *hostapipb.EncryptArgs) (*hostapipb.EncryptReply, error) {
	v, err := s.impl.Encrypt(a.Plaintext)
	return &hostapipb.EncryptReply{Result: v, Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) Decrypt(ctx context.Context, a *hostapipb.DecryptArgs) (*hostapipb.EncryptReply, error) {
	v, err := s.impl.Decrypt(a.Ciphertext)
	return &hostapipb.EncryptReply{Result: v, Err: errStr(err)}, nil
}
func (s *HostAPIGRPCServer) IsValidLanguageCode(ctx context.Context, a *hostapipb.IsValidLangArgs) (*hostapipb.BoolReply, error) {
	return &hostapipb.BoolReply{V: s.impl.IsValidLanguageCode(a.Code)}, nil
}
func (s *HostAPIGRPCServer) DisablePasswordLogin(ctx context.Context, _ *commonpb.Empty) (*hostapipb.BoolReply, error) {
	return &hostapipb.BoolReply{V: s.impl.DisablePasswordLogin()}, nil
}
func (s *HostAPIGRPCServer) FormatPublicURL(ctx context.Context, a *hostapipb.FormatPublicURLArgs) (*hostapipb.StringReply, error) {
	return &hostapipb.StringReply{V: s.impl.FormatPublicURL(a.Domain)}, nil
}
func (s *HostAPIGRPCServer) IsDevelopmentMode(ctx context.Context, _ *commonpb.Empty) (*hostapipb.BoolReply, error) {
	return &hostapipb.BoolReply{V: s.impl.IsDevelopmentMode()}, nil
}
func (s *HostAPIGRPCServer) GetPostgresURL(ctx context.Context, _ *commonpb.Empty) (*hostapipb.StringReply, error) {
	return &hostapipb.StringReply{V: s.impl.GetPostgresURL()}, nil
}
