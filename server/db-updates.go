package main

import (
	"log"
	"strconv"

	"github.com/google/uuid"
)

func RunDBSchemaUpdates() {
	targetVersion := 17
	log.Printf("Initializing database with schema version %d...\n", targetVersion)
	curVersion, err := GetSettingsRepository().GetGlobalInt(SettingDatabaseVersion.Name)
	if err != nil {
		curVersion = 0
	}
	repositories := []Repository{
		GetAuthProviderRepository(),
		GetAuthStateRepository(),
		GetAuthAttemptRepository(),
		GetBookingRepository(),
		GetLocationRepository(),
		GetOrganizationRepository(),
		GetSpaceRepository(),
		GetUserRepository(),
		GetUserPreferencesRepository(),
		GetSettingsRepository(),
		GetSignupRepository(),
		GetSubscriptionRepository(),
		GetRefreshTokenRepository(),
		GetDebugTimeIssuesRepository(),
		GetSpaceAttributeRepository(),
		GetSpaceAttributeValueRepository(),
	}
	for _, repository := range repositories {
		repository.RunSchemaUpgrade(curVersion, targetVersion)
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
	log.Println("Configuring default settings for orgs...")
	list, err := GetOrganizationRepository().GetAllIDs()
	if err != nil {
		panic(err)
	}
	if err := GetSettingsRepository().InitDefaultSettings(list); err != nil {
		panic(err)
	}
}

func InitDefaultUserPreferences() {
	log.Println("Configuring default preferences for users...")
	list, err := GetUserRepository().GetAllIDs()
	if err != nil {
		panic(err)
	}
	if err := GetUserPreferencesRepository().InitDefaultSettings(list); err != nil {
		panic(err)
	}
}
