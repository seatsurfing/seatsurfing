package measure

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	perfconfig "github.com/seatsurfing/seatsurfing/server/perftest/config"
)

// ScenarioResult holds every timed sample for one scenario plus the
// gating outcome once compared against a threshold.
type ScenarioResult struct {
	ScenarioID     string          `json:"scenario"`
	Samples        int             `json:"samples"`
	Errors         int             `json:"errors"`
	MinMs          float64         `json:"minMs"`
	AvgMs          float64         `json:"avgMs"`
	P50Ms          float64         `json:"p50Ms"`
	P95Ms          float64         `json:"p95Ms"`
	P99Ms          float64         `json:"p99Ms"`
	MaxMs          float64         `json:"maxMs"`
	ThresholdP95Ms int             `json:"thresholdP95Ms,omitempty"`
	ThresholdP99Ms int             `json:"thresholdP99Ms,omitempty"`
	Gated          bool            `json:"gated"`
	Pass           bool            `json:"pass"`
	durations      []time.Duration `json:"-"`
}

func newScenarioResult(id string, durations []time.Duration, errCount int) ScenarioResult {
	r := ScenarioResult{ScenarioID: id, Samples: len(durations), Errors: errCount, durations: durations}
	if len(durations) == 0 {
		return r
	}
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	toMs := func(d time.Duration) float64 { return float64(d.Microseconds()) / 1000.0 }
	var sum time.Duration
	for _, d := range sorted {
		sum += d
	}
	r.MinMs = toMs(sorted[0])
	r.MaxMs = toMs(sorted[len(sorted)-1])
	r.AvgMs = toMs(sum) / float64(len(sorted))
	r.P50Ms = toMs(percentile(sorted, 50))
	r.P95Ms = toMs(percentile(sorted, 95))
	r.P99Ms = toMs(percentile(sorted, 99))
	return r
}

// percentile returns the value at percentile p (0-100) of an
// already-sorted slice using nearest-rank interpolation.
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p / 100.0)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func (r *ScenarioResult) applyThreshold(t perfconfig.ScenarioThreshold, gated bool) {
	r.Gated = gated
	r.ThresholdP95Ms = t.P95Ms
	r.ThresholdP99Ms = t.P99Ms
	if !gated {
		r.Pass = true
		return
	}
	r.Pass = true
	if t.P95Ms > 0 && r.P95Ms > float64(t.P95Ms) {
		r.Pass = false
	}
	if t.P99Ms > 0 && r.P99Ms > float64(t.P99Ms) {
		r.Pass = false
	}
	if r.Errors > 0 {
		r.Pass = false
	}
}

// Report prints a human-readable table to stdout, optionally writes a
// JSON report to jsonOutPath, and returns false if any gated scenario
// failed its threshold.
func Report(results []ScenarioResult, jsonOutPath string) (allPass bool, err error) {
	allPass = true
	fmt.Printf("%-28s %8s %8s %8s %8s %8s %10s %6s\n", "SCENARIO", "SAMPLES", "P50(ms)", "P95(ms)", "P99(ms)", "MAX(ms)", "THRESHOLD", "RESULT")
	for _, r := range results {
		status := "-"
		if r.Gated {
			if r.Pass {
				status = "PASS"
			} else {
				status = "FAIL"
				allPass = false
			}
		}
		threshold := "-"
		if r.Gated {
			threshold = fmt.Sprintf("p95<=%d", r.ThresholdP95Ms)
			if r.ThresholdP99Ms > 0 {
				threshold += fmt.Sprintf(",p99<=%d", r.ThresholdP99Ms)
			}
		}
		fmt.Printf("%-28s %8d %8.1f %8.1f %8.1f %8.1f %10s %6s\n",
			r.ScenarioID, r.Samples, r.P50Ms, r.P95Ms, r.P99Ms, r.MaxMs, threshold, status)
		if r.Errors > 0 {
			fmt.Printf("  %d/%d requests errored\n", r.Errors, r.Samples+r.Errors)
		}
	}

	gatedCount, passCount := 0, 0
	for _, r := range results {
		if r.Gated {
			gatedCount++
			if r.Pass {
				passCount++
			}
		}
	}
	fmt.Printf("\nSUMMARY: %d/%d gated scenarios within threshold\n", passCount, gatedCount)

	if jsonOutPath != "" {
		f, ferr := os.Create(jsonOutPath)
		if ferr != nil {
			return allPass, ferr
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if ferr := enc.Encode(results); ferr != nil {
			return allPass, ferr
		}
	}

	return allPass, nil
}
