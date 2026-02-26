package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all CLI configuration values.
type Config struct {
	// Connection
	URL      string
	Token    string
	Username string
	// Password is never a CLI flag — resolved from env or interactive prompt

	// Scope
	Team           string
	IncludePrivate bool

	// Time window
	Days int

	// Thresholds (posts per day)
	ThresholdDead    float64
	ThresholdVeryLow float64
	ThresholdLow     float64
	ThresholdActive  float64

	// Output
	Format  string // "table", "csv", "json"
	Output  string // file path, empty = stdout
	Full    bool
	SortBy  string // "rating", "posts", "name"
	Verbose bool

	// Concurrency
	Workers int
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		Days:             30,
		ThresholdDead:    0.0,
		ThresholdVeryLow: 0.5,
		ThresholdLow:     2.0,
		ThresholdActive:  10.0,
		Format:           "table",
		SortBy:           "rating",
		Workers:          5,
	}
}

// ResolveEnv overlays environment variables onto the config.
// CLI flags take precedence (non-zero values are not overwritten).
func (c *Config) ResolveEnv() {
	if c.URL == "" {
		c.URL = os.Getenv("MM_URL")
	}
	if c.Token == "" {
		c.Token = os.Getenv("MM_TOKEN")
	}
	if c.Username == "" {
		c.Username = os.Getenv("MM_USERNAME")
	}
}

// Validate checks the config for errors and returns a descriptive message.
func (c *Config) Validate() error {
	c.URL = strings.TrimRight(c.URL, "/")

	if c.URL == "" {
		return fmt.Errorf("--url is required (or set MM_URL)")
	}

	if c.Token == "" && c.Username == "" {
		return fmt.Errorf("either --token (or MM_TOKEN) or --username (or MM_USERNAME) is required")
	}

	if c.Token != "" && c.Username != "" {
		return fmt.Errorf("--token and --username are mutually exclusive; use one or the other")
	}

	if c.Days < 1 {
		return fmt.Errorf("--days must be >= 1, got %d", c.Days)
	}

	if c.Workers < 1 {
		return fmt.Errorf("--workers must be >= 1, got %d", c.Workers)
	}

	validFormats := map[string]bool{"table": true, "csv": true, "json": true}
	if !validFormats[c.Format] {
		return fmt.Errorf("--format must be one of: table, csv, json; got %q", c.Format)
	}

	validSorts := map[string]bool{"rating": true, "posts": true, "name": true}
	if !validSorts[c.SortBy] {
		return fmt.Errorf("--sort-by must be one of: rating, posts, name; got %q", c.SortBy)
	}

	// Thresholds must be in ascending order
	if c.ThresholdDead > c.ThresholdVeryLow {
		return fmt.Errorf("threshold-dead (%.2f) must be <= threshold-very-low (%.2f)", c.ThresholdDead, c.ThresholdVeryLow)
	}
	if c.ThresholdVeryLow > c.ThresholdLow {
		return fmt.Errorf("threshold-very-low (%.2f) must be <= threshold-low (%.2f)", c.ThresholdVeryLow, c.ThresholdLow)
	}
	if c.ThresholdLow > c.ThresholdActive {
		return fmt.Errorf("threshold-low (%.2f) must be <= threshold-active (%.2f)", c.ThresholdLow, c.ThresholdActive)
	}

	return nil
}
