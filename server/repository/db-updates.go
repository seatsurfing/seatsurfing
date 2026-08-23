package repository

import (
	"log"
	"strconv"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

func RunDBSchemaUpdates() {
	targetVersion := 55
	curVersion, err := GetSettingsRepository().GetGlobalInt(SettingDatabaseVersion.Name)
	log.Printf("Initializing database with schema version %d (current: %d) …\n", targetVersion, curVersion)
	if err != nil {
		curVersion = 0
	}
	repositories := []Repository{
		GetAuthProviderRepository(),
		GetAuthStateRepository(),
		GetAuthAttemptRepository(),
		GetBookingRepository(),
		GetBuddyRepository(),
		GetGroupRepository(),
		GetLocationRepository(),
		GetOrganizationRepository(),
		GetSpaceRepository(),
		GetUserRepository(),
		GetUserPreferencesRepository(),
		GetSettingsRepository(),
		GetRecurringBookingRepository(),
		GetRefreshTokenRepository(),
		GetSpaceAttributeRepository(),
		GetSpaceAttributeValueRepository(),
		GetMailLogRepository(),
		GetSessionRepository(),
		GetPasskeyRepository(),
		GetLocationFloorPlanRepository(),
		GetAuthProviderMappingRepository(),
		GetUserRoleRepository(),
		GetRoleRepository(),
	}
	for _, repository := range repositories {
		repository.RunSchemaUpgrade(curVersion, targetVersion)
	}

	if curVersion < 43 {
		if _, err := GetDatabase().DB().Exec("DROP TABLE IF EXISTS debug_time_issues"); err != nil {
			panic(err)
		}
	}
	if curVersion < 53 {
		// Dropped last, after the role repository has read it to seed role
		// assignments. Nothing consults it at runtime any more: what it
		// encoded is now either a role assignment or users.account_type.
		if _, err := GetDatabase().DB().Exec("ALTER TABLE users DROP COLUMN IF EXISTS role"); err != nil {
			panic(err)
		}
	}
	GetSettingsRepository().SetGlobal(SettingDatabaseVersion.Name, strconv.Itoa(targetVersion))
	SetGlobalInstallID()
}

func SetGlobalInstallID() {
	ID, err := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	if (err != nil) || (ID == "") {
		GetSettingsRepository().SetGlobal(SettingInstallID.Name, uuid.New().String())
	}
}

func InitDefaultOrgSettings() {
	log.Println("Configuring default settings for orgs …")
	list, err := GetOrganizationRepository().GetAllIDs()
	if err != nil {
		panic(err)
	}
	if err := GetSettingsRepository().InitDefaultSettings(list); err != nil {
		panic(err)
	}
}

func InitDefaultUserPreferences() {
	log.Println("Configuring default preferences for users …")
	list, err := GetUserRepository().GetAllIDs()
	if err != nil {
		panic(err)
	}
	if err := GetUserPreferencesRepository().InitDefaultSettings(list); err != nil {
		panic(err)
	}
}
