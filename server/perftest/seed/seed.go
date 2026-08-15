// Package seed bulk-loads a large, realistic dataset directly into
// Postgres (bypassing the REST API) so the perftest/measure tool can
// exercise REST endpoints against production-scale data volumes.
package seed

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	. "github.com/seatsurfing/seatsurfing/server/api"
	"github.com/seatsurfing/seatsurfing/server/repository"
)

// Config controls the scale and shape of the seeded dataset.
type Config struct {
	Orgs            int
	UsersPerOrg     int
	SpacesPerOrg    int
	LocationsPerOrg int
	BookingsPerOrg  int
	ActorsFile      string
}

// OrgActors captures the IDs a measurement run needs to build realistic,
// authenticated requests against one seeded organization.
type OrgActors struct {
	OrgIndex     int      `json:"orgIndex"`
	OrgID        string   `json:"orgId"`
	OrgAdminID   string   `json:"orgAdminId"`
	SpaceAdminID string   `json:"spaceAdminId"`
	UsersPerOrg  int      `json:"usersPerOrg"`
	LocationIDs  []string `json:"locationIds"`
	SpaceIDs     []string `json:"spaceIds"`
}

// ActorsFile is written by Run and consumed by perftest/measure. It
// intentionally does not contain JWTs -- those are short-lived, so the
// measure tool mints fresh ones for these user IDs right before it starts
// timing requests.
type ActorsFile struct {
	GeneratedAt     time.Time   `json:"generatedAt"`
	SuperAdminID    string      `json:"superAdminId"`
	SuperAdminOrgID string      `json:"superAdminOrgId"`
	Orgs            []OrgActors `json:"orgs"`
}

const (
	// popularSpaceFraction is the share of a org's spaces treated as
	// "popular" -- they receive a disproportionate share of bookings, to
	// realistically stress the (currently unindexed) space_id lookup path.
	popularSpaceFraction = 0.2
	popularSpaceShare    = 0.6

	// bookingWindowDays is the +/- range (in days) around "now" that
	// booking enter times are spread across.
	bookingWindowDays = 180
	// nearWindowDays is the +/- range that a concentrated fraction of
	// bookings falls into, so date-range-scoped queries (this/last/next
	// week, this/last month) have realistic result set sizes.
	nearWindowDays     = 14
	nearWindowFraction = 0.3
)

// Run seeds the database according to cfg and writes the actors file.
func Run(cfg Config) error {
	// Force table/index creation the same way the server does at startup,
	// so the tool also works standalone against a fresh database.
	repository.GetOrganizationRepository()
	repository.GetLocationRepository()
	repository.GetSpaceRepository()
	repository.GetUserRepository()
	repository.GetBookingRepository()
	repository.GetSettingsRepository()
	repository.GetSessionRepository()

	db := repository.GetDatabase().DB()
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := &ActorsFile{GeneratedAt: time.Now().UTC()}

	log.Printf("Seeding %d organizations...", cfg.Orgs)
	orgIDs := make([]string, cfg.Orgs)
	for i := range orgIDs {
		orgIDs[i] = uuid.New().String()
	}
	if err := seedOrganizations(db, orgIDs); err != nil {
		return fmt.Errorf("seed organizations: %w", err)
	}
	if err := seedOrgDomains(db, orgIDs); err != nil {
		return fmt.Errorf("seed org domains: %w", err)
	}
	for _, id := range orgIDs {
		if err := repository.GetSettingsRepository().InitDefaultSettingsForOrg(id); err != nil {
			return fmt.Errorf("init default settings for org %s: %w", id, err)
		}
	}

	for orgIdx, orgID := range orgIDs {
		orgStart := time.Now()
		actors := OrgActors{OrgIndex: orgIdx, OrgID: orgID, UsersPerOrg: cfg.UsersPerOrg}

		locationIDs := make([]string, cfg.LocationsPerOrg)
		for i := range locationIDs {
			locationIDs[i] = uuid.New().String()
		}
		if err := seedLocations(db, orgID, locationIDs); err != nil {
			return fmt.Errorf("seed locations for org %d: %w", orgIdx, err)
		}
		actors.LocationIDs = sampleStrings(locationIDs, 5)

		spaceIDs := make([]string, cfg.SpacesPerOrg)
		spaceLocationIdx := make([]int, cfg.SpacesPerOrg)
		for i := range spaceIDs {
			spaceIDs[i] = uuid.New().String()
			spaceLocationIdx[i] = i % cfg.LocationsPerOrg
		}
		if err := seedSpaces(db, spaceIDs, locationIDs, spaceLocationIdx); err != nil {
			return fmt.Errorf("seed spaces for org %d: %w", orgIdx, err)
		}
		actors.SpaceIDs = sampleStrings(spaceIDs, 20)

		userIDs := make([]string, cfg.UsersPerOrg)
		userRoles := make([]UserRole, cfg.UsersPerOrg)
		for i := range userIDs {
			userIDs[i] = uuid.New().String()
			userRoles[i] = UserRoleUser
		}
		if cfg.UsersPerOrg > 0 {
			userRoles[0] = UserRoleOrgAdmin
			actors.OrgAdminID = userIDs[0]
		}
		if cfg.UsersPerOrg > 1 {
			userRoles[1] = UserRoleSpaceAdmin
			actors.SpaceAdminID = userIDs[1]
		}
		// The very first org additionally gets a super admin (bumped from
		// org admin), used for the one cross-org scenario (organization
		// list).
		if orgIdx == 0 && cfg.UsersPerOrg > 0 {
			userRoles[0] = UserRoleSuperAdmin
			result.SuperAdminID = userIDs[0]
			result.SuperAdminOrgID = orgID
		}
		if err := seedUsers(db, orgID, orgIdx, userIDs, userRoles); err != nil {
			return fmt.Errorf("seed users for org %d: %w", orgIdx, err)
		}

		if err := seedBookings(db, rnd, orgIdx, userIDs, spaceIDs, cfg.BookingsPerOrg); err != nil {
			return fmt.Errorf("seed bookings for org %d: %w", orgIdx, err)
		}

		result.Orgs = append(result.Orgs, actors)
		log.Printf("Org %d/%d seeded (%d users, %d spaces, %d bookings) in %s",
			orgIdx+1, cfg.Orgs, cfg.UsersPerOrg, cfg.SpacesPerOrg, cfg.BookingsPerOrg, time.Since(orgStart))
	}

	return writeActorsFile(cfg.ActorsFile, result)
}

func sampleStrings(all []string, n int) []string {
	if len(all) <= n {
		return all
	}
	return all[:n]
}

func writeActorsFile(path string, result *ActorsFile) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func seedOrganizations(db *sql.DB, orgIDs []string) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("organizations", "id", "name", "contact_firstname", "contact_lastname", "contact_email", "language", "signup_date"))
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for i, id := range orgIDs {
		if _, err := stmt.Exec(id, fmt.Sprintf("Perf Test Org %04d", i), "Perf", "Test", fmt.Sprintf("org%04d@perf.test", i), "en", now); err != nil {
			return err
		}
	}
	if _, err := stmt.Exec(); err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}
	return txn.Commit()
}

func seedOrgDomains(db *sql.DB, orgIDs []string) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("organizations_domains", "domain", "organization_id", "active", "verify_token", "primary_domain"))
	if err != nil {
		return err
	}
	for i, id := range orgIDs {
		domain := fmt.Sprintf("org%04d.perf.test", i)
		if _, err := stmt.Exec(domain, id, true, uuid.New().String(), true); err != nil {
			return err
		}
	}
	if _, err := stmt.Exec(); err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}
	return txn.Commit()
}

func seedLocations(db *sql.DB, orgID string, locationIDs []string) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("locations", "id", "organization_id", "name", "description", "tz", "enabled", "max_concurrent_bookings", "bookable_days"))
	if err != nil {
		return err
	}
	for i, id := range locationIDs {
		if _, err := stmt.Exec(id, orgID, fmt.Sprintf("Location %02d", i), "", "UTC", true, 0, "1,2,3,4,5,6,7"); err != nil {
			return err
		}
	}
	if _, err := stmt.Exec(); err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}
	return txn.Commit()
}

func seedSpaces(db *sql.DB, spaceIDs []string, locationIDs []string, spaceLocationIdx []int) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("spaces", "id", "location_id", "name", "x", "y", "width", "height", "rotation", "require_subject", "enabled", "kiosk_enabled", "shape", "font_size"))
	if err != nil {
		return err
	}
	for i, id := range spaceIDs {
		locID := locationIDs[spaceLocationIdx[i]]
		if _, err := stmt.Exec(id, locID, fmt.Sprintf("Space %04d", i), 0, 0, 100, 100, 0, false, true, false, "rect", "normal"); err != nil {
			return err
		}
	}
	if _, err := stmt.Exec(); err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}
	return txn.Commit()
}

func seedUsers(db *sql.DB, orgID string, orgIdx int, userIDs []string, roles []UserRole) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("users", "id", "organization_id", "email", "role", "firstname", "lastname", "disabled"))
	if err != nil {
		return err
	}
	for i, id := range userIDs {
		email := fmt.Sprintf("user%06d@org%04d.perf.test", i, orgIdx)
		if _, err := stmt.Exec(id, orgID, email, int(roles[i]), "Perf", fmt.Sprintf("User%06d", i), false); err != nil {
			return err
		}
	}
	if _, err := stmt.Exec(); err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}
	return txn.Commit()
}

// seedBookings generates bookingsPerOrg rows for one organization, batched
// into chunks to bound per-transaction memory/WAL use.
func seedBookings(db *sql.DB, rnd *rand.Rand, orgIdx int, userIDs, spaceIDs []string, bookingsPerOrg int) error {
	const chunkSize = 200_000
	numPopular := int(float64(len(spaceIDs)) * popularSpaceFraction)
	if numPopular < 1 && len(spaceIDs) > 0 {
		numPopular = 1
	}

	pickSpace := func() string {
		if numPopular > 0 && rnd.Float64() < popularSpaceShare {
			return spaceIDs[rnd.Intn(numPopular)]
		}
		return spaceIDs[rnd.Intn(len(spaceIDs))]
	}
	pickEnterTime := func() time.Time {
		now := time.Now().UTC()
		var offsetDays int
		if rnd.Float64() < nearWindowFraction {
			offsetDays = rnd.Intn(2*nearWindowDays+1) - nearWindowDays
		} else {
			offsetDays = rnd.Intn(2*bookingWindowDays+1) - bookingWindowDays
		}
		offsetMinutes := rnd.Intn(24 * 60)
		return now.AddDate(0, 0, offsetDays).Truncate(24 * time.Hour).Add(time.Duration(offsetMinutes) * time.Minute)
	}

	for offset := 0; offset < bookingsPerOrg; offset += chunkSize {
		n := chunkSize
		if offset+n > bookingsPerOrg {
			n = bookingsPerOrg - offset
		}
		if err := seedBookingChunk(db, rnd, userIDs, n, pickSpace, pickEnterTime); err != nil {
			return err
		}
	}
	return nil
}

func seedBookingChunk(db *sql.DB, rnd *rand.Rand, userIDs []string, n int, pickSpace func() string, pickEnterTime func() time.Time) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("bookings", "id", "user_id", "space_id", "enter_time", "leave_time", "approved", "subject"))
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		enter := pickEnterTime()
		durationMinutes := 30 + rnd.Intn(8*60-30)
		leave := enter.Add(time.Duration(durationMinutes) * time.Minute)
		approved := rnd.Float64() < 0.9
		userID := userIDs[rnd.Intn(len(userIDs))]
		if _, err := stmt.Exec(uuid.New().String(), userID, pickSpace(), enter, leave, approved, ""); err != nil {
			return err
		}
	}
	if _, err := stmt.Exec(); err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}
	return txn.Commit()
}
