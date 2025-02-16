package repository

import (
	"strings"
	"sync"
	"time"
)

type SignupRepository struct {
}

type Signup struct {
	ID           string
	Date         time.Time
	Email        string
	Password     string
	Firstname    string
	Lastname     string
	Organization string
	Language     string
	Domain       string
}

var signupRepository *SignupRepository
var signupRepositoryOnce sync.Once

func GetSignupRepository() *SignupRepository {
	signupRepositoryOnce.Do(func() {
		signupRepository = &SignupRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS signups (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"date TIMESTAMP NOT NULL, " +
			"email VARCHAR NOT NULL, " +
			"password VARCHAR NOT NULL, " +
			"firstname VARCHAR NOT NULL, " +
			"lastname VARCHAR NOT NULL, " +
			"organization VARCHAR NOT NULL, " +
			"domain VARCHAR NOT NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
	})
	return signupRepository
}

func (r *SignupRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 6 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE signups " +
			"ADD COLUMN country VARCHAR, " +
			"ADD COLUMN language VARCHAR"); err != nil {
			panic(err)
		}
	}
	if curVersion < 15 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE signups " +
			"DROP COLUMN country"); err != nil {
			panic(err)
		}
	}
}

func (r *SignupRepository) Create(e *Signup) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO signups "+
		"(date, email, password, firstname, lastname, organization, language, domain) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) "+
		"RETURNING id",
		e.Date, e.Email, e.Password, e.Firstname, e.Lastname, e.Organization, e.Language, e.Domain).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *SignupRepository) GetOne(id string) (*Signup, error) {
	e := &Signup{}
	err := GetDatabase().DB().QueryRow("SELECT id, date, email, password, firstname, lastname, organization, language, domain "+
		"FROM signups "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.Date, &e.Email, &e.Password, &e.Firstname, &e.Lastname, &e.Organization, &e.Language, &e.Domain)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *SignupRepository) GetByEmail(email string) (*Signup, error) {
	e := &Signup{}
	err := GetDatabase().DB().QueryRow("SELECT id, date, email, password, firstname, lastname, organization, language, domain "+
		"FROM signups "+
		"WHERE LOWER(email) = $1",
		strings.ToLower(email)).Scan(&e.ID, &e.Date, &e.Email, &e.Password, &e.Firstname, &e.Lastname, &e.Organization, &e.Language, &e.Domain)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *SignupRepository) Delete(e *Signup) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM signups WHERE id = $1", e.ID)
	return err
}

func (r *SignupRepository) DeleteExpired() error {
	now := time.Now().Add(time.Hour * -2)
	_, err := GetDatabase().DB().Exec("DELETE FROM signups WHERE date < $1", now)
	return err
}
