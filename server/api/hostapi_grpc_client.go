package api

import (
	"context"

	"github.com/seatsurfing/seatsurfing/server/api/commonpb"
	"github.com/seatsurfing/seatsurfing/server/api/hostapipb"
	"google.golang.org/grpc"
)

// HostAPIGRPC runs in the PLUGIN process. It implements HostAPI by forwarding
// every call to the host over gRPC. Held for the plugin process lifetime -
// dialed once, eagerly, at plugin startup (not lazily inside OnInit), so the
// plugin can recover from a host restart on its own (see the connect-driven
// resilience design).
type HostAPIGRPC struct {
	conn   *grpc.ClientConn
	client hostapipb.HostAPIServiceClient
}

func NewHostAPIGRPC(conn *grpc.ClientConn) *HostAPIGRPC {
	return &HostAPIGRPC{conn: conn, client: hostapipb.NewHostAPIServiceClient(conn)}
}

var _ HostAPI = (*HostAPIGRPC)(nil)

func (h *HostAPIGRPC) GetSettingsRepository() SettingsRepository {
	return &settingsRepositoryGRPC{client: h.client}
}
func (h *HostAPIGRPC) GetUserRepository() UserRepository {
	return &userRepositoryGRPC{client: h.client}
}
func (h *HostAPIGRPC) GetOrganizationRepository() OrganizationRepository {
	return &organizationRepositoryGRPC{client: h.client}
}
func (h *HostAPIGRPC) GetGroupRepository() GroupRepository {
	return &groupRepositoryGRPC{client: h.client}
}
func (h *HostAPIGRPC) GetBookingRepository() BookingRepository {
	return &bookingRepositoryGRPC{client: h.client}
}
func (h *HostAPIGRPC) GetSpaceRepository() SpaceRepository {
	return &spaceRepositoryGRPC{client: h.client}
}
func (h *HostAPIGRPC) GetLocationRepository() LocationRepository {
	return &locationRepositoryGRPC{client: h.client}
}
func (h *HostAPIGRPC) GetAuthProviderRepository() AuthProviderRepository {
	return &authProviderRepositoryGRPC{client: h.client}
}
func (h *HostAPIGRPC) GetAuthStateRepository() AuthStateRepository {
	return &authStateRepositoryGRPC{client: h.client}
}
func (h *HostAPIGRPC) SendEmail(recipient, subject, body, language, orgID string) error {
	r, err := h.client.SendEmail(context.Background(), &hostapipb.SendEmailArgs{
		Recipient: recipient, Subject: subject, Body: body, Language: language, OrgId: orgID,
	})
	if err != nil {
		return err
	}
	return strErr(r.Err)
}
func (h *HostAPIGRPC) Encrypt(plaintext string) (string, error) {
	r, err := h.client.Encrypt(context.Background(), &hostapipb.EncryptArgs{Plaintext: plaintext})
	if err != nil {
		return "", err
	}
	return r.Result, strErr(r.Err)
}
func (h *HostAPIGRPC) Decrypt(ciphertext string) (string, error) {
	r, err := h.client.Decrypt(context.Background(), &hostapipb.DecryptArgs{Ciphertext: ciphertext})
	if err != nil {
		return "", err
	}
	return r.Result, strErr(r.Err)
}
func (h *HostAPIGRPC) IsValidLanguageCode(code string) bool {
	r, err := h.client.IsValidLanguageCode(context.Background(), &hostapipb.IsValidLangArgs{Code: code})
	if err != nil {
		return false
	}
	return r.V
}
func (h *HostAPIGRPC) DisablePasswordLogin() bool {
	r, err := h.client.DisablePasswordLogin(context.Background(), &commonpb.Empty{})
	if err != nil {
		return false
	}
	return r.V
}
func (h *HostAPIGRPC) FormatPublicURL(domain string) string {
	r, err := h.client.FormatPublicURL(context.Background(), &hostapipb.FormatPublicURLArgs{Domain: domain})
	if err != nil {
		return "https://" + domain
	}
	return r.V
}
func (h *HostAPIGRPC) IsDevelopmentMode() bool {
	r, err := h.client.IsDevelopmentMode(context.Background(), &commonpb.Empty{})
	if err != nil {
		return false
	}
	return r.V
}
func (h *HostAPIGRPC) GetPostgresURL() string {
	r, err := h.client.GetPostgresURL(context.Background(), &commonpb.Empty{})
	if err != nil {
		// No sane fallback exists here (unlike FormatPublicURL/
		// IsDevelopmentMode) - an empty string will make the subsequent
		// sql.Open/Ping in PluginDatabase.GetDatabase() fail loudly, which
		// is preferable to silently guessing a connection string.
		return ""
	}
	return r.V
}
func (h *HostAPIGRPC) GetEmailHTMLLayout() (string, error) {
	r, err := h.client.GetEmailHTMLLayout(context.Background(), &commonpb.Empty{})
	if err != nil {
		return "", err
	}
	return r.Html, strErr(r.Err)
}

// ─── Per-repository gRPC proxies (plugin side) ───────────────────────────────

type settingsRepositoryGRPC struct {
	client hostapipb.HostAPIServiceClient
}

func (r *settingsRepositoryGRPC) Get(orgID, name string) (string, error) {
	reply, err := r.client.SettingsGet(context.Background(), &hostapipb.SettingsGetArgs{OrgId: orgID, Name: name})
	if err != nil {
		return "", err
	}
	return reply.Value, strErr(reply.Err)
}
func (r *settingsRepositoryGRPC) GetBool(orgID, name string) (bool, error) {
	reply, err := r.client.SettingsGetBool(context.Background(), &hostapipb.SettingsGetArgs{OrgId: orgID, Name: name})
	if err != nil {
		return false, err
	}
	return reply.Value, strErr(reply.Err)
}
func (r *settingsRepositoryGRPC) GetInt(orgID, name string) (int, error) {
	reply, err := r.client.SettingsGetInt(context.Background(), &hostapipb.SettingsGetArgs{OrgId: orgID, Name: name})
	if err != nil {
		return 0, err
	}
	return int(reply.V), strErr(reply.Err)
}
func (r *settingsRepositoryGRPC) GetNullUUID() string {
	reply, err := r.client.SettingsGetNullUUID(context.Background(), &commonpb.Empty{})
	if err != nil {
		return ""
	}
	return reply.V
}
func (r *settingsRepositoryGRPC) GetOrgIDsByValue(name, value string) ([]string, error) {
	reply, err := r.client.SettingsGetOrgIDsByValue(context.Background(), &hostapipb.SettingsGetOrgIDsArgs{Name: name, Value: value})
	if err != nil {
		return nil, err
	}
	return reply.V, strErr(reply.Err)
}
func (r *settingsRepositoryGRPC) Set(orgID, name, value string) error {
	reply, err := r.client.SettingsSet(context.Background(), &hostapipb.SettingsSetArgs{OrgId: orgID, Name: name, Value: value})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *settingsRepositoryGRPC) Delete(orgID, name string) error {
	reply, err := r.client.SettingsDelete(context.Background(), &hostapipb.SettingsDeleteArgs{OrgId: orgID, Name: name})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}

type userRepositoryGRPC struct {
	client hostapipb.HostAPIServiceClient
}

func (r *userRepositoryGRPC) GetOne(id string) (*User, error) {
	reply, err := r.client.UserGetOne(context.Background(), &hostapipb.UserGetOneArgs{Id: id})
	if err != nil {
		return nil, err
	}
	return userFromProto(reply.User), strErr(reply.Err)
}
func (r *userRepositoryGRPC) GetAll(orgID string, maxResults, offset int) ([]*User, error) {
	reply, err := r.client.UserGetAll(context.Background(), &hostapipb.UserGetAllArgs{OrgId: orgID, MaxResults: int32(maxResults), Offset: int32(offset)})
	if err != nil {
		return nil, err
	}
	return usersFromProto(reply.Users), strErr(reply.Err)
}
func (r *userRepositoryGRPC) GetByEmail(orgID, email string) (*User, error) {
	reply, err := r.client.UserGetByEmail(context.Background(), &hostapipb.UserGetByEmailArgs{OrgId: orgID, Email: email})
	if err != nil {
		return nil, err
	}
	return userFromProto(reply.User), strErr(reply.Err)
}
func (r *userRepositoryGRPC) GetCount(orgID string) (int, error) {
	reply, err := r.client.UserGetCount(context.Background(), &hostapipb.UserGetCountArgs{OrgId: orgID})
	if err != nil {
		return 0, err
	}
	return int(reply.V), strErr(reply.Err)
}
func (r *userRepositoryGRPC) GetHashedPassword(password string) string {
	reply, err := r.client.UserGetHashedPassword(context.Background(), &hostapipb.UserHashPasswordArgs{Password: password})
	if err != nil {
		return ""
	}
	return reply.V
}
func (r *userRepositoryGRPC) GetUsersWithEmail(email string) ([]*User, error) {
	reply, err := r.client.UserGetUsersWithEmail(context.Background(), &hostapipb.UserGetUsersWithEmailArgs{Email: email})
	if err != nil {
		return nil, err
	}
	return usersFromProto(reply.Users), strErr(reply.Err)
}
func (r *userRepositoryGRPC) IsOrgAdmin(user *User) bool {
	reply, err := r.client.UserIsOrgAdmin(context.Background(), &hostapipb.UserIsAdminArgs{User: userToProto(user)})
	if err != nil {
		return false
	}
	return reply.V
}
func (r *userRepositoryGRPC) IsSuperAdmin(user *User) bool {
	reply, err := r.client.UserIsSuperAdmin(context.Background(), &hostapipb.UserIsAdminArgs{User: userToProto(user)})
	if err != nil {
		return false
	}
	return reply.V
}
func (r *userRepositoryGRPC) Create(e *User) error {
	reply, err := r.client.UserCreate(context.Background(), &hostapipb.UserMutateArgs{User: userToProto(e)})
	if err != nil {
		return err
	}
	if reply.User != nil {
		*e = *userFromProto(reply.User)
	}
	return strErr(reply.Err)
}
func (r *userRepositoryGRPC) Update(e *User) error {
	reply, err := r.client.UserUpdate(context.Background(), &hostapipb.UserMutateArgs{User: userToProto(e)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *userRepositoryGRPC) Delete(e *User) error {
	reply, err := r.client.UserDelete(context.Background(), &hostapipb.UserMutateArgs{User: userToProto(e)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}

type organizationRepositoryGRPC struct {
	client hostapipb.HostAPIServiceClient
}

func (r *organizationRepositoryGRPC) GetOne(id string) (*Organization, error) {
	reply, err := r.client.OrgGetOne(context.Background(), &hostapipb.OrgGetOneArgs{Id: id})
	if err != nil {
		return nil, err
	}
	return orgFromProto(reply.Org), strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) GetAll() ([]*Organization, error) {
	reply, err := r.client.OrgGetAll(context.Background(), &commonpb.Empty{})
	if err != nil {
		return nil, err
	}
	return orgsFromProto(reply.Orgs), strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) GetOneByDomain(domain string) (*Organization, error) {
	reply, err := r.client.OrgGetOneByDomain(context.Background(), &hostapipb.OrgGetByDomainArgs{Domain: domain})
	if err != nil {
		return nil, err
	}
	return orgFromProto(reply.Org), strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) GetByEmail(email string) (*Organization, error) {
	reply, err := r.client.OrgGetByEmail(context.Background(), &hostapipb.OrgGetByEmailArgs{Email: email})
	if err != nil {
		return nil, err
	}
	return orgFromProto(reply.Org), strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) GetAllDaysPassedSinceSignup(daysPassed int, settingExists string) ([]*Organization, error) {
	reply, err := r.client.OrgGetAllDaysPassedSinceSignup(context.Background(), &hostapipb.OrgGetDaysPassedArgs{DaysPassed: int32(daysPassed), SettingExists: settingExists})
	if err != nil {
		return nil, err
	}
	return orgsFromProto(reply.Orgs), strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) GetPrimaryDomain(e *Organization) (*Domain, error) {
	reply, err := r.client.OrgGetPrimaryDomain(context.Background(), &hostapipb.OrgGetPrimaryDomainArgs{Org: orgToProto(e)})
	if err != nil {
		return nil, err
	}
	return domainFromProto(reply.Domain), strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) Create(e *Organization) error {
	reply, err := r.client.OrgCreate(context.Background(), &hostapipb.OrgMutateArgs{Org: orgToProto(e)})
	if err != nil {
		return err
	}
	if reply.Org != nil {
		*e = *orgFromProto(reply.Org)
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) Update(e *Organization) error {
	reply, err := r.client.OrgUpdate(context.Background(), &hostapipb.OrgMutateArgs{Org: orgToProto(e)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) Delete(e *Organization) error {
	reply, err := r.client.OrgDelete(context.Background(), &hostapipb.OrgMutateArgs{Org: orgToProto(e)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) AddDomain(e *Organization, domain string, active bool) error {
	reply, err := r.client.OrgAddDomain(context.Background(), &hostapipb.OrgAddDomainArgs{Org: orgToProto(e), Domain: domain, Active: active})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) SetPrimaryDomain(e *Organization, domain string) error {
	reply, err := r.client.OrgSetPrimaryDomain(context.Background(), &hostapipb.OrgSetPrimaryDomainArgs{Org: orgToProto(e), Domain: domain})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *organizationRepositoryGRPC) CreateSampleData(org *Organization) error {
	reply, err := r.client.OrgCreateSampleData(context.Background(), &hostapipb.OrgCreateSampleDataArgs{Org: orgToProto(org)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}

type groupRepositoryGRPC struct {
	client hostapipb.HostAPIServiceClient
}

func (r *groupRepositoryGRPC) GetOne(id string) (*Group, error) {
	reply, err := r.client.GroupGetOne(context.Background(), &hostapipb.GroupGetOneArgs{Id: id})
	if err != nil {
		return nil, err
	}
	return groupFromProto(reply.Group), strErr(reply.Err)
}
func (r *groupRepositoryGRPC) GetAll(orgID string) ([]*Group, error) {
	reply, err := r.client.GroupGetAll(context.Background(), &hostapipb.GroupGetAllArgs{OrgId: orgID})
	if err != nil {
		return nil, err
	}
	return groupsFromProto(reply.Groups), strErr(reply.Err)
}
func (r *groupRepositoryGRPC) GetByName(orgID, name string) (*Group, error) {
	reply, err := r.client.GroupGetByName(context.Background(), &hostapipb.GroupGetByNameArgs{OrgId: orgID, Name: name})
	if err != nil {
		return nil, err
	}
	return groupFromProto(reply.Group), strErr(reply.Err)
}
func (r *groupRepositoryGRPC) GetMemberUserIDs(e *Group) ([]string, error) {
	reply, err := r.client.GroupGetMemberUserIDs(context.Background(), &hostapipb.GroupGetMemberIDsArgs{Group: groupToProto(e)})
	if err != nil {
		return nil, err
	}
	return reply.Ids, strErr(reply.Err)
}
func (r *groupRepositoryGRPC) AddMembers(e *Group, userIDs []string) error {
	reply, err := r.client.GroupAddMembers(context.Background(), &hostapipb.GroupMembersArgs{Group: groupToProto(e), UserIds: userIDs})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *groupRepositoryGRPC) RemoveMembers(e *Group, userIDs []string) error {
	reply, err := r.client.GroupRemoveMembers(context.Background(), &hostapipb.GroupMembersArgs{Group: groupToProto(e), UserIds: userIDs})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *groupRepositoryGRPC) Create(e *Group) error {
	reply, err := r.client.GroupCreate(context.Background(), &hostapipb.GroupMutateArgs{Group: groupToProto(e)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *groupRepositoryGRPC) Update(e *Group) error {
	reply, err := r.client.GroupUpdate(context.Background(), &hostapipb.GroupMutateArgs{Group: groupToProto(e)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *groupRepositoryGRPC) Delete(e *Group) error {
	reply, err := r.client.GroupDelete(context.Background(), &hostapipb.GroupMutateArgs{Group: groupToProto(e)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}

type bookingRepositoryGRPC struct {
	client hostapipb.HostAPIServiceClient
}

func (r *bookingRepositoryGRPC) GetOne(id string) (*BookingDetails, error) {
	reply, err := r.client.BookingGetOne(context.Background(), &hostapipb.BookingGetOneArgs{Id: id})
	if err != nil {
		return nil, err
	}
	return bookingDetailsFromProto(reply.Booking), strErr(reply.Err)
}

type spaceRepositoryGRPC struct {
	client hostapipb.HostAPIServiceClient
}

func (r *spaceRepositoryGRPC) GetOne(id string) (*Space, error) {
	reply, err := r.client.SpaceGetOne(context.Background(), &hostapipb.SpaceGetOneArgs{Id: id})
	if err != nil {
		return nil, err
	}
	return spaceFromProto(reply.Space), strErr(reply.Err)
}
func (r *spaceRepositoryGRPC) GetCount(orgID string) (int, error) {
	reply, err := r.client.SpaceGetCount(context.Background(), &hostapipb.SpaceGetCountArgs{OrgId: orgID})
	if err != nil {
		return 0, err
	}
	return int(reply.V), strErr(reply.Err)
}

type locationRepositoryGRPC struct {
	client hostapipb.HostAPIServiceClient
}

func (r *locationRepositoryGRPC) GetOne(id string) (*Location, error) {
	reply, err := r.client.LocationGetOne(context.Background(), &hostapipb.LocationGetOneArgs{Id: id})
	if err != nil {
		return nil, err
	}
	return locationFromProto(reply.Location), strErr(reply.Err)
}
func (r *locationRepositoryGRPC) GetCount(orgID string) (int, error) {
	reply, err := r.client.LocationGetCount(context.Background(), &hostapipb.LocationGetCountArgs{OrgId: orgID})
	if err != nil {
		return 0, err
	}
	return int(reply.V), strErr(reply.Err)
}
func (r *locationRepositoryGRPC) GetTimezone(location *Location) string {
	reply, err := r.client.LocationGetTimezone(context.Background(), &hostapipb.LocationGetTimezoneArgs{Location: locationToProto(location)})
	if err != nil {
		return ""
	}
	return reply.V
}

type authProviderRepositoryGRPC struct {
	client hostapipb.HostAPIServiceClient
}

func (r *authProviderRepositoryGRPC) Create(e *AuthProvider) error {
	reply, err := r.client.AuthProviderCreate(context.Background(), &hostapipb.AuthProviderMutateArgs{AuthProvider: authProviderToProto(e)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}
func (r *authProviderRepositoryGRPC) Update(e *AuthProvider) error {
	reply, err := r.client.AuthProviderUpdate(context.Background(), &hostapipb.AuthProviderMutateArgs{AuthProvider: authProviderToProto(e)})
	if err != nil {
		return err
	}
	return strErr(reply.Err)
}

type authStateRepositoryGRPC struct {
	client hostapipb.HostAPIServiceClient
}

func (r *authStateRepositoryGRPC) Create(e *AuthState) error {
	reply, err := r.client.AuthStateCreate(context.Background(), &hostapipb.AuthStateMutateArgs{AuthState: authStateToProto(e)})
	if err != nil {
		return err
	}
	if reply.AuthState != nil {
		*e = *authStateFromProto(reply.AuthState)
	}
	return strErr(reply.Err)
}
