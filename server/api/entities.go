package api

import (
	"strings"
	"time"
)

// ─── User ────────────────────────────────────────────────────────────────────

type UserRole int

const DefaultUserLimit int = 10

const (
	UserRoleUser             UserRole = 0
	UserRoleSpaceAdmin       UserRole = 10
	UserRoleOrgAdmin         UserRole = 20
	UserRoleServiceAccountRO UserRole = 21
	UserRoleServiceAccountRW UserRole = 22
	UserRoleSuperAdmin       UserRole = 90
)

type User struct {
	ID                     string
	OrganizationID         string
	Email                  string
	Firstname              string
	Lastname               string
	AtlassianID            NullString
	HashedPassword         NullString
	AuthProviderID         NullUUID
	PasswordPending        bool
	PasswordUpdateRequired bool
	Role                   UserRole
	Disabled               bool
	BanExpiry              *time.Time
	LastActivityAtUTC      *time.Time
	TotpSecret             NullString
	ApiToken               NullString
}

func (u *User) GetDisplayName() string {
	if u.Firstname != "" || u.Lastname != "" {
		return strings.TrimSpace(u.Firstname + " " + u.Lastname)
	}
	return u.Email
}

func (u *User) GetSafeRecipientName() string {
	if u.Firstname != "" {
		return u.Firstname
	}
	idx := strings.LastIndex(u.Email, "@")
	if idx == -1 {
		return strings.Title(u.Email) //nolint:staticcheck
	}
	return strings.Title(u.Email[:idx]) //nolint:staticcheck
}

// ─── Organization ────────────────────────────────────────────────────────────

type Organization struct {
	ID               string
	Name             string
	ContactFirstname string
	ContactLastname  string
	ContactEmail     string
	Language         string
	SignupDate       time.Time
}

type Domain struct {
	DomainName     string
	OrganizationID string
	Active         bool
	VerifyToken    string
	Primary        bool
	Accessible     bool
	AccessCheck    *time.Time
}

// ─── Group ───────────────────────────────────────────────────────────────────

type Group struct {
	ID             string
	OrganizationID string
	Name           string
}

// ─── Location ────────────────────────────────────────────────────────────────

type Location struct {
	ID                    string
	OrganizationID        string
	Name                  string
	MapWidth              uint
	MapHeight             uint
	MapScale              float64
	MapMimeType           string
	MapType               string
	Description           string
	MaxConcurrentBookings uint
	Timezone              string
	Enabled               bool
	BookableDays          string
}

// ─── Space ───────────────────────────────────────────────────────────────────

type Space struct {
	ID             string
	LocationID     string
	Name           string
	X              uint
	Y              uint
	Width          uint
	Height         uint
	Rotation       uint
	RequireSubject bool
	Enabled        bool
	KioskEnabled   bool
	Shape          string
	FontSize       string
}

type SpaceDetails struct {
	Location Location
	Space
}

// ─── Booking ─────────────────────────────────────────────────────────────────

type Booking struct {
	ID                    string
	UserID                string
	SpaceID               string
	Enter                 time.Time
	Leave                 time.Time
	CalDavID              string
	Approved              bool
	Subject               string
	RecurringID           NullUUID
	CreatedAtUTC          *time.Time
	LastInfoMailSentAtUTC *time.Time
	ReminderSentAtUTC     *time.Time
}

type BookingDetails struct {
	Space         SpaceDetails
	UserEmail     string
	UserFirstname string
	UserLastname  string
	Booking
}

// ─── AuthProvider ─────────────────────────────────────────────────────────────

type AuthProviderType int

const OAuth2 AuthProviderType = 1

type AuthProvider struct {
	ID                     string
	OrganizationID         string
	Name                   string
	ProviderType           int
	AuthURL                string
	TokenURL               string
	AuthStyle              int
	Scopes                 string
	UserInfoURL            string
	UserInfoEmailField     string
	UserInfoFirstnameField string
	UserInfoLastnameField  string
	ClientID               string
	ClientSecret           string
	LogoutURL              string
	ProfilePageURL         string
	ReadOnly               bool
}

// ─── AuthState ────────────────────────────────────────────────────────────────

type AuthStateType int

const (
	AuthRequestState         AuthStateType = 1
	AuthResponseCache        AuthStateType = 2
	AuthAtlassian            AuthStateType = 3
	AuthMergeRequest         AuthStateType = 4
	AuthResetPasswordRequest AuthStateType = 5
	AuthChangeOrgEmail       AuthStateType = 6
	AuthDeleteOrg            AuthStateType = 7
	AuthTotpSetup            AuthStateType = 8
	AuthInviteUser           AuthStateType = 9
	AuthPasskeyRegistration  AuthStateType = 10
	AuthPasskeyLogin         AuthStateType = 11
	AuthPasskey2FA           AuthStateType = 12
)

type AuthState struct {
	ID             string
	AuthProviderID string
	Expiry         time.Time
	AuthStateType  AuthStateType
	Payload        string
}

// AuthStateLoginPayload is used for JSON marshaling into AuthState.Payload.
type AuthStateLoginPayload struct {
	UserID    string `json:"userId"`
	LoginType string `json:"type"`
	Redirect  string `json:"redirect,omitempty"`
}

// ─── Settings ────────────────────────────────────────────────────────────────

type SettingName struct {
	Name string
	Type SettingType
}

const (
	SettingSubjectDefaultDisabled = 1
	SettingSubjectDefaultOptional = 2
	SettingSubjectDefaultRequired = 3
)

const (
	SettingEnforceTOTPDisabled   = 0
	SettingEnforceTOTPAllUsers   = 1
	SettingEnforceTOTPAdminsOnly = 2
)

var (
	SettingInstallID                      SettingName = SettingName{Name: "install_id", Type: SettingTypeString}
	SettingDatabaseVersion                SettingName = SettingName{Name: "db_version", Type: SettingTypeInt}
	SettingEmailFooterPrefix              SettingName = SettingName{Name: "email_footer_", Type: SettingTypeString}
	SettingAllowAnyUser                   SettingName = SettingName{Name: "allow_any_user", Type: SettingTypeBool}
	SettingConfluenceServerSharedSecret   SettingName = SettingName{Name: "confluence_server_shared_secret", Type: SettingTypeString}
	SettingConfluenceAnonymous            SettingName = SettingName{Name: "confluence_anonymous", Type: SettingTypeBool}
	SettingMaxBookingsPerUser             SettingName = SettingName{Name: "max_bookings_per_user", Type: SettingTypeInt}
	SettingMaxConcurrentBookingsPerUser   SettingName = SettingName{Name: "max_concurrent_bookings_per_user", Type: SettingTypeInt}
	SettingMaxDaysInAdvance               SettingName = SettingName{Name: "max_days_in_advance", Type: SettingTypeInt}
	SettingBookingRetentionEnabled        SettingName = SettingName{Name: "booking_retention_enabled", Type: SettingTypeBool}
	SettingBookingRetentionDays           SettingName = SettingName{Name: "booking_retention_days", Type: SettingTypeInt}
	SettingEnableMaxHourBeforeDelete      SettingName = SettingName{Name: "enable_max_hours_before_delete", Type: SettingTypeBool}
	SettingMaxHoursBeforeDelete           SettingName = SettingName{Name: "max_hours_before_delete", Type: SettingTypeInt}
	SettingMinBookingDurationHours        SettingName = SettingName{Name: "min_booking_duration_hours", Type: SettingTypeInt}
	SettingMaxBookingDurationHours        SettingName = SettingName{Name: "max_booking_duration_hours", Type: SettingTypeInt}
	SettingTargetUtilizationHoursPerWeek  SettingName = SettingName{Name: "target_utilization_hours_per_week", Type: SettingTypeInt}
	SettingMaxHoursPartiallyBooked        SettingName = SettingName{Name: "max_hours_partially_booked", Type: SettingTypeInt}
	SettingMaxHoursPartiallyBookedEnabled SettingName = SettingName{Name: "max_hours_partially_booked_enabled", Type: SettingTypeBool}
	SettingDailyBasisBooking              SettingName = SettingName{Name: "daily_basis_booking", Type: SettingTypeBool}
	SettingNoAdminRestrictions            SettingName = SettingName{Name: "no_admin_restrictions", Type: SettingTypeBool}
	SettingCustomLogoUrl                  SettingName = SettingName{Name: "custom_logo_url", Type: SettingTypeString}
	SettingShowNames                      SettingName = SettingName{Name: "show_names", Type: SettingTypeBool}
	SettingAllowBookingsNonExistingUsers  SettingName = SettingName{Name: "allow_booking_nonexist_users", Type: SettingTypeBool}
	SettingDisableBuddies                 SettingName = SettingName{Name: "disable_buddies", Type: SettingTypeBool}
	SettingDefaultTimezone                SettingName = SettingName{Name: "default_timezone", Type: SettingTypeString}
	SettingAllowRecurringBookings         SettingName = SettingName{Name: "allow_recurring_bookings", Type: SettingTypeBool}
	SettingNewUserDefaultMailNotification SettingName = SettingName{Name: "new_user_default_mail_notification", Type: SettingTypeBool}
	SettingSubjectDefault                 SettingName = SettingName{Name: "subject_default", Type: SettingTypeInt}
	SettingFeatureNoUserLimit             SettingName = SettingName{Name: "feature_no_user_limit", Type: SettingTypeBool}
	SettingFeatureCustomDomains           SettingName = SettingName{Name: "feature_custom_domains", Type: SettingTypeBool}
	SettingFeatureGroups                  SettingName = SettingName{Name: "feature_groups", Type: SettingTypeBool}
	SettingFeatureAuthProviders           SettingName = SettingName{Name: "feature_auth_providers", Type: SettingTypeBool}
	SettingFeatureRecurringBookings       SettingName = SettingName{Name: "feature_recurring_bookings", Type: SettingTypeBool}
	SettingEnforceTOTP                    SettingName = SettingName{Name: "enforce_totp", Type: SettingTypeInt}
	SettingKioskSecret                    SettingName = SettingName{Name: "kiosk_access_secret", Type: SettingTypeString}
	SettingKioskModeEnabled               SettingName = SettingName{Name: "kiosk_mode_enabled", Type: SettingTypeBool}
	SettingFeatureKioskMode               SettingName = SettingName{Name: "feature_kiosk_mode", Type: SettingTypeBool}
	SettingHideReports                    SettingName = SettingName{Name: "hide_reports", Type: SettingTypeBool}
	SettingHideStats                      SettingName = SettingName{Name: "hide_stats", Type: SettingTypeBool}
)
