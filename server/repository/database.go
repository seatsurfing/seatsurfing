package repository

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"log"
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/config"

	_ "github.com/lib/pq"
)

type Database struct {
	Connection *sql.DB
}

var _databaseInstance *Database
var _databaseOnce sync.Once

func GetDatabase() *Database {
	_databaseOnce.Do(func() {
		_databaseInstance = &Database{}
		_databaseInstance.Open()
	})
	return _databaseInstance
}

func (db *Database) Open() {
	log.Println("Connecting to database...")
	conn, err := sql.Open("postgres", GetConfig().PostgresURL)
	if err != nil {
		panic(err)
	}
	err = conn.Ping()
	if err != nil {
		panic(err)
	}
	_, err = conn.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if err != nil {
		panic(err)
	}
	db.Connection = conn
	log.Println("Database connection established.")
}

func (db *Database) DB() *sql.DB {
	return db.Connection
}

func (db *Database) Close() {
	log.Println("Closing database connection...")
	db.Connection.Close()
}

type NullString string
type NullTime *time.Time

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
