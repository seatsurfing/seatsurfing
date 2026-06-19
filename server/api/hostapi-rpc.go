package api

import (
	"errors"
	"net/rpc"
)

// ─── Args / Reply structs ────────────────────────────────────────────────────

type BoolReply struct{ V bool }
type IntReply struct{ V int }
type StringReply struct{ V string }
type StringSliceReply struct {
	V   []string
	Err string
}
type ErrorReply struct{ Err string }

// Settings
type SettingsGetArgs struct{ OrgID, Name string }
type SettingsGetReply struct {
	Value string
	Err   string
}
type SettingsGetBoolReply struct {
	Value bool
	Err   string
}
type SettingsGetOrgIDsArgs struct{ Name, Value string }
type SettingsSetArgs struct{ OrgID, Name, Value string }
type SettingsDeleteArgs struct{ OrgID, Name string }

// Users
type UserGetOneArgs struct{ ID string }
type UserGetOneReply struct {
	User *User
	Err  string
}
type UserGetAllArgs struct {
	OrgID      string
	MaxResults int
	Offset     int
}
type UserGetAllReply struct {
	Users []*User
	Err   string
}
type UserGetByEmailArgs struct{ OrgID, Email string }
type UserGetUsersWithEmailArgs struct{ Email string }
type UserGetUsersWithEmailReply struct {
	Users []*User
	Err   string
}
type UserGetCountArgs struct{ OrgID string }
type UserHashPasswordArgs struct{ Password string }
type UserIsAdminArgs struct{ User *User }
type UserMutateArgs struct{ User *User }

// Organizations
type OrgGetOneArgs struct{ ID string }
type OrgGetOneReply struct {
	Org *Organization
	Err string
}
type OrgGetAllReply struct {
	Orgs []*Organization
	Err  string
}
type OrgGetByDomainArgs struct{ Domain string }
type OrgGetByEmailArgs struct{ Email string }
type OrgGetDaysPassedArgs struct {
	DaysPassed    int
	SettingExists string
}
type OrgGetPrimaryDomainArgs struct{ Org *Organization }
type OrgGetPrimaryDomainReply struct {
	Domain *Domain
	Err    string
}
type OrgMutateArgs struct{ Org *Organization }
type OrgAddDomainArgs struct {
	Org    *Organization
	Domain string
	Active bool
}
type OrgSetPrimaryDomainArgs struct {
	Org    *Organization
	Domain string
}
type OrgCreateSampleDataArgs struct{ Org *Organization }

// Groups
type GroupGetOneArgs struct{ ID string }
type GroupGetOneReply struct {
	Group *Group
	Err   string
}
type GroupGetAllArgs struct{ OrgID string }
type GroupGetAllReply struct {
	Groups []*Group
	Err    string
}
type GroupGetByNameArgs struct{ OrgID, Name string }
type GroupGetMemberIDsArgs struct{ Group *Group }
type GroupGetMemberIDsReply struct {
	IDs []string
	Err string
}
type GroupMembersArgs struct {
	Group   *Group
	UserIDs []string
}
type GroupMutateArgs struct{ Group *Group }

// Bookings
type BookingGetOneArgs struct{ ID string }
type BookingGetOneReply struct {
	Booking *BookingDetails
	Err     string
}

// Spaces
type SpaceGetOneArgs struct{ ID string }
type SpaceGetOneReply struct {
	Space *Space
	Err   string
}
type SpaceGetCountArgs struct{ OrgID string }

// Locations
type LocationGetOneArgs struct{ ID string }
type LocationGetOneReply struct {
	Location *Location
	Err      string
}
type LocationGetCountArgs struct{ OrgID string }
type LocationGetTimezoneArgs struct{ Location *Location }

// AuthProvider
type AuthProviderMutateArgs struct{ AuthProvider *AuthProvider }

// AuthState
type AuthStateMutateArgs struct{ AuthState *AuthState }

// General
type SendEmailArgs struct{ Recipient, Subject, Body, Language, OrgID string }
type EncryptArgs struct{ Plaintext string }
type EncryptReply struct {
	Result string
	Err    string
}
type DecryptArgs struct{ Ciphertext string }
type IsValidLangArgs struct{ Code string }

// ─── Host-side RPC Server ────────────────────────────────────────────────────

// HostAPIRPCServer runs in the HOST process. It wraps real repositories and
// exposes them as net/rpc methods for the plugin subprocess.
// go-plugin's AcceptAndServe registers it under the "Plugin" name, so methods
// are called as "Plugin.SettingsGet" etc.
type HostAPIRPCServer struct {
	impl HostAPI
}

func NewHostAPIRPCServer(h HostAPI) *HostAPIRPCServer {
	return &HostAPIRPCServer{impl: h}
}

func errStr(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
func strErr(s string) error {
	if s != "" {
		return errors.New(s)
	}
	return nil
}

// Settings
func (s *HostAPIRPCServer) SettingsGet(a SettingsGetArgs, r *SettingsGetReply) error {
	v, err := s.impl.GetSettingsRepository().Get(a.OrgID, a.Name)
	r.Value, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) SettingsGetBool(a SettingsGetArgs, r *SettingsGetBoolReply) error {
	v, err := s.impl.GetSettingsRepository().GetBool(a.OrgID, a.Name)
	r.Value, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) SettingsGetInt(a SettingsGetArgs, r *IntReply) error {
	v, err := s.impl.GetSettingsRepository().GetInt(a.OrgID, a.Name)
	r.V = v
	if err != nil {
		return err
	}
	return nil
}
func (s *HostAPIRPCServer) SettingsGetNullUUID(_ struct{}, r *StringReply) error {
	r.V = s.impl.GetSettingsRepository().GetNullUUID()
	return nil
}
func (s *HostAPIRPCServer) SettingsGetOrgIDsByValue(a SettingsGetOrgIDsArgs, r *StringSliceReply) error {
	v, err := s.impl.GetSettingsRepository().GetOrgIDsByValue(a.Name, a.Value)
	r.V, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) SettingsSet(a SettingsSetArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetSettingsRepository().Set(a.OrgID, a.Name, a.Value))
	return nil
}
func (s *HostAPIRPCServer) SettingsDelete(a SettingsDeleteArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetSettingsRepository().Delete(a.OrgID, a.Name))
	return nil
}

// Users
func (s *HostAPIRPCServer) UserGetOne(a UserGetOneArgs, r *UserGetOneReply) error {
	v, err := s.impl.GetUserRepository().GetOne(a.ID)
	r.User, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) UserGetAll(a UserGetAllArgs, r *UserGetAllReply) error {
	v, err := s.impl.GetUserRepository().GetAll(a.OrgID, a.MaxResults, a.Offset)
	r.Users, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) UserGetByEmail(a UserGetByEmailArgs, r *UserGetOneReply) error {
	v, err := s.impl.GetUserRepository().GetByEmail(a.OrgID, a.Email)
	r.User, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) UserGetCount(a UserGetCountArgs, r *IntReply) error {
	v, err := s.impl.GetUserRepository().GetCount(a.OrgID)
	r.V = v
	if err != nil {
		return err
	}
	return nil
}
func (s *HostAPIRPCServer) UserGetHashedPassword(a UserHashPasswordArgs, r *StringReply) error {
	r.V = s.impl.GetUserRepository().GetHashedPassword(a.Password)
	return nil
}
func (s *HostAPIRPCServer) UserGetUsersWithEmail(a UserGetUsersWithEmailArgs, r *UserGetUsersWithEmailReply) error {
	v, err := s.impl.GetUserRepository().GetUsersWithEmail(a.Email)
	r.Users, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) UserIsOrgAdmin(a UserIsAdminArgs, r *BoolReply) error {
	r.V = s.impl.GetUserRepository().IsOrgAdmin(a.User)
	return nil
}
func (s *HostAPIRPCServer) UserIsSuperAdmin(a UserIsAdminArgs, r *BoolReply) error {
	r.V = s.impl.GetUserRepository().IsSuperAdmin(a.User)
	return nil
}
func (s *HostAPIRPCServer) UserCreate(a UserMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetUserRepository().Create(a.User))
	return nil
}
func (s *HostAPIRPCServer) UserUpdate(a UserMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetUserRepository().Update(a.User))
	return nil
}
func (s *HostAPIRPCServer) UserDelete(a UserMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetUserRepository().Delete(a.User))
	return nil
}

// Organizations
func (s *HostAPIRPCServer) OrgGetOne(a OrgGetOneArgs, r *OrgGetOneReply) error {
	v, err := s.impl.GetOrganizationRepository().GetOne(a.ID)
	r.Org, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) OrgGetAll(_ struct{}, r *OrgGetAllReply) error {
	v, err := s.impl.GetOrganizationRepository().GetAll()
	r.Orgs, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) OrgGetOneByDomain(a OrgGetByDomainArgs, r *OrgGetOneReply) error {
	v, err := s.impl.GetOrganizationRepository().GetOneByDomain(a.Domain)
	r.Org, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) OrgGetByEmail(a OrgGetByEmailArgs, r *OrgGetOneReply) error {
	v, err := s.impl.GetOrganizationRepository().GetByEmail(a.Email)
	r.Org, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) OrgGetAllDaysPassedSinceSignup(a OrgGetDaysPassedArgs, r *OrgGetAllReply) error {
	v, err := s.impl.GetOrganizationRepository().GetAllDaysPassedSinceSignup(a.DaysPassed, a.SettingExists)
	r.Orgs, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) OrgGetPrimaryDomain(a OrgGetPrimaryDomainArgs, r *OrgGetPrimaryDomainReply) error {
	v, err := s.impl.GetOrganizationRepository().GetPrimaryDomain(a.Org)
	r.Domain, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) OrgCreate(a OrgMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetOrganizationRepository().Create(a.Org))
	return nil
}
func (s *HostAPIRPCServer) OrgUpdate(a OrgMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetOrganizationRepository().Update(a.Org))
	return nil
}
func (s *HostAPIRPCServer) OrgDelete(a OrgMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetOrganizationRepository().Delete(a.Org))
	return nil
}
func (s *HostAPIRPCServer) OrgAddDomain(a OrgAddDomainArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetOrganizationRepository().AddDomain(a.Org, a.Domain, a.Active))
	return nil
}
func (s *HostAPIRPCServer) OrgSetPrimaryDomain(a OrgSetPrimaryDomainArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetOrganizationRepository().SetPrimaryDomain(a.Org, a.Domain))
	return nil
}
func (s *HostAPIRPCServer) OrgCreateSampleData(a OrgCreateSampleDataArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetOrganizationRepository().CreateSampleData(a.Org))
	return nil
}

// Groups
func (s *HostAPIRPCServer) GroupGetOne(a GroupGetOneArgs, r *GroupGetOneReply) error {
	v, err := s.impl.GetGroupRepository().GetOne(a.ID)
	r.Group, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) GroupGetAll(a GroupGetAllArgs, r *GroupGetAllReply) error {
	v, err := s.impl.GetGroupRepository().GetAll(a.OrgID)
	r.Groups, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) GroupGetByName(a GroupGetByNameArgs, r *GroupGetOneReply) error {
	v, err := s.impl.GetGroupRepository().GetByName(a.OrgID, a.Name)
	r.Group, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) GroupGetMemberUserIDs(a GroupGetMemberIDsArgs, r *GroupGetMemberIDsReply) error {
	v, err := s.impl.GetGroupRepository().GetMemberUserIDs(a.Group)
	r.IDs, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) GroupAddMembers(a GroupMembersArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetGroupRepository().AddMembers(a.Group, a.UserIDs))
	return nil
}
func (s *HostAPIRPCServer) GroupRemoveMembers(a GroupMembersArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetGroupRepository().RemoveMembers(a.Group, a.UserIDs))
	return nil
}
func (s *HostAPIRPCServer) GroupCreate(a GroupMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetGroupRepository().Create(a.Group))
	return nil
}
func (s *HostAPIRPCServer) GroupUpdate(a GroupMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetGroupRepository().Update(a.Group))
	return nil
}
func (s *HostAPIRPCServer) GroupDelete(a GroupMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetGroupRepository().Delete(a.Group))
	return nil
}

// Bookings
func (s *HostAPIRPCServer) BookingGetOne(a BookingGetOneArgs, r *BookingGetOneReply) error {
	v, err := s.impl.GetBookingRepository().GetOne(a.ID)
	r.Booking, r.Err = v, errStr(err)
	return nil
}

// Spaces
func (s *HostAPIRPCServer) SpaceGetOne(a SpaceGetOneArgs, r *SpaceGetOneReply) error {
	v, err := s.impl.GetSpaceRepository().GetOne(a.ID)
	r.Space, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) SpaceGetCount(a SpaceGetCountArgs, r *IntReply) error {
	v, err := s.impl.GetSpaceRepository().GetCount(a.OrgID)
	r.V = v
	if err != nil {
		return err
	}
	return nil
}

// Locations
func (s *HostAPIRPCServer) LocationGetOne(a LocationGetOneArgs, r *LocationGetOneReply) error {
	v, err := s.impl.GetLocationRepository().GetOne(a.ID)
	r.Location, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) LocationGetCount(a LocationGetCountArgs, r *IntReply) error {
	v, err := s.impl.GetLocationRepository().GetCount(a.OrgID)
	r.V = v
	if err != nil {
		return err
	}
	return nil
}
func (s *HostAPIRPCServer) LocationGetTimezone(a LocationGetTimezoneArgs, r *StringReply) error {
	r.V = s.impl.GetLocationRepository().GetTimezone(a.Location)
	return nil
}

// AuthProvider
func (s *HostAPIRPCServer) AuthProviderCreate(a AuthProviderMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetAuthProviderRepository().Create(a.AuthProvider))
	return nil
}
func (s *HostAPIRPCServer) AuthProviderUpdate(a AuthProviderMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetAuthProviderRepository().Update(a.AuthProvider))
	return nil
}

// AuthState
func (s *HostAPIRPCServer) AuthStateCreate(a AuthStateMutateArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.GetAuthStateRepository().Create(a.AuthState))
	return nil
}

// General
func (s *HostAPIRPCServer) SendEmail(a SendEmailArgs, r *ErrorReply) error {
	r.Err = errStr(s.impl.SendEmail(a.Recipient, a.Subject, a.Body, a.Language, a.OrgID))
	return nil
}
func (s *HostAPIRPCServer) Encrypt(a EncryptArgs, r *EncryptReply) error {
	v, err := s.impl.Encrypt(a.Plaintext)
	r.Result, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) Decrypt(a DecryptArgs, r *EncryptReply) error {
	v, err := s.impl.Decrypt(a.Ciphertext)
	r.Result, r.Err = v, errStr(err)
	return nil
}
func (s *HostAPIRPCServer) IsValidLanguageCode(a IsValidLangArgs, r *BoolReply) error {
	r.V = s.impl.IsValidLanguageCode(a.Code)
	return nil
}
func (s *HostAPIRPCServer) DisablePasswordLogin(_ struct{}, r *BoolReply) error {
	r.V = s.impl.DisablePasswordLogin()
	return nil
}

// ─── Plugin-side RPC Client ──────────────────────────────────────────────────

// HostAPIRPC runs in the PLUGIN process. It implements HostAPI by forwarding
// every call to the host via net/rpc.
type HostAPIRPC struct {
	client *rpc.Client
}

func NewHostAPIRPC(client *rpc.Client) *HostAPIRPC {
	return &HostAPIRPC{client: client}
}

func (h *HostAPIRPC) GetSettingsRepository() SettingsRepository {
	return &settingsRepositoryRPC{client: h.client}
}
func (h *HostAPIRPC) GetUserRepository() UserRepository {
	return &userRepositoryRPC{client: h.client}
}
func (h *HostAPIRPC) GetOrganizationRepository() OrganizationRepository {
	return &organizationRepositoryRPC{client: h.client}
}
func (h *HostAPIRPC) GetGroupRepository() GroupRepository {
	return &groupRepositoryRPC{client: h.client}
}
func (h *HostAPIRPC) GetBookingRepository() BookingRepository {
	return &bookingRepositoryRPC{client: h.client}
}
func (h *HostAPIRPC) GetSpaceRepository() SpaceRepository {
	return &spaceRepositoryRPC{client: h.client}
}
func (h *HostAPIRPC) GetLocationRepository() LocationRepository {
	return &locationRepositoryRPC{client: h.client}
}
func (h *HostAPIRPC) GetAuthProviderRepository() AuthProviderRepository {
	return &authProviderRepositoryRPC{client: h.client}
}
func (h *HostAPIRPC) GetAuthStateRepository() AuthStateRepository {
	return &authStateRepositoryRPC{client: h.client}
}
func (h *HostAPIRPC) SendEmail(recipient, subject, body, language, orgID string) error {
	var r ErrorReply
	if err := h.client.Call("Plugin.SendEmail", SendEmailArgs{recipient, subject, body, language, orgID}, &r); err != nil {
		return err
	}
	return strErr(r.Err)
}
func (h *HostAPIRPC) Encrypt(plaintext string) (string, error) {
	var r EncryptReply
	if err := h.client.Call("Plugin.Encrypt", EncryptArgs{plaintext}, &r); err != nil {
		return "", err
	}
	return r.Result, strErr(r.Err)
}
func (h *HostAPIRPC) Decrypt(ciphertext string) (string, error) {
	var r EncryptReply
	if err := h.client.Call("Plugin.Decrypt", DecryptArgs{ciphertext}, &r); err != nil {
		return "", err
	}
	return r.Result, strErr(r.Err)
}
func (h *HostAPIRPC) IsValidLanguageCode(code string) bool {
	var r BoolReply
	_ = h.client.Call("Plugin.IsValidLanguageCode", IsValidLangArgs{code}, &r)
	return r.V
}
func (h *HostAPIRPC) DisablePasswordLogin() bool {
	var r BoolReply
	_ = h.client.Call("Plugin.DisablePasswordLogin", struct{}{}, &r)
	return r.V
}

// ─── Per-repository RPC proxies (plugin side) ────────────────────────────────

type settingsRepositoryRPC struct{ client *rpc.Client }

func (r *settingsRepositoryRPC) Get(orgID, name string) (string, error) {
	var reply SettingsGetReply
	if err := r.client.Call("Plugin.SettingsGet", SettingsGetArgs{orgID, name}, &reply); err != nil {
		return "", err
	}
	return reply.Value, strErr(reply.Err)
}
func (r *settingsRepositoryRPC) GetBool(orgID, name string) (bool, error) {
	var reply SettingsGetBoolReply
	if err := r.client.Call("Plugin.SettingsGetBool", SettingsGetArgs{orgID, name}, &reply); err != nil {
		return false, err
	}
	return reply.Value, strErr(reply.Err)
}
func (r *settingsRepositoryRPC) GetInt(orgID, name string) (int, error) {
	var reply IntReply
	if err := r.client.Call("Plugin.SettingsGetInt", SettingsGetArgs{orgID, name}, &reply); err != nil {
		return 0, err
	}
	return reply.V, nil
}
func (r *settingsRepositoryRPC) GetNullUUID() string {
	var reply StringReply
	_ = r.client.Call("Plugin.SettingsGetNullUUID", struct{}{}, &reply)
	return reply.V
}
func (r *settingsRepositoryRPC) GetOrgIDsByValue(name, value string) ([]string, error) {
	var reply StringSliceReply
	if err := r.client.Call("Plugin.SettingsGetOrgIDsByValue", SettingsGetOrgIDsArgs{name, value}, &reply); err != nil {
		return nil, err
	}
	return reply.V, strErr(reply.Err)
}
func (r *settingsRepositoryRPC) Set(orgID, name, value string) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.SettingsSet", SettingsSetArgs{orgID, name, value}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *settingsRepositoryRPC) Delete(orgID, name string) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.SettingsDelete", SettingsDeleteArgs{orgID, name}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}

type userRepositoryRPC struct{ client *rpc.Client }

func (r *userRepositoryRPC) GetOne(id string) (*User, error) {
	var reply UserGetOneReply
	if err := r.client.Call("Plugin.UserGetOne", UserGetOneArgs{id}, &reply); err != nil {
		return nil, err
	}
	return reply.User, strErr(reply.Err)
}
func (r *userRepositoryRPC) GetAll(orgID string, maxResults, offset int) ([]*User, error) {
	var reply UserGetAllReply
	if err := r.client.Call("Plugin.UserGetAll", UserGetAllArgs{orgID, maxResults, offset}, &reply); err != nil {
		return nil, err
	}
	return reply.Users, strErr(reply.Err)
}
func (r *userRepositoryRPC) GetByEmail(orgID, email string) (*User, error) {
	var reply UserGetOneReply
	if err := r.client.Call("Plugin.UserGetByEmail", UserGetByEmailArgs{orgID, email}, &reply); err != nil {
		return nil, err
	}
	return reply.User, strErr(reply.Err)
}
func (r *userRepositoryRPC) GetCount(orgID string) (int, error) {
	var reply IntReply
	if err := r.client.Call("Plugin.UserGetCount", UserGetCountArgs{orgID}, &reply); err != nil {
		return 0, err
	}
	return reply.V, nil
}
func (r *userRepositoryRPC) GetHashedPassword(password string) string {
	var reply StringReply
	_ = r.client.Call("Plugin.UserGetHashedPassword", UserHashPasswordArgs{password}, &reply)
	return reply.V
}
func (r *userRepositoryRPC) GetUsersWithEmail(email string) ([]*User, error) {
	var reply UserGetUsersWithEmailReply
	if err := r.client.Call("Plugin.UserGetUsersWithEmail", UserGetUsersWithEmailArgs{email}, &reply); err != nil {
		return nil, err
	}
	return reply.Users, strErr(reply.Err)
}
func (r *userRepositoryRPC) IsOrgAdmin(user *User) bool {
	var reply BoolReply
	_ = r.client.Call("Plugin.UserIsOrgAdmin", UserIsAdminArgs{user}, &reply)
	return reply.V
}
func (r *userRepositoryRPC) IsSuperAdmin(user *User) bool {
	var reply BoolReply
	_ = r.client.Call("Plugin.UserIsSuperAdmin", UserIsAdminArgs{user}, &reply)
	return reply.V
}
func (r *userRepositoryRPC) Create(e *User) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.UserCreate", UserMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *userRepositoryRPC) Update(e *User) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.UserUpdate", UserMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *userRepositoryRPC) Delete(e *User) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.UserDelete", UserMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}

type organizationRepositoryRPC struct{ client *rpc.Client }

func (r *organizationRepositoryRPC) GetOne(id string) (*Organization, error) {
	var reply OrgGetOneReply
	if err := r.client.Call("Plugin.OrgGetOne", OrgGetOneArgs{id}, &reply); err != nil {
		return nil, err
	}
	return reply.Org, strErr(reply.Err)
}
func (r *organizationRepositoryRPC) GetAll() ([]*Organization, error) {
	var reply OrgGetAllReply
	if err := r.client.Call("Plugin.OrgGetAll", struct{}{}, &reply); err != nil {
		return nil, err
	}
	return reply.Orgs, strErr(reply.Err)
}
func (r *organizationRepositoryRPC) GetOneByDomain(domain string) (*Organization, error) {
	var reply OrgGetOneReply
	if err := r.client.Call("Plugin.OrgGetOneByDomain", OrgGetByDomainArgs{domain}, &reply); err != nil {
		return nil, err
	}
	return reply.Org, strErr(reply.Err)
}
func (r *organizationRepositoryRPC) GetByEmail(email string) (*Organization, error) {
	var reply OrgGetOneReply
	if err := r.client.Call("Plugin.OrgGetByEmail", OrgGetByEmailArgs{email}, &reply); err != nil {
		return nil, err
	}
	return reply.Org, strErr(reply.Err)
}
func (r *organizationRepositoryRPC) GetAllDaysPassedSinceSignup(daysPassed int, settingExists string) ([]*Organization, error) {
	var reply OrgGetAllReply
	if err := r.client.Call("Plugin.OrgGetAllDaysPassedSinceSignup", OrgGetDaysPassedArgs{daysPassed, settingExists}, &reply); err != nil {
		return nil, err
	}
	return reply.Orgs, strErr(reply.Err)
}
func (r *organizationRepositoryRPC) GetPrimaryDomain(e *Organization) (*Domain, error) {
	var reply OrgGetPrimaryDomainReply
	if err := r.client.Call("Plugin.OrgGetPrimaryDomain", OrgGetPrimaryDomainArgs{e}, &reply); err != nil {
		return nil, err
	}
	return reply.Domain, strErr(reply.Err)
}
func (r *organizationRepositoryRPC) Create(e *Organization) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.OrgCreate", OrgMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryRPC) Update(e *Organization) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.OrgUpdate", OrgMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryRPC) Delete(e *Organization) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.OrgDelete", OrgMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryRPC) AddDomain(e *Organization, domain string, active bool) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.OrgAddDomain", OrgAddDomainArgs{e, domain, active}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryRPC) SetPrimaryDomain(e *Organization, domain string) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.OrgSetPrimaryDomain", OrgSetPrimaryDomainArgs{e, domain}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryRPC) CreateSampleData(org *Organization) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.OrgCreateSampleData", OrgCreateSampleDataArgs{org}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}

type groupRepositoryRPC struct{ client *rpc.Client }

func (r *groupRepositoryRPC) GetOne(id string) (*Group, error) {
	var reply GroupGetOneReply
	if err := r.client.Call("Plugin.GroupGetOne", GroupGetOneArgs{id}, &reply); err != nil {
		return nil, err
	}
	return reply.Group, strErr(reply.Err)
}
func (r *groupRepositoryRPC) GetAll(orgID string) ([]*Group, error) {
	var reply GroupGetAllReply
	if err := r.client.Call("Plugin.GroupGetAll", GroupGetAllArgs{orgID}, &reply); err != nil {
		return nil, err
	}
	return reply.Groups, strErr(reply.Err)
}
func (r *groupRepositoryRPC) GetByName(orgID, name string) (*Group, error) {
	var reply GroupGetOneReply
	if err := r.client.Call("Plugin.GroupGetByName", GroupGetByNameArgs{orgID, name}, &reply); err != nil {
		return nil, err
	}
	return reply.Group, strErr(reply.Err)
}
func (r *groupRepositoryRPC) GetMemberUserIDs(e *Group) ([]string, error) {
	var reply GroupGetMemberIDsReply
	if err := r.client.Call("Plugin.GroupGetMemberUserIDs", GroupGetMemberIDsArgs{e}, &reply); err != nil {
		return nil, err
	}
	return reply.IDs, strErr(reply.Err)
}
func (r *groupRepositoryRPC) AddMembers(e *Group, userIDs []string) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.GroupAddMembers", GroupMembersArgs{e, userIDs}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *groupRepositoryRPC) RemoveMembers(e *Group, userIDs []string) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.GroupRemoveMembers", GroupMembersArgs{e, userIDs}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *groupRepositoryRPC) Create(e *Group) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.GroupCreate", GroupMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *groupRepositoryRPC) Update(e *Group) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.GroupUpdate", GroupMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *groupRepositoryRPC) Delete(e *Group) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.GroupDelete", GroupMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}

type bookingRepositoryRPC struct{ client *rpc.Client }

func (r *bookingRepositoryRPC) GetOne(id string) (*BookingDetails, error) {
	var reply BookingGetOneReply
	if err := r.client.Call("Plugin.BookingGetOne", BookingGetOneArgs{id}, &reply); err != nil {
		return nil, err
	}
	return reply.Booking, strErr(reply.Err)
}

type spaceRepositoryRPC struct{ client *rpc.Client }

func (r *spaceRepositoryRPC) GetOne(id string) (*Space, error) {
	var reply SpaceGetOneReply
	if err := r.client.Call("Plugin.SpaceGetOne", SpaceGetOneArgs{id}, &reply); err != nil {
		return nil, err
	}
	return reply.Space, strErr(reply.Err)
}
func (r *spaceRepositoryRPC) GetCount(orgID string) (int, error) {
	var reply IntReply
	if err := r.client.Call("Plugin.SpaceGetCount", SpaceGetCountArgs{orgID}, &reply); err != nil {
		return 0, err
	}
	return reply.V, nil
}

type locationRepositoryRPC struct{ client *rpc.Client }

func (r *locationRepositoryRPC) GetOne(id string) (*Location, error) {
	var reply LocationGetOneReply
	if err := r.client.Call("Plugin.LocationGetOne", LocationGetOneArgs{id}, &reply); err != nil {
		return nil, err
	}
	return reply.Location, strErr(reply.Err)
}
func (r *locationRepositoryRPC) GetCount(orgID string) (int, error) {
	var reply IntReply
	if err := r.client.Call("Plugin.LocationGetCount", LocationGetCountArgs{orgID}, &reply); err != nil {
		return 0, err
	}
	return reply.V, nil
}
func (r *locationRepositoryRPC) GetTimezone(location *Location) string {
	var reply StringReply
	_ = r.client.Call("Plugin.LocationGetTimezone", LocationGetTimezoneArgs{location}, &reply)
	return reply.V
}

type authProviderRepositoryRPC struct{ client *rpc.Client }

func (r *authProviderRepositoryRPC) Create(e *AuthProvider) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.AuthProviderCreate", AuthProviderMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *authProviderRepositoryRPC) Update(e *AuthProvider) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.AuthProviderUpdate", AuthProviderMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}

type authStateRepositoryRPC struct{ client *rpc.Client }

func (r *authStateRepositoryRPC) Create(e *AuthState) error {
	var reply ErrorReply
	if err := r.client.Call("Plugin.AuthStateCreate", AuthStateMutateArgs{e}, &reply); err != nil {
		return err
	}
	return strErr(reply.Err)
}
