package repository

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type OrganizationRepository struct {
}

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
}

var organizationRepository *OrganizationRepository
var organizationRepositoryOnce sync.Once

func GetOrganizationRepository() *OrganizationRepository {
	organizationRepositoryOnce.Do(func() {
		organizationRepository = &OrganizationRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS organizations (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"name VARCHAR NOT NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS organizations_domains (" +
			"domain VARCHAR NOT NULL, " +
			"organization_id uuid NOT NULL, " +
			"PRIMARY KEY (domain))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_organizations_domains_organization_id ON organizations_domains(organization_id)")
		if err != nil {
			panic(err)
		}
	})
	return organizationRepository
}

func (r *OrganizationRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 3 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE organizations " +
			"ADD COLUMN contact_firstname VARCHAR, " +
			"ADD COLUMN contact_lastname VARCHAR, " +
			"ADD COLUMN contact_email VARCHAR"); err != nil {
			panic(err)
		}
	}
	if curVersion < 4 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE organizations_domains " +
			"ADD COLUMN active boolean NOT NULL DEFAULT FALSE, " +
			"ADD COLUMN verify_token uuid"); err != nil {
			panic(err)
		}
	}
	if curVersion < 5 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE organizations_domains " +
			"DROP CONSTRAINT organizations_domains_pkey"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_organizations_domains_domain ON organizations_domains(domain)"); err != nil {
			panic(err)
		}
	}
	if curVersion < 6 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE organizations " +
			"ADD COLUMN country VARCHAR, " +
			"ADD COLUMN language VARCHAR"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("UPDATE organizations SET country = 'DE', language = 'de'"); err != nil {
			panic(err)
		}
	}
	if curVersion < 8 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE organizations " +
			"ADD COLUMN signup_date TIMESTAMP NOT NULL DEFAULT '2021-03-28 16:00:00'"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("ALTER TABLE organizations " +
			"ALTER COLUMN signup_date DROP DEFAULT"); err != nil {
			panic(err)
		}
	}
	if curVersion < 15 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE organizations " +
			"DROP COLUMN country"); err != nil {
			panic(err)
		}
	}
	if curVersion < 20 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE organizations_domains " +
			"ADD COLUMN primary_domain boolean NOT NULL DEFAULT FALSE"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("UPDATE organizations_domains " +
			"SET primary_domain = TRUE " +
			"WHERE domain IN (SELECT domain FROM (SELECT DISTINCT ON (organization_id) organization_id, domain FROM organizations_domains ORDER BY organization_id))"); err != nil {
			panic(err)
		}
	}
}

func (r *OrganizationRepository) Create(e *Organization) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO organizations "+
		"(name, contact_firstname, contact_lastname, contact_email, language, signup_date) "+
		"VALUES ($1, $2, $3, $4, $5, $6) "+
		"RETURNING id",
		e.Name, e.ContactFirstname, e.ContactLastname, e.ContactEmail, e.Language, e.SignupDate).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	GetSettingsRepository().InitDefaultSettingsForOrg(e.ID)
	return nil
}

func (r *OrganizationRepository) GetOneByDomain(domain string) (*Organization, error) {
	e := &Organization{}
	err := GetDatabase().DB().QueryRow("SELECT organizations.id, organizations.name, organizations.contact_firstname, organizations.contact_lastname, organizations.contact_email, organizations.language, organizations.signup_date "+
		"FROM organizations_domains "+
		"INNER JOIN organizations ON organizations.id = organizations_domains.organization_id "+
		"WHERE LOWER(organizations_domains.domain) = $1 AND organizations_domains.active = TRUE",
		strings.ToLower(domain)).Scan(&e.ID, &e.Name, &e.ContactFirstname, &e.ContactLastname, &e.ContactEmail, &e.Language, &e.SignupDate)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *OrganizationRepository) GetOne(id string) (*Organization, error) {
	e := &Organization{}
	err := GetDatabase().DB().QueryRow("SELECT id, name, contact_firstname, contact_lastname, contact_email, language, signup_date "+
		"FROM organizations "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.Name, &e.ContactFirstname, &e.ContactLastname, &e.ContactEmail, &e.Language, &e.SignupDate)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *OrganizationRepository) GetByEmail(email string) (*Organization, error) {
	e := &Organization{}
	err := GetDatabase().DB().QueryRow("SELECT id, name, contact_firstname, contact_lastname, contact_email, language, signup_date "+
		"FROM organizations "+
		"WHERE LOWER(contact_email) = $1",
		strings.ToLower(email)).Scan(&e.ID, &e.Name, &e.ContactFirstname, &e.ContactLastname, &e.ContactEmail, &e.Language, &e.SignupDate)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *OrganizationRepository) GetAll() ([]*Organization, error) {
	var result []*Organization
	rows, err := GetDatabase().DB().Query("SELECT id, name, contact_firstname, contact_lastname, contact_email, language, signup_date " +
		"FROM organizations ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Organization{}
		err = rows.Scan(&e.ID, &e.Name, &e.ContactFirstname, &e.ContactLastname, &e.ContactEmail, &e.Language, &e.SignupDate)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *OrganizationRepository) GetNumOrgs() (int, error) {
	var result int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(*) FROM organizations").Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *OrganizationRepository) GetAllIDs() ([]string, error) {
	var result []string
	rows, err := GetDatabase().DB().Query("SELECT id " +
		"FROM organizations")
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

func (r *OrganizationRepository) Update(e *Organization) error {
	_, err := GetDatabase().DB().Exec("UPDATE organizations SET "+
		"name = $1, contact_firstname = $2, contact_lastname = $3, contact_email = $4, language = $5, signup_date = $6 "+
		"WHERE id = $7",
		e.Name, e.ContactFirstname, e.ContactLastname, e.ContactEmail, e.Language, e.SignupDate, e.ID)
	return err
}

func (r *OrganizationRepository) Delete(e *Organization) error {
	if err := GetAuthProviderRepository().DeleteAll(e.ID); err != nil {
		return err
	}
	if err := GetLocationRepository().DeleteAll(e.ID); err != nil {
		return err
	}
	if err := GetSettingsRepository().DeleteAll(e.ID); err != nil {
		return err
	}
	if err := GetUserRepository().DeleteAll(e.ID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM organizations_domains WHERE organization_id = $1", e.ID)
	if err != nil {
		return err
	}
	_, err = GetDatabase().DB().Exec("DELETE FROM organizations WHERE id = $1", e.ID)
	return err
}

func (r *OrganizationRepository) GetDomain(org *Organization, domain string) (*Domain, error) {
	e := &Domain{}
	err := GetDatabase().DB().QueryRow("SELECT domain, organization_id, active, verify_token, primary_domain "+
		"FROM organizations_domains "+
		"WHERE domain = LOWER($1) AND organization_id = $2",
		strings.ToLower(domain), org.ID).Scan(&e.DomainName, &e.OrganizationID, &e.Active, &e.VerifyToken, &e.Primary)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *OrganizationRepository) AddDomain(e *Organization, domain string, active bool) error {
	verifyToken := uuid.New().String()
	_, err := GetDatabase().DB().Exec("INSERT INTO organizations_domains "+
		"(domain, organization_id, active, verify_token) "+
		"VALUES ($1, $2, $3, $4)",
		strings.ToLower(domain), e.ID, active, verifyToken)
	return err
}

func (r *OrganizationRepository) RemoveDomain(e *Organization, domain string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM organizations_domains "+
		"WHERE domain = LOWER($1) AND organization_id = $2",
		strings.ToLower(domain), e.ID)
	return err
}

func (r *OrganizationRepository) ActivateDomain(e *Organization, domain string) error {
	_, err := GetDatabase().DB().Exec("UPDATE organizations_domains "+
		"SET active = TRUE "+
		"WHERE domain = LOWER($1) AND organization_id = $2",
		strings.ToLower(domain), e.ID)
	return err
}

func (r *OrganizationRepository) SetPrimaryDomain(e *Organization, domain string) error {
	_, err := GetDatabase().DB().Exec("UPDATE organizations_domains "+
		"SET primary_domain = FALSE "+
		"WHERE organization_id = $1",
		e.ID)
	if err != nil {
		return err
	}
	_, err = GetDatabase().DB().Exec("UPDATE organizations_domains "+
		"SET primary_domain = TRUE "+
		"WHERE domain = LOWER($1) AND organization_id = $2",
		strings.ToLower(domain), e.ID)
	return err
}

func (r *OrganizationRepository) GetPrimaryDomain(e *Organization) (*Domain, error) {
	domains, err := r.GetDomains(e)
	if err != nil {
		return nil, err
	}
	for _, domain := range domains {
		if domain.Active && domain.Primary {
			return domain, nil
		}
	}
	for _, domain := range domains {
		if domain.Active {
			return domain, nil
		}
	}
	return nil, errors.New("no primary domain found")
}

func (r *OrganizationRepository) GetDomains(e *Organization) ([]*Domain, error) {
	var result []*Domain
	rows, err := GetDatabase().DB().Query("SELECT domain, organization_id, active, verify_token, primary_domain "+
		"FROM organizations_domains "+
		"WHERE organization_id = $1 "+
		"ORDER BY domain",
		e.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		domain := &Domain{}
		err = rows.Scan(&domain.DomainName, &domain.OrganizationID, &domain.Active, &domain.VerifyToken, &domain.Primary)
		if err != nil {
			return nil, err
		}
		result = append(result, domain)
	}
	return result, nil
}

func (r *OrganizationRepository) IsValidCustomDomainForOrg(email string, org *Organization) bool {
	domain := GetDomainFromEmail(email)
	if domain == "" {
		return false
	}
	domains, err := GetOrganizationRepository().GetDomains(org)
	if err != nil {
		return false
	}
	for _, curDomain := range domains {
		if strings.ToLower(curDomain.DomainName) == domain {
			if curDomain.Active {
				return true
			}
		}
	}
	return false
}

func (r *OrganizationRepository) CreateSampleData(org *Organization) error {
	location := &Location{
		OrganizationID: org.ID,
		Name:           "Sample Floor",
		Description:    "Sample Map provided by Marco Garbelini under the Creative Commons Attribution 2.0 Generic (CC BY 2.0) License: https://www.flickr.com/photos/garbelini/300134781",
		Enabled:        true,
	}
	if err := GetLocationRepository().Create(location); err != nil {
		return err
	}
	mapFile, _ := filepath.Abs(filepath.Join(GetConfig().FilesystemBasePath, "./res/floorplan.jpg"))
	mapData, err := os.ReadFile(mapFile)
	if err != nil {
		return err
	}
	locationMap := &LocationMap{
		MimeType: "jpeg",
		Width:    2047,
		Height:   802,
		Data:     mapData,
	}
	if err := GetLocationRepository().SetMap(location, locationMap); err != nil {
		return err
	}
	spaces := []*Space{
		{LocationID: location.ID, Name: "Conference 1", X: 990, Y: 76, Width: 204, Height: 70, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 1", X: 755, Y: 60, Width: 120, Height: 55, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 2", X: 843, Y: 337, Width: 108, Height: 53, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 3", X: 624, Y: 518, Width: 104, Height: 52, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 4", X: 625, Y: 571, Width: 104, Height: 52, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 5", X: 729, Y: 518, Width: 47, Height: 105, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 9", X: 896, Y: 569, Width: 51, Height: 104, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 10", X: 948, Y: 569, Width: 51, Height: 104, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 7", X: 1057, Y: 382, Width: 51, Height: 104, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 8", X: 1110, Y: 382, Width: 51, Height: 104, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 6", X: 898, Y: 390, Width: 51, Height: 104, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 11", X: 1103, Y: 570, Width: 51, Height: 104, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 12", X: 1155, Y: 570, Width: 51, Height: 104, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 13", X: 1815, Y: 353, Width: 51, Height: 104, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 14", X: 1985, Y: 435, Width: 51, Height: 104, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 15", X: 1933, Y: 541, Width: 104, Height: 52, Rotation: 0},
		{LocationID: location.ID, Name: "Desk 16", X: 1933, Y: 626, Width: 104, Height: 52, Rotation: 0},
	}
	for _, space := range spaces {
		if err := GetSpaceRepository().Create(space); err != nil {
			return err
		}
	}
	return nil
}
