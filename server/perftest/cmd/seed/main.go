// Command perftest-seed bulk-loads a large, realistic dataset directly
// into Postgres for the performance test suite (see server/perftest).
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/seatsurfing/seatsurfing/server/perftest/seed"
)

func main() {
	postgresURL := flag.String("postgres-url", "", "Postgres connection URL (falls back to POSTGRES_URL env var)")
	orgs := flag.Int("orgs", 100, "number of organizations to seed")
	usersPerOrg := flag.Int("users-per-org", 1000, "number of users per organization")
	spacesPerOrg := flag.Int("spaces-per-org", 500, "number of spaces per organization")
	locationsPerOrg := flag.Int("locations-per-org", 5, "number of locations per organization")
	bookingsPerOrg := flag.Int("bookings-per-org", 200_000, "number of bookings per organization")
	actorsOut := flag.String("out", "perftest-actors.json", "path to write the actors file consumed by perftest-measure")
	flag.Parse()

	if *postgresURL != "" {
		os.Setenv("POSTGRES_URL", *postgresURL)
	}

	start := time.Now()
	cfg := seed.Config{
		Orgs:            *orgs,
		UsersPerOrg:     *usersPerOrg,
		SpacesPerOrg:    *spacesPerOrg,
		LocationsPerOrg: *locationsPerOrg,
		BookingsPerOrg:  *bookingsPerOrg,
		ActorsFile:      *actorsOut,
	}
	log.Printf("Seeding: %d orgs, %d users/org, %d spaces/org, %d bookings/org (~%d bookings total)",
		cfg.Orgs, cfg.UsersPerOrg, cfg.SpacesPerOrg, cfg.BookingsPerOrg, cfg.Orgs*cfg.BookingsPerOrg)

	if err := seed.Run(cfg); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}
	log.Printf("Seeding completed in %s. Actors written to %s", time.Since(start), *actorsOut)
}
