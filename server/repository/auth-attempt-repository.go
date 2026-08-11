package repository

import (
	"strconv"
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/config"
)

type AuthAttemptRepository struct {
}

type AuthAttempt struct {
	ID               string
	OrganizationID   string
	UserID           string
	Email            string
	Timestamp        time.Time
	Successful       bool
	Method           string
	AuthProviderID   string
	AuthProviderName string
	ErrorCode        string
	ErrorDetail      string
	Device           string
}

const (
	AuthMethodPassword   = "password"
	AuthMethodTOTP       = "totp"
	AuthMethodPasskey    = "passkey"
	AuthMethodPasskey2FA = "passkey_2fa"
	AuthMethodOAuth      = "oauth"
	AuthMethodConfluence = "confluence"
)

const (
	AuthErrorUserNotFound          = "user_not_found"
	AuthErrorPasswordPending       = "password_pending"
	AuthErrorNoPasswordSet         = "no_password_set"
	AuthErrorBoundToAuthProvider   = "bound_to_auth_provider"
	AuthErrorUserDisabled          = "user_disabled"
	AuthErrorServiceAccount        = "service_account"
	AuthErrorWrongPassword         = "wrong_password"
	AuthErrorPasswordUpdateReq     = "password_update_required"
	AuthErrorTotpMissing           = "totp_missing"
	AuthErrorTotpReplay            = "totp_replay"
	AuthErrorTotpInvalid           = "totp_invalid"
	AuthErrorPasskeyStateInvalid   = "passkey_state_invalid"
	AuthErrorPasskeyAssertion      = "passkey_assertion_invalid"
	AuthErrorPasskeyCloneDetected  = "passkey_clone_detected"
	AuthErrorIdpProviderNotFound   = "idp_provider_not_found"
	AuthErrorIdpConfigInvalid      = "idp_config_invalid"
	AuthErrorIdpStateInvalid       = "idp_state_invalid"
	AuthErrorIdpCodeExchangeFailed = "idp_code_exchange_failed"
	AuthErrorIdpUserinfoFailed     = "idp_userinfo_failed"
	AuthErrorIdpAttributeMapping   = "idp_attribute_mapping_failed"
	AuthErrorIdpProviderMismatch   = "idp_provider_mismatch"
	AuthErrorUserLimitReached      = "user_limit_reached"
	AuthErrorUserCreateFailed      = "user_create_failed"
	AuthErrorOrgMismatch           = "org_mismatch"
	AuthErrorInternal              = "internal_error"
	AuthErrorConfluenceJwtInvalid  = "confluence_jwt_invalid"
)

const maxAuthErrorDetailLength = 2000

type AuthEvent struct {
	User           *User
	OrganizationID string
	Email          string
	AuthProviderID string
	Method         string
	Successful     bool
	ErrorCode      string
	ErrorDetail    string
	Device         string
	BanCheck       bool
}

type AuthAttemptFilter struct {
	OrganizationID string
	Start          time.Time
	End            time.Time
	EmailLike      string
	Method         string
	ErrorCode      string
	Successful     *bool
}

var authAttemptRepository *AuthAttemptRepository
var authAttemptRepositoryOnce sync.Once

func GetAuthAttemptRepository() *AuthAttemptRepository {
	authAttemptRepositoryOnce.Do(func() {
		authAttemptRepository = &AuthAttemptRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS auth_attempts (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"user_id uuid NULL, " +
			"email VARCHAR NOT NULL, " +
			"timestamp TIMESTAMP NOT NULL, " +
			"successful BOOLEAN, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_auth_attempts_user_id ON auth_attempts(user_id)")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_auth_attempts_email ON auth_attempts(email)")
		if err != nil {
			panic(err)
		}
	})
	return authAttemptRepository
}

func (r *AuthAttemptRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 50 {
		_, err := GetDatabase().DB().Exec("ALTER TABLE auth_attempts " +
			"ADD COLUMN IF NOT EXISTS organization_id uuid NULL, " +
			"ADD COLUMN IF NOT EXISTS method VARCHAR NOT NULL DEFAULT '', " +
			"ADD COLUMN IF NOT EXISTS auth_provider_id uuid NULL, " +
			"ADD COLUMN IF NOT EXISTS error_code VARCHAR NOT NULL DEFAULT '', " +
			"ADD COLUMN IF NOT EXISTS error_detail VARCHAR NOT NULL DEFAULT '', " +
			"ADD COLUMN IF NOT EXISTS device VARCHAR NOT NULL DEFAULT ''")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_auth_attempts_org_ts ON auth_attempts(organization_id, timestamp DESC)")
		if err != nil {
			panic(err)
		}
		// columns are added by GetAuthAttemptRepository(); backfill org for existing rows
		if _, err := GetDatabase().DB().Exec("UPDATE auth_attempts a SET organization_id = u.organization_id " +
			"FROM users u WHERE a.user_id = u.id AND a.organization_id IS NULL"); err != nil {
			panic(err)
		}
	}
}

func (r *AuthAttemptRepository) Create(e *AuthAttempt) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO auth_attempts "+
		"(organization_id, user_id, email, timestamp, successful, method, auth_provider_id, error_code, error_detail, device) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) "+
		"RETURNING id",
		CheckNullUUID(NullUUID(e.OrganizationID)), CheckNullUUID(NullUUID(e.UserID)), e.Email, e.Timestamp, e.Successful,
		e.Method, CheckNullUUID(NullUUID(e.AuthProviderID)), e.ErrorCode, e.ErrorDetail, e.Device).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *AuthAttemptRepository) RecordAuthEvent(e *AuthEvent) error {
	attempt := &AuthAttempt{
		OrganizationID: e.OrganizationID,
		Email:          e.Email,
		Timestamp:      time.Now(),
		Successful:     e.Successful,
		Method:         e.Method,
		AuthProviderID: e.AuthProviderID,
		ErrorCode:      e.ErrorCode,
		ErrorDetail:    e.ErrorDetail,
		Device:         e.Device,
	}
	if e.User != nil {
		attempt.UserID = e.User.ID
		attempt.OrganizationID = e.User.OrganizationID
		if attempt.Email == "" {
			attempt.Email = e.User.Email
		}
	}
	if len(attempt.ErrorDetail) > maxAuthErrorDetailLength {
		attempt.ErrorDetail = attempt.ErrorDetail[:maxAuthErrorDetailLength]
	}
	if err := r.Create(attempt); err != nil {
		return err
	}
	if e.BanCheck && e.User != nil {
		if err := r.checkBanUser(e.User); err != nil {
			return err
		}
	}
	return nil
}

func (r *AuthAttemptRepository) buildFilterQuery(f *AuthAttemptFilter) (string, []interface{}) {
	where := "WHERE a.organization_id = $1 AND a.timestamp BETWEEN $2 AND $3"
	args := []interface{}{f.OrganizationID, f.Start, f.End}
	if f.EmailLike != "" {
		args = append(args, "%"+f.EmailLike+"%")
		where += " AND a.email ILIKE $" + strconv.Itoa(len(args))
	}
	if f.Method != "" {
		args = append(args, f.Method)
		where += " AND a.method = $" + strconv.Itoa(len(args))
	}
	if f.ErrorCode != "" {
		args = append(args, f.ErrorCode)
		where += " AND a.error_code = $" + strconv.Itoa(len(args))
	}
	if f.Successful != nil {
		args = append(args, *f.Successful)
		where += " AND a.successful = $" + strconv.Itoa(len(args))
	}
	return where, args
}

func (r *AuthAttemptRepository) GetFiltered(f *AuthAttemptFilter, maxResults, offset int) ([]*AuthAttempt, error) {
	where, args := r.buildFilterQuery(f)
	args = append(args, maxResults, offset)
	limitParam := strconv.Itoa(len(args) - 1)
	offsetParam := strconv.Itoa(len(args))
	rows, err := GetDatabase().DB().Query("SELECT a.id, a.organization_id, a.user_id, a.email, a.timestamp, a.successful, "+
		"a.method, a.auth_provider_id, COALESCE(p.name, ''), a.error_code, a.error_detail, a.device "+
		"FROM auth_attempts a "+
		"LEFT JOIN auth_providers p ON p.id = a.auth_provider_id "+
		where+" "+
		"ORDER BY a.timestamp DESC "+
		"LIMIT $"+limitParam+" OFFSET $"+offsetParam, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*AuthAttempt{}
	for rows.Next() {
		e := &AuthAttempt{}
		var organizationID, userID, authProviderID NullUUID
		err = rows.Scan(&e.ID, &organizationID, &userID, &e.Email, &e.Timestamp, &e.Successful,
			&e.Method, &authProviderID, &e.AuthProviderName, &e.ErrorCode, &e.ErrorDetail, &e.Device)
		if err != nil {
			return nil, err
		}
		e.OrganizationID = string(organizationID)
		e.UserID = string(userID)
		e.AuthProviderID = string(authProviderID)
		result = append(result, e)
	}
	return result, nil
}

func (r *AuthAttemptRepository) CountFiltered(f *AuthAttemptFilter) (int, error) {
	where, args := r.buildFilterQuery(f)
	var count int
	if err := GetDatabase().DB().QueryRow("SELECT COUNT(a.id) FROM auth_attempts a "+where, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *AuthAttemptRepository) DeleteAll(organizationID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM auth_attempts WHERE organization_id = $1", organizationID)
	return err
}

func (r *AuthAttemptRepository) PurgeOld(maxAge time.Duration, batchSize int) (int, error) {
	limit := time.Now().Add(-maxAge)
	result, err := GetDatabase().DB().Exec("DELETE FROM auth_attempts WHERE id IN ("+
		"SELECT id FROM auth_attempts WHERE timestamp < $1 ORDER BY timestamp ASC LIMIT $2)",
		limit, batchSize)
	if err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(num), nil
}

func (r *AuthAttemptRepository) checkBanUser(user *User) error {
	var lastSuccessfulLogin time.Time
	if err := GetDatabase().DB().QueryRow("SELECT timestamp FROM auth_attempts WHERE user_id = $1 AND successful = TRUE ORDER BY timestamp DESC LIMIT 1",
		user.ID).Scan(&lastSuccessfulLogin); err != nil {
		lastSuccessfulLogin = time.Unix(0, 0)
	}
	var numFailedLogins int
	limit := time.Now().Add(time.Second * time.Duration(GetConfig().LoginProtectionSlidingWindowSeconds*-1))
	if err := GetDatabase().DB().QueryRow("SELECT COUNT(id) FROM auth_attempts "+
		"WHERE user_id = $1 AND timestamp > $2 AND timestamp > $3",
		user.ID, limit, lastSuccessfulLogin).Scan(&numFailedLogins); err != nil {
		return err
	}
	if numFailedLogins >= GetConfig().LoginProtectionMaxFails {
		banExpiry := time.Now().Add(time.Minute * time.Duration(GetConfig().LoginProtectionBanMinutes))
		user.Disabled = true
		user.BanExpiry = &banExpiry
		if err := GetUserRepository().Update(user); err != nil {
			return err
		}
	}
	return nil
}
