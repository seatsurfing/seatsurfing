package measure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"syscall"
	"time"

	perfconfig "github.com/seatsurfing/seatsurfing/server/perftest/config"
)

// transportErrorReason collapses a client-side failure into a short,
// aggregatable label. The distinction that matters most here is "the server
// took too long" versus "the server was not reachable", since those point at
// very different problems.
func transportErrorReason(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return "client timeout"
		}
		err = urlErr.Err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "client timeout"
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return "connection refused"
	}
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "connection reset"
	}
	return "transport error: " + err.Error()
}

func totalCount(counts map[string]int) int {
	total := 0
	for _, n := range counts {
		total += n
	}
	return total
}

// formatReasons renders a reason histogram as a stable, most-frequent-first
// string, e.g. `HTTP 401 x48, client timeout x2`.
func formatReasons(counts map[string]int) string {
	type kv struct {
		reason string
		count  int
	}
	pairs := make([]kv, 0, len(counts))
	for r, n := range counts {
		pairs = append(pairs, kv{r, n})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].reason < pairs[j].reason
	})
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, fmt.Sprintf("%s x%d", p.reason, p.count))
	}
	return strings.Join(parts, ", ")
}

// RunConfig controls one measurement run.
type RunConfig struct {
	BaseURL    string
	Warmup     int
	Iterations int
	Timeout    time.Duration
	// MaxDuration, when non-zero, aborts the run once the elapsed wall-clock
	// time exceeds it. Scenarios that have not run yet are reported as
	// skipped rather than silently producing a wall of failures.
	MaxDuration time.Duration
	Thresholds  *perfconfig.ThresholdConfig
	JSONOutPath string
}

// Run executes every scenario in Scenarios against baseURL, using the
// supplied actor pools, and returns per-scenario results plus whether all
// gated scenarios passed their threshold.
func Run(cfg RunConfig, orgAdmins, spaceAdmins []Actor) (bool, []ScenarioResult, error) {
	if len(orgAdmins) == 0 || len(spaceAdmins) == 0 {
		return false, nil, fmt.Errorf("no actors available to run scenarios")
	}

	client := &http.Client{Timeout: cfg.Timeout}
	rnd := rand.New(rand.NewSource(1))

	// attempt issues one request and reports why it failed, if it did. The
	// reason is deliberately coarse (an HTTP status or a transport error
	// class) so that it aggregates into a small histogram rather than 50
	// unique strings.
	attempt := func(sc Scenario, actor *Actor) (elapsed time.Duration, reason string) {
		req, err := sc.Request(cfg.BaseURL, actor, rnd)
		if err != nil {
			return 0, "request build error: " + err.Error()
		}
		start := time.Now()
		resp, err := client.Do(req)
		elapsed = time.Since(start)
		if err != nil {
			return elapsed, transportErrorReason(err)
		}
		defer resp.Body.Close()
		io.Copy(io.Discard, resp.Body)
		if resp.StatusCode >= 400 {
			return elapsed, fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return elapsed, ""
	}

	runStart := time.Now()
	var results []ScenarioResult
	for _, sc := range Scenarios {
		if cfg.MaxDuration > 0 && time.Since(runStart) > cfg.MaxDuration {
			fmt.Printf("Skipping scenario %s: run exceeded --max-duration (%s)\n", sc.ID, cfg.MaxDuration)
			skipped := ScenarioResult{ScenarioID: sc.ID, Skipped: true}
			_, skipped.Gated = cfg.Thresholds.For(sc.ID)
			results = append(results, skipped)
			continue
		}
		pool := spaceAdmins
		switch sc.ID {
		case "organization.getOne", "user.count":
			pool = orgAdmins
		}

		fmt.Printf("Running scenario %s (warmup=%d, iterations=%d)...\n", sc.ID, cfg.Warmup, cfg.Iterations)

		warmupErrors := map[string]int{}
		for i := 0; i < cfg.Warmup; i++ {
			if _, reason := attempt(sc, &pool[rnd.Intn(len(pool))]); reason != "" {
				warmupErrors[reason]++
			}
		}
		if len(warmupErrors) > 0 {
			fmt.Printf("  warmup: %d/%d requests errored (%s)\n",
				totalCount(warmupErrors), cfg.Warmup, formatReasons(warmupErrors))
		}

		durations := make([]time.Duration, 0, cfg.Iterations)
		errorsByReason := map[string]int{}
		var errDurations []time.Duration
		for i := 0; i < cfg.Iterations; i++ {
			elapsed, reason := attempt(sc, &pool[rnd.Intn(len(pool))])
			if reason != "" {
				errorsByReason[reason]++
				errDurations = append(errDurations, elapsed)
				continue
			}
			durations = append(durations, elapsed)
		}

		result := newScenarioResult(sc.ID, durations, errorsByReason, errDurations)
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
