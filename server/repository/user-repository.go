package repository

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type UserStore struct {
}

var userRepository *UserStore
var userRepositoryOnce sync.Once

func GetUserRepository() *UserStore {
	userRepositoryOnce.Do(func() {
		userRepository = &UserStore{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS users (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"organization_id uuid NOT NULL, " +
			"email VARCHAR NOT NULL, " +
			"org_admin boolean NOT NULL DEFAULT FALSE, " +
			"super_admin boolean NOT NULL DEFAULT FALSE, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS users_email ON users(email)")
		if err != nil {
			panic(err)
		}
	})
	return userRepository
}

func (r *UserStore) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 1 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS password VARCHAR, " +
			"ADD COLUMN IF NOT EXISTS auth_provider_id uuid"); err != nil {
			panic(err)
		}
	}
	if curVersion < 2 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ALTER COLUMN id SET DEFAULT uuid_generate_v4()"); err != nil {
			panic(err)
		}
	}
	if curVersion < 7 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS atlassian_id VARCHAR"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS users_atlassian_id ON users(atlassian_id)"); err != nil {
			panic(err)
		}
	}
	if curVersion < 13 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS role INT"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("UPDATE users SET role = " + strconv.Itoa(int(UserRoleUser))); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("UPDATE users SET role = " + strconv.Itoa(int(UserRoleOrgAdmin)) + " WHERE org_admin IS TRUE"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("UPDATE users SET role = " + strconv.Itoa(int(UserRoleSuperAdmin)) + " WHERE super_admin IS TRUE"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"DROP COLUMN IF EXISTS org_admin, " +
			"DROP COLUMN IF EXISTS super_admin"); err != nil {
			panic(err)
		}
	}
	if curVersion < 14 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS disabled boolean NOT NULL DEFAULT FALSE, " +
			"ADD COLUMN IF NOT EXISTS ban_expiry TIMESTAMP NULL DEFAULT NULL"); err != nil {
			panic(err)
		}
	}
	if curVersion < 19 {
		if _, err := GetDatabase().DB().Exec("DROP INDEX IF EXISTS users_email"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS users_email ON users(email, organization_id)"); err != nil {
			panic(err)
		}
	}
	if curVersion < 26 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS firstname VARCHAR NOT NULL DEFAULT '', " +
			"ADD COLUMN IF NOT EXISTS lastname VARCHAR NOT NULL DEFAULT ''"); err != nil {
			panic(err)
		}
	}
	if curVersion < 28 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS last_activity_at_utc TIMESTAMP NULL DEFAULT NULL"); err != nil {
			panic(err)
		}
	}
	if curVersion < 35 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS totp_secret VARCHAR NULL"); err != nil {
			panic(err)
		}
	}
	if curVersion < 36 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS password_pending boolean NOT NULL DEFAULT FALSE"); err != nil {
			panic(err)
		}
	}
	if curVersion < 40 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS password_update_required boolean NOT NULL DEFAULT FALSE"); err != nil {
			panic(err)
		}
	}
	if curVersion < 43 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN IF NOT EXISTS api_token VARCHAR NULL"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS users_api_token ON users(api_token) WHERE api_token IS NOT NULL"); err != nil {
			panic(err)
		}
	}
}

func (r *UserStore) Create(e *User) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO users "+
		"(organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, totp_secret, password_pending, password_update_required) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) "+
		"RETURNING id",
		e.OrganizationID, strings.ToLower(e.Email), e.Role, CheckNullString(e.HashedPassword), CheckNullUUID(e.AuthProviderID), CheckNullString(e.AtlassianID), e.Disabled, e.BanExpiry, e.Firstname, e.Lastname, CheckNullString(e.TotpSecret), e.PasswordPending, e.PasswordUpdateRequired).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	GetUserPreferencesRepository().InitDefaultSettingsForUser(e.ID)

	mailNotification, _ := GetSettingsRepository().GetBool(e.OrganizationID, SettingNewUserDefaultMailNotification.Name)
	if mailNotification {
		GetUserPreferencesRepository().Set(e.ID, PreferenceMailNotifications.Name, "1")
		GetUserPreferencesRepository().Set(e.ID, PreferenceMailReminder.Name, "1")
	}

	for _, plg := range GetPlugins() {
		plg.OnUserCreated(e.ID)
	}
	return nil
}

func (r *UserStore) GetOne(id string) (*User, error) {
	e := &User{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, last_activity_at_utc, totp_secret, password_pending, password_update_required, api_token "+
		"FROM users "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry, &e.Firstname, &e.Lastname, &e.LastActivityAtUTC, &e.TotpSecret, &e.PasswordPending, &e.PasswordUpdateRequired, &e.ApiToken)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *UserStore) GetByEmail(organizationID string, email string) (*User, error) {
	e := &User{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, last_activity_at_utc, totp_secret, password_pending, password_update_required, api_token "+
		"FROM users "+
		"WHERE LOWER(email) = $1 AND organization_id = $2",
		strings.ToLower(email), organizationID).Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry, &e.Firstname, &e.Lastname, &e.LastActivityAtUTC, &e.TotpSecret, &e.PasswordPending, &e.PasswordUpdateRequired, &e.ApiToken)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *UserStore) GetUsersWithEmail(email string) ([]*User, error) {
	var result []*User
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, last_activity_at_utc, totp_secret, password_pending, password_update_required, api_token "+
		"FROM users "+
		"WHERE LOWER(email) = $1",
		strings.ToLower(email))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &User{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry, &e.Firstname, &e.Lastname, &e.LastActivityAtUTC, &e.TotpSecret, &e.PasswordPending, &e.PasswordUpdateRequired, &e.ApiToken)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserStore) GetByAtlassianID(atlassianID string) (*User, error) {
	e := &User{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, last_activity_at_utc, totp_secret, password_pending, password_update_required, api_token "+
		"FROM users "+
		"WHERE LOWER(atlassian_id) = $1",
		strings.ToLower(atlassianID)).Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry, &e.Firstname, &e.Lastname, &e.LastActivityAtUTC, &e.TotpSecret, &e.PasswordPending, &e.PasswordUpdateRequired, &e.ApiToken)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *UserStore) GetUsersWithAtlassianID(organizationID string) ([]*User, error) {
	var result []*User
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, last_activity_at_utc, totp_secret, password_pending, password_update_required, api_token "+
		"FROM users "+
		"WHERE organization_id = $1 AND (atlassian_id IS NOT NULL OR atlassian_id != '') "+
		"ORDER BY email", organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &User{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry, &e.Firstname, &e.Lastname, &e.LastActivityAtUTC, &e.TotpSecret, &e.PasswordPending, &e.PasswordUpdateRequired, &e.ApiToken)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserStore) UpdateAtlassianClientIDForUser(organizationID, userId, atlassianID string) error {
	_, err := GetDatabase().DB().Exec("UPDATE users SET "+
		"atlassian_id =  $3 "+
		"WHERE organization_id = $1 AND id = $2",
		organizationID, userId, strings.ToLower(atlassianID))
	return err
}

func (r *UserStore) UpdateAtlassianClientID(organizationID, oldClientID, newClientID string) error {
	_, err := GetDatabase().DB().Exec("UPDATE users SET "+
		"atlassian_id = REPLACE(atlassian_id, '@"+oldClientID+"', '@"+newClientID+"') ,"+
		"email = REPLACE(email, '@"+oldClientID+"', '@"+newClientID+"')"+
		"WHERE organization_id = $1 AND (atlassian_id IS NOT NULL OR atlassian_id != '')",
		organizationID)
	return err
}

func (r *UserStore) GetByKeyword(organizationID string, keyword string) ([]*User, error) {
	var result []*User
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, last_activity_at_utc, totp_secret, password_pending, password_update_required, api_token "+
		"FROM users "+
		"WHERE organization_id = $1 AND (LOWER(email) LIKE '%' || $2 || '%' OR LOWER(firstname) LIKE '%' || $2 || '%' OR LOWER(lastname) LIKE '%' || $2 || '%') "+
		"ORDER BY email", organizationID, strings.ToLower(keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &User{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry, &e.Firstname, &e.Lastname, &e.LastActivityAtUTC, &e.TotpSecret, &e.PasswordPending, &e.PasswordUpdateRequired, &e.ApiToken)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserStore) GetAll(organizationID string, maxResults int, offset int) ([]*User, error) {
	var result []*User
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, last_activity_at_utc, totp_secret, password_pending, password_update_required, api_token "+
		"FROM users "+
		"WHERE organization_id = $1 "+
		"ORDER BY email "+
		"LIMIT $2 OFFSET $3", organizationID, maxResults, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &User{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry, &e.Firstname, &e.Lastname, &e.LastActivityAtUTC, &e.TotpSecret, &e.PasswordPending, &e.PasswordUpdateRequired, &e.ApiToken)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserStore) GetAllByIDs(userIDs []string) ([]*User, error) {
	var result []*User
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, last_activity_at_utc, totp_secret, password_pending, password_update_required, api_token "+
		"FROM users "+
		"WHERE id = ANY($1) "+
		"ORDER BY email",
		pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &User{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry, &e.Firstname, &e.Lastname, &e.LastActivityAtUTC, &e.TotpSecret, &e.PasswordPending, &e.PasswordUpdateRequired, &e.ApiToken)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserStore) UsersExistAndBelongToOrg(organizationID string, userIDs []string) (bool, error) {
	var count int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(id) "+
		"FROM users "+
		"WHERE id = ANY($1) AND organization_id = $2",
		pq.Array(userIDs), organizationID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == len(userIDs), nil
}

func (r *UserStore) GetAllIDs() ([]string, error) {
	var result []string
	rows, err := GetDatabase().DB().Query("SELECT id " +
		"FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ID string
		err = rows.Scan(&ID)
		if err != nil {
			return nil, err
		}
		result = append(result, ID)
	}
	return result, nil
}

func (r *UserStore) GetByApiToken(tokenHash string) (*User, error) {
	e := &User{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry, firstname, lastname, last_activity_at_utc, totp_secret, password_pending, password_update_required, api_token "+
		"FROM users "+
		"WHERE api_token = $1 AND role IN ($2, $3)",
		tokenHash, UserRoleServiceAccountRO, UserRoleServiceAccountRW).Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry, &e.Firstname, &e.Lastname, &e.LastActivityAtUTC, &e.TotpSecret, &e.PasswordPending, &e.PasswordUpdateRequired, &e.ApiToken)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *UserStore) SetApiToken(userID string, tokenHash NullString) error {
	_, err := GetDatabase().DB().Exec("UPDATE users SET api_token = $1 WHERE id = $2",
		CheckNullString(tokenHash), userID)
	return err
}

func (r *UserStore) Update(e *User) error {
	_, err := GetDatabase().DB().Exec("UPDATE users SET "+
		"organization_id = $1, "+
		"email = $2, "+
		"role = $3, "+
		"password = $4, "+
		"auth_provider_id = $5, "+
		"atlassian_id = $6, "+
		"disabled = $7, "+
		"ban_expiry = $8, "+
		"firstname = $9, "+
		"lastname = $10, "+
		"last_activity_at_utc = $11, "+
		"totp_secret = $12, "+
		"password_pending = $13, "+
		"password_update_required = $14 "+
		"WHERE id = $15",
		e.OrganizationID, strings.ToLower(e.Email), e.Role, CheckNullString(e.HashedPassword), CheckNullUUID(e.AuthProviderID), CheckNullString(e.AtlassianID), e.Disabled, e.BanExpiry, e.Firstname, e.Lastname, e.LastActivityAtUTC, CheckNullString(e.TotpSecret), e.PasswordPending, e.PasswordUpdateRequired, e.ID)
	if err != nil {
		return err
	}
	for _, plg := range GetPlugins() {
		plg.OnUserUpdated(e.ID)
	}
	return nil
}

func (r *UserStore) Delete(e *User) error {
	for _, plg := range GetPlugins() {
		plg.OnBeforeUserDelete(e.ID)
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM bookings WHERE "+
		"bookings.user_id = $1", e.ID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM recurring_bookings WHERE "+
		"recurring_bookings.user_id = $1", e.ID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM users_groups WHERE "+
		"user_id = $1", e.ID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM users_preferences WHERE "+
		"user_id = $1", e.ID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM buddies WHERE "+
		"owner_id = $1 OR buddy_id = $1", e.ID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM users WHERE id = $1", e.ID)
	return err
}

func (r *UserStore) DeleteAll(organizationID string) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM buddies "+
		"WHERE owner_id IN (SELECT id FROM users WHERE organization_id = $1) OR "+
		"buddy_id IN (SELECT id FROM users WHERE organization_id = $1)", organizationID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM users_preferences WHERE "+
		"user_id IN (SELECT id FROM users WHERE organization_id = $1)", organizationID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM users_groups WHERE "+
		"user_id IN (SELECT id FROM users WHERE organization_id = $1)", organizationID); err != nil {
		return err
	}
	// Also delete refresh tokens
	if _, err := GetDatabase().DB().Exec("DELETE FROM refresh_tokens WHERE "+
		"user_id IN (SELECT id FROM users WHERE organization_id = $1)", organizationID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM users WHERE organization_id = $1", organizationID)
	return err
}

func (r *UserStore) GetCountAll() (int, error) {
	var res int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(id) " +
		"FROM users").Scan(&res)
	return res, err
}

func (r *UserStore) GetCount(organizationID string) (int, error) {
	var res int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(id) "+
		"FROM users "+
		"WHERE organization_id = $1",
		organizationID).Scan(&res)
	return res, err
}

func (r *UserStore) GetHashedPassword(password string) string {
	pwHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(pwHash)
}

func (r *UserStore) CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (r *UserStore) MergeUsers(source, target *User) error {
	if source.OrganizationID != target.OrganizationID {
		return errors.New("Organization ID of source and target users don't match")
	}
	if _, err := GetDatabase().DB().Exec("UPDATE bookings SET user_id = $2 WHERE user_id = $1", source.ID, target.ID); err != nil {
		return err
	}
	if target.AtlassianID == "" {
		target.AtlassianID = source.AtlassianID
	}
	target.Role = UserRole(MaxOf(int(target.Role), int(source.Role)))
	if err := r.Delete(source); err != nil {
		return err
	}
	if err := r.Update(target); err != nil {
		return err
	}
	return nil
}

func (r *UserStore) EnableUsersWithExpiredBan() error {
	_, err := GetDatabase().DB().Exec("UPDATE users "+
		"SET disabled = FALSE, ban_expiry = NULL "+
		"WHERE disabled = TRUE AND ban_expiry <= $1", time.Now())
	return err
}

func (r *UserStore) CanCreateUser(org *Organization) bool {
	noUserLimit, _ := GetSettingsRepository().GetBool(org.ID, SettingFeatureNoUserLimit.Name)
	if noUserLimit {
		return true
	}
	curUsers, _ := GetUserRepository().GetCount(org.ID)
	return curUsers < DefaultUserLimit
}

func (r *UserStore) IsSpaceAdmin(user *User) bool {
	return int(user.Role) >= int(UserRoleSpaceAdmin)
}

func (r *UserStore) IsOrgAdmin(user *User) bool {
	return int(user.Role) >= int(UserRoleOrgAdmin)
}

func (r *UserStore) IsSuperAdmin(user *User) bool {
	return int(user.Role) >= int(UserRoleSuperAdmin)
}

func (r *UserStore) DeleteObsoleteConfluenceAnonymousUsers() (int, error) {
	timestamp := time.Now().Add(-24 * time.Hour)
	rows, err := GetDatabase().DB().Query("DELETE FROM users u "+
		"WHERE u.email LIKE 'confluence-anonymous-%' and "+
		"u.id not in (select distinct aa.user_id from auth_attempts aa where aa.successful = true and aa.timestamp > $1) "+
		"RETURNING u.id",
		timestamp)
	if err != nil {
		return 0, err
	}
	var userIDs []string
	defer rows.Close()
	for rows.Next() {
		var ID string
		err = rows.Scan(&ID)
		if err != nil {
			return 0, err
		}
		userIDs = append(userIDs, ID)
	}
	if len(userIDs) > 0 {
		if _, err := GetDatabase().DB().Exec("DELETE FROM bookings WHERE "+
			"bookings.user_id = ANY($1)", pq.Array(&userIDs)); err != nil {
			return 0, err
		}
	}
	return len(userIDs), nil
}

func (r *UserStore) HasAnyUserInOrgPasswordSet(organizationID string) (bool, error) {
	var result int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(*) FROM users WHERE "+
		"organization_id = $1 AND password IS NOT NULL AND password != ''", organizationID).Scan(&result)
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (r *UserStore) HasAnyUserWithAuthProvider(authProviderID string) (bool, error) {
	var result int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(*) FROM users WHERE "+
		"auth_provider_id = $1", authProviderID).Scan(&result)
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
