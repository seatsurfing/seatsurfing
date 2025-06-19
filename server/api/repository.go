package api

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	RunSchemaUpgrade(curVersion, targetVersion int)
}

type NullString string
type NullTime *time.Time
type NullUUID string

func CheckNullString(s NullString) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: string(s),
		Valid:  true,
	}
}

func CheckNullUUID(s NullUUID) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: string(s),
		Valid:  true,
	}
}

func (s *NullString) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	strVal, ok := value.(string)
	if !ok {
		return errors.New("column is not a string")
	}
	*s = NullString(strVal)
	return nil
}

func (s NullString) Value() (driver.Value, error) {
	if len(s) == 0 { // if nil or empty string
		return nil, nil
	}
	return string(s), nil
}

func (s *NullUUID) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	uintVal, ok := value.([]uint8)
	if !ok {
		return errors.New("column is not a uuid")
	}
	u, err := uuid.ParseBytes([]byte(uintVal))
	if err != nil {
		return err
	}
	*s = NullUUID(u.String())
	return nil
}

func (s NullUUID) Value() (driver.Value, error) {
	if len(s) == 0 { // if nil or empty string
		return nil, nil
	}
	return string(s), nil
}
