// Package config loads the max-response-time thresholds used to gate the
// performance test run.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ScenarioThreshold defines the max acceptable response time (in
// milliseconds) for a single measured scenario. A zero value means "not
// gated" for that percentile.
type ScenarioThreshold struct {
	P95Ms int `yaml:"p95_ms"`
	P99Ms int `yaml:"p99_ms"`
}

// ThresholdConfig is the top-level shape of the thresholds YAML file.
type ThresholdConfig struct {
	Scenarios map[string]ScenarioThreshold `yaml:"scenarios"`
}

// Load reads and parses a thresholds YAML file from disk.
func Load(path string) (*ThresholdConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &ThresholdConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Scenarios == nil {
		cfg.Scenarios = map[string]ScenarioThreshold{}
	}
	return cfg, nil
}

// For returns the threshold configured for a scenario ID and whether one
// was configured at all. A scenario without a configured threshold is
// measured but not gated.
func (c *ThresholdConfig) For(scenarioID string) (ScenarioThreshold, bool) {
	t, ok := c.Scenarios[scenarioID]
	return t, ok
}
