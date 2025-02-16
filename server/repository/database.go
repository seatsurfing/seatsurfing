package repository

import (
	"database/sql"
	"log"
	"sync"

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
