// Command perftest-measure issues real HTTP requests against a running
// seatsurfing server instance, times REST API endpoints under the data
// volume seeded by perftest-seed, and fails (non-zero exit) if any gated
// scenario exceeds its configured max response time (see
// server/perftest/testdata/thresholds.yaml).
package main

import (
	"flag"
	"log"
	"os"
	"time"

	perfconfig "github.com/seatsurfing/seatsurfing/server/perftest/config"
	"github.com/seatsurfing/seatsurfing/server/perftest/measure"
)

func main() {
	postgresURL := flag.String("postgres-url", "", "Postgres connection URL, used only to mint actor JWTs (falls back to POSTGRES_URL env var)")
	baseURL := flag.String("base-url", "http://localhost:8080", "base URL of the running server")
	actorsFile := flag.String("actors-file", "perftest-actors.json", "path to the actors file written by perftest-seed")
	thresholdsPath := flag.String("thresholds", "perftest/testdata/thresholds.yaml", "path to the thresholds YAML file")
	warmup := flag.Int("warmup", 5, "untimed warmup requests per scenario")
	iterations := flag.Int("iterations", 50, "timed requests per scenario")
	jsonOut := flag.String("json-out", "perftest-results.json", "path to write the machine-readable JSON report")
	timeout := flag.Duration("timeout", 30*time.Second, "per-request HTTP timeout")
	flag.Parse()

	if *postgresURL != "" {
		os.Setenv("POSTGRES_URL", *postgresURL)
	}

	thresholds, err := perfconfig.Load(*thresholdsPath)
	if err != nil {
		log.Fatalf("Failed to load thresholds from %s: %v", *thresholdsPath, err)
	}

	log.Printf("Loading actors and minting fresh access tokens from %s...", *actorsFile)
	orgAdmins, spaceAdmins, superAdmin, err := measure.LoadActors(*actorsFile)
	if err != nil {
		log.Fatalf("Failed to load actors: %v", err)
	}
	log.Printf("Loaded %d org-admin actors, %d space-admin actors", len(orgAdmins), len(spaceAdmins))

	cfg := measure.RunConfig{
		BaseURL:     *baseURL,
		Warmup:      *warmup,
		Iterations:  *iterations,
		Timeout:     *timeout,
		Thresholds:  thresholds,
		JSONOutPath: *jsonOut,
	}

	allPass, _, err := measure.Run(cfg, orgAdmins, spaceAdmins, *superAdmin)
	if err != nil {
		log.Fatalf("Measurement run failed: %v", err)
	}
	if !allPass {
		log.Println("One or more scenarios exceeded their configured response-time threshold.")
		os.Exit(1)
	}
	log.Println("All gated scenarios are within their configured response-time threshold.")
}
