package main

import (
	"strconv"
	"sync"
)

type UserPreferencesRepository struct {
}

type UserPreference struct {
	UserID string
	Name   string
	Value  string
}

type PreferenceName struct {
	Name string
	Type SettingType
}

var (
	PreferenceEnterTime            PreferenceName = PreferenceName{Name: "enter_time", Type: SettingTypeInt}
	PreferenceWorkdayStart         PreferenceName = PreferenceName{Name: "workday_start", Type: SettingTypeInt}
	PreferenceWorkdayEnd           PreferenceName = PreferenceName{Name: "workday_end", Type: SettingTypeInt}
	PreferenceWorkdays             PreferenceName = PreferenceName{Name: "workdays", Type: SettingTypeIntArray}
	PreferenceLocation             PreferenceName = PreferenceName{Name: "location_id", Type: SettingTypeString}
	PreferenceBookedColor          PreferenceName = PreferenceName{Name: "booked_color", Type: SettingTypeString}
	PreferenceNotBookedColor       PreferenceName = PreferenceName{Name: "not_booked_color", Type: SettingTypeString}
	PreferenceSelfBookedColor      PreferenceName = PreferenceName{Name: "self_booked_color", Type: SettingTypeString}
	PreferencePartiallyBookedColor PreferenceName = PreferenceName{Name: "partially_booked_color", Type: SettingTypeString}
	PreferenceBuddyBookedColor     PreferenceName = PreferenceName{Name: "buddy_booked_color", Type: SettingTypeString}
	PreferenceCalDAVURL            PreferenceName = PreferenceName{Name: "caldav_url", Type: SettingTypeString}
	PreferenceCalDAVUser           PreferenceName = PreferenceName{Name: "caldav_user", Type: SettingTypeString}
	PreferenceCalDAVPass           PreferenceName = PreferenceName{Name: "caldav_pass", Type: SettingTypeEncryptedString}
	PreferenceCalDAVPath           PreferenceName = PreferenceName{Name: "caldav_path", Type: SettingTypeString}
)

var (
	PreferenceEnterTimeNow         int = 1
	PreferenceEnterTimeNextDay     int = 2
	PreferenceEnterTimeNextWorkday int = 3
)

var userPreferencesRepository *UserPreferencesRepository
var userPreferencesRepositoryOnce sync.Once

func GetUserPreferencesRepository() *UserPreferencesRepository {
	userPreferencesRepositoryOnce.Do(func() {
		userPreferencesRepository = &UserPreferencesRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS users_preferences (" +
			"user_id uuid NOT NULL, " +
			"name VARCHAR NOT NULL, " +
			"value VARCHAR NOT NULL DEFAULT '', " +
			"PRIMARY KEY (user_id, name))")
		if err != nil {
			panic(err)
		}
	})
	return userPreferencesRepository
}

func (r *UserPreferencesRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// nothing yet
}

func (r *UserPreferencesRepository) Set(userID string, name string, value string) error {
	_, err := GetDatabase().DB().Exec("INSERT INTO users_preferences (user_id, name, value) "+
		"VALUES ($1, $2, $3) "+
		"ON CONFLICT (user_id, name) DO UPDATE SET value = $3",
		userID, name, value)
	return err
}

func (r *UserPreferencesRepository) Get(userID string, name string) (string, error) {
	var res string
	err := GetDatabase().DB().QueryRow("SELECT value FROM users_preferences "+
		"WHERE user_id = $1 AND name = $2",
		userID, name).Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (r *UserPreferencesRepository) GetInt(userID string, name string) (int, error) {
	res, err := r.Get(userID, name)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(res)
	return i, err
}

/*
func (r *UserPreferencesRepository) GetIntArray(userID string, name string) (int, error) {
	res, err := r.Get(userID, name)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(res)
	return i, err
} */

func (r *UserPreferencesRepository) GetBool(userID string, name string) (bool, error) {
	res, err := r.Get(userID, name)
	if err != nil {
		return false, err
	}
	b := (res == "1")
	return b, err
}

func (r *UserPreferencesRepository) GetAll(userID string) ([]*UserPreference, error) {
	var result []*UserPreference
	rows, err := GetDatabase().DB().Query("SELECT user_id, name, value FROM users_preferences "+
		"WHERE user_id = $1 "+
		"ORDER BY name", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &UserPreference{}
		err = rows.Scan(&e.UserID, &e.Name, &e.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserPreferencesRepository) InitDefaultSettingsForUser(userID string) error {
	_, err := GetDatabase().DB().Exec("INSERT INTO users_preferences (user_id, name, value) "+
		"VALUES "+
		"($1, '"+PreferenceEnterTime.Name+"', '"+strconv.Itoa(PreferenceEnterTimeNow)+"'), "+
		"($1, '"+PreferenceWorkdayStart.Name+"', '9'), "+
		"($1, '"+PreferenceWorkdayEnd.Name+"', '17'), "+
		"($1, '"+PreferenceWorkdays.Name+"', '1,2,3,4,5'), "+
		"($1, '"+PreferenceLocation.Name+"', ''), "+
		"($1, '"+PreferenceBookedColor.Name+"', '#ff453a'), "+
		"($1, '"+PreferenceNotBookedColor.Name+"', '#30d158'), "+
		"($1, '"+PreferenceSelfBookedColor.Name+"', '#b825de'), "+
		"($1, '"+PreferencePartiallyBookedColor.Name+"', '#ff9100'), "+
		"($1, '"+PreferenceBuddyBookedColor.Name+"', '#2415c5') "+
		"ON CONFLICT (user_id, name) DO NOTHING",
		userID)
	return err
}

func (r *UserPreferencesRepository) InitDefaultSettings(userIDs []string) error {
	for _, userID := range userIDs {
		if err := r.InitDefaultSettingsForUser(userID); err != nil {
			return err
		}
	}
	return nil
}

func (r *UserPreferencesRepository) DeleteAll(userID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM users_preferences WHERE user_id = $1", userID)
	return err
}
