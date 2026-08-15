package measure

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	perfconfig "github.com/seatsurfing/seatsurfing/server/perftest/config"
)

// RunConfig controls one measurement run.
type RunConfig struct {
	BaseURL     string
	Warmup      int
	Iterations  int
	Timeout     time.Duration
	Thresholds  *perfconfig.ThresholdConfig
	JSONOutPath string
}

// Run executes every scenario in Scenarios against baseURL, using the
// supplied actor pools, and returns per-scenario results plus whether all
// gated scenarios passed their threshold.
func Run(cfg RunConfig, orgAdmins, spaceAdmins []Actor, superAdmin Actor) (bool, []ScenarioResult, error) {
	if len(orgAdmins) == 0 || len(spaceAdmins) == 0 {
		return false, nil, fmt.Errorf("no actors available to run scenarios")
	}

	client := &http.Client{Timeout: cfg.Timeout}
	rnd := rand.New(rand.NewSource(1))

	var results []ScenarioResult
	for _, sc := range Scenarios {
		pool := spaceAdmins
		switch sc.ID {
		case "organization.getOne", "user.count":
			pool = orgAdmins
		}
		if sc.UseSuperAdmin {
			pool = []Actor{superAdmin}
		}

		fmt.Printf("Running scenario %s (warmup=%d, iterations=%d)...\n", sc.ID, cfg.Warmup, cfg.Iterations)

		for i := 0; i < cfg.Warmup; i++ {
			actor := pool[rnd.Intn(len(pool))]
			req, err := sc.Request(cfg.BaseURL, actor, rnd)
			if err != nil {
				continue
			}
			resp, err := client.Do(req)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}

		durations := make([]time.Duration, 0, cfg.Iterations)
		errCount := 0
		for i := 0; i < cfg.Iterations; i++ {
			actor := pool[rnd.Intn(len(pool))]
			req, err := sc.Request(cfg.BaseURL, actor, rnd)
			if err != nil {
				errCount++
				continue
			}
			start := time.Now()
			resp, err := client.Do(req)
			elapsed := time.Since(start)
			if err != nil {
				errCount++
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode >= 400 {
				errCount++
				continue
			}
			durations = append(durations, elapsed)
		}

		result := newScenarioResult(sc.ID, durations, errCount)
		if t, ok := cfg.Thresholds.For(sc.ID); ok {
			result.applyThreshold(t, true)
		} else {
			result.applyThreshold(perfconfig.ScenarioThreshold{}, false)
		}
		results = append(results, result)
	}

	allPass, err := Report(results, cfg.JSONOutPath)
	if err != nil {
		return false, results, err
	}
	return allPass, results, nil
}
