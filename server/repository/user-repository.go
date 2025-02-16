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

type UserRepository struct {
}

type UserRole int

const (
	UserRoleUser       UserRole = 0
	UserRoleSpaceAdmin UserRole = 10
	UserRoleOrgAdmin   UserRole = 20
	UserRoleSuperAdmin UserRole = 90
)

type User struct {
	ID             string
	OrganizationID string
	Email          string
	AtlassianID    NullString
	HashedPassword NullString
	AuthProviderID NullString
	Role           UserRole
	Disabled       bool
	BanExpiry      *time.Time
}

var userRepository *UserRepository
var userRepositoryOnce sync.Once

func GetUserRepository() *UserRepository {
	userRepositoryOnce.Do(func() {
		userRepository = &UserRepository{}
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

func (r *UserRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 1 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN password VARCHAR, " +
			"ADD COLUMN auth_provider_id uuid"); err != nil {
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
			"ADD COLUMN atlassian_id VARCHAR"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS users_atlassian_id ON users(atlassian_id)"); err != nil {
			panic(err)
		}
	}
	if curVersion < 13 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN role INT"); err != nil {
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
			"DROP COLUMN org_admin, " +
			"DROP COLUMN super_admin"); err != nil {
			panic(err)
		}
	}
	if curVersion < 14 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users " +
			"ADD COLUMN disabled boolean NOT NULL DEFAULT FALSE, " +
			"ADD COLUMN ban_expiry TIMESTAMP NULL DEFAULT NULL"); err != nil {
			panic(err)
		}
	}
}

func (r *UserRepository) Create(e *User) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO users "+
		"(organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) "+
		"RETURNING id",
		e.OrganizationID, strings.ToLower(e.Email), e.Role, CheckNullString(e.HashedPassword), CheckNullString(e.AuthProviderID), CheckNullString(e.AtlassianID), e.Disabled, e.BanExpiry).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	GetUserPreferencesRepository().InitDefaultSettingsForUser(e.ID)
	return nil
}

func (r *UserRepository) GetOne(id string) (*User, error) {
	e := &User{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry "+
		"FROM users "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *UserRepository) GetByEmail(email string) (*User, error) {
	e := &User{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry "+
		"FROM users "+
		"WHERE LOWER(email) = $1",
		strings.ToLower(email)).Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry)
	if err != nil {
		return nil, err
	}
	return e, nil
}
func (r *UserRepository) GetByAtlassianID(atlassianID string) (*User, error) {
	e := &User{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry "+
		"FROM users "+
		"WHERE LOWER(atlassian_id) = $1",
		strings.ToLower(atlassianID)).Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *UserRepository) GetUsersWithAtlassianID(organizationID string) ([]*User, error) {
	var result []*User
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry "+
		"FROM users "+
		"WHERE organization_id = $1 AND (atlassian_id IS NOT NULL OR atlassian_id != '') "+
		"ORDER BY email", organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &User{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserRepository) UpdateAtlassianClientIDForUser(organizationID, userId, atlassianID string) error {
	_, err := GetDatabase().DB().Exec("UPDATE users SET "+
		"atlassian_id =  $3 "+
		"WHERE organization_id = $1 AND id = $2",
		organizationID, userId, strings.ToLower(atlassianID))
	return err
}

func (r *UserRepository) UpdateAtlassianClientID(organizationID, oldClientID, newClientID string) error {
	_, err := GetDatabase().DB().Exec("UPDATE users SET "+
		"atlassian_id = REPLACE(atlassian_id, '@"+oldClientID+"', '@"+newClientID+"') ,"+
		"email = REPLACE(email, '@"+oldClientID+"', '@"+newClientID+"')"+
		"WHERE organization_id = $1 AND (atlassian_id IS NOT NULL OR atlassian_id != '')",
		organizationID)
	return err
}

func (r *UserRepository) GetByKeyword(organizationID string, keyword string) ([]*User, error) {
	var result []*User
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry "+
		"FROM users "+
		"WHERE organization_id = $1 AND LOWER(email) LIKE '%' || $2 || '%' "+
		"ORDER BY email", organizationID, strings.ToLower(keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &User{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserRepository) GetAll(organizationID string, maxResults int, offset int) ([]*User, error) {
	var result []*User
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, email, role, password, auth_provider_id, atlassian_id, disabled, ban_expiry "+
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
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Email, &e.Role, &e.HashedPassword, &e.AuthProviderID, &e.AtlassianID, &e.Disabled, &e.BanExpiry)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserRepository) GetAllIDs() ([]string, error) {
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

func (r *UserRepository) Update(e *User) error {
	_, err := GetDatabase().DB().Exec("UPDATE users SET "+
		"organization_id = $1, "+
		"email = $2, "+
		"role = $3, "+
		"password = $4, "+
		"auth_provider_id = $5, "+
		"atlassian_id = $6, "+
		"disabled = $7, "+
		"ban_expiry = $8 "+
		"WHERE id = $9",
		e.OrganizationID, strings.ToLower(e.Email), e.Role, CheckNullString(e.HashedPassword), CheckNullString(e.AuthProviderID), CheckNullString(e.AtlassianID), e.Disabled, e.BanExpiry, e.ID)
	return err
}

func (r *UserRepository) Delete(e *User) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM bookings WHERE "+
		"bookings.user_id = $1", e.ID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM users WHERE id = $1", e.ID)
	return err
}

func (r *UserRepository) DeleteAll(organizationID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM users WHERE organization_id = $1", organizationID)
	return err
}

func (r *UserRepository) GetCount(organizationID string) (int, error) {
	var res int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(id) "+
		"FROM users "+
		"WHERE organization_id = $1",
		organizationID).Scan(&res)
	return res, err
}

func (r *UserRepository) GetHashedPassword(password string) string {
	pwHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(pwHash)
}

func (r *UserRepository) CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (r *UserRepository) MergeUsers(source, target *User) error {
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

func (r *UserRepository) EnableUsersWithExpiredBan() error {
	_, err := GetDatabase().DB().Exec("UPDATE users "+
		"SET disabled = FALSE, ban_expiry = NULL "+
		"WHERE disabled = TRUE AND ban_expiry <= $1", time.Now())
	return err
}

func (r *UserRepository) CanCreateUser(org *Organization) bool {
	maxUsers, _ := GetSettingsRepository().GetInt(org.ID, SettingSubscriptionMaxUsers.Name)
	curUsers, _ := GetUserRepository().GetCount(org.ID)
	return curUsers < maxUsers
}

func (r *UserRepository) IsSpaceAdmin(user *User) bool {
	return int(user.Role) >= int(UserRoleSpaceAdmin)
}

func (r *UserRepository) IsOrgAdmin(user *User) bool {
	return int(user.Role) >= int(UserRoleOrgAdmin)
}

func (r *UserRepository) IsSuperAdmin(user *User) bool {
	return int(user.Role) >= int(UserRoleSuperAdmin)
}

func (r *UserRepository) DeleteObsoleteConfluenceAnonymousUsers() (int, error) {
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

func (r *UserRepository) HasAnyUserInOrgPasswordSet(organizationID string) (bool, error) {
	var result int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(*) FROM users WHERE "+
		"organization_id = $1 AND password IS NOT NULL AND password != ''", organizationID).Scan(&result)
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
