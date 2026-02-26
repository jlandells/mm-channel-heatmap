package analysis

import (
	"testing"

	"github.com/jlandells/mm-channel-heatmap/internal/config"
)

func defaultCfg() *config.Config {
	c := config.DefaultConfig()
	return &c
}

func TestAssignRating_ExactlyZero(t *testing.T) {
	r := AssignRating(0.0, defaultCfg())
	if r != RatingDead {
		t.Errorf("expected Dead for 0.0, got %s", r.Label())
	}
}

func TestAssignRating_JustAboveZero(t *testing.T) {
	r := AssignRating(0.01, defaultCfg())
	if r != RatingVeryLow {
		t.Errorf("expected Very Low for 0.01, got %s", r.Label())
	}
}

func TestAssignRating_ExactlyVeryLow(t *testing.T) {
	r := AssignRating(0.5, defaultCfg())
	if r != RatingVeryLow {
		t.Errorf("expected Very Low for 0.5 (boundary), got %s", r.Label())
	}
}

func TestAssignRating_JustAboveVeryLow(t *testing.T) {
	r := AssignRating(0.51, defaultCfg())
	if r != RatingLow {
		t.Errorf("expected Low for 0.51, got %s", r.Label())
	}
}

func TestAssignRating_ExactlyLow(t *testing.T) {
	r := AssignRating(2.0, defaultCfg())
	if r != RatingLow {
		t.Errorf("expected Low for 2.0 (boundary), got %s", r.Label())
	}
}

func TestAssignRating_JustAboveLow(t *testing.T) {
	r := AssignRating(2.01, defaultCfg())
	if r != RatingActive {
		t.Errorf("expected Active for 2.01, got %s", r.Label())
	}
}

func TestAssignRating_ExactlyActive(t *testing.T) {
	r := AssignRating(10.0, defaultCfg())
	if r != RatingActive {
		t.Errorf("expected Active for 10.0 (boundary), got %s", r.Label())
	}
}

func TestAssignRating_VeryActive(t *testing.T) {
	r := AssignRating(10.01, defaultCfg())
	if r != RatingVeryHigh {
		t.Errorf("expected Very Active for 10.01, got %s", r.Label())
	}
}

func TestSelectTop_FewerThan5(t *testing.T) {
	results := []ChannelResult{
		{ChannelName: "a", PostCount: 100},
		{ChannelName: "b", PostCount: 50},
	}
	top := selectTop(results, 5)
	if len(top) != 2 {
		t.Errorf("expected 2 results, got %d", len(top))
	}
	if top[0].ChannelName != "a" {
		t.Errorf("expected 'a' first, got %s", top[0].ChannelName)
	}
}

func TestSelectBottom_FewerThan5(t *testing.T) {
	results := []ChannelResult{
		{ChannelName: "a", PostCount: 100},
		{ChannelName: "b", PostCount: 50},
		{ChannelName: "c", PostCount: 10},
	}
	bottom := selectBottom(results, 5)
	if len(bottom) != 3 {
		t.Errorf("expected 3 results, got %d", len(bottom))
	}
	if bottom[0].ChannelName != "c" {
		t.Errorf("expected 'c' first (lowest), got %s", bottom[0].ChannelName)
	}
}

func TestSortResults_ByName(t *testing.T) {
	results := []ChannelResult{
		{TeamName: "B", ChannelName: "z"},
		{TeamName: "A", ChannelName: "b"},
		{TeamName: "A", ChannelName: "a"},
	}
	SortResults(results, "name")
	if results[0].ChannelName != "a" || results[0].TeamName != "A" {
		t.Errorf("expected A/a first, got %s/%s", results[0].TeamName, results[0].ChannelName)
	}
}

func TestSortResults_ByPosts(t *testing.T) {
	results := []ChannelResult{
		{ChannelName: "low", PostCount: 5},
		{ChannelName: "high", PostCount: 100},
		{ChannelName: "mid", PostCount: 50},
	}
	SortResults(results, "posts")
	if results[0].ChannelName != "high" {
		t.Errorf("expected 'high' first, got %s", results[0].ChannelName)
	}
}

func TestSortResults_ByRating(t *testing.T) {
	results := []ChannelResult{
		{ChannelName: "dead", Rating: RatingDead, PostCount: 0},
		{ChannelName: "active", Rating: RatingVeryHigh, PostCount: 500},
		{ChannelName: "low", Rating: RatingLow, PostCount: 20},
	}
	SortResults(results, "rating")
	if results[0].ChannelName != "active" {
		t.Errorf("expected 'active' first, got %s", results[0].ChannelName)
	}
	if results[2].ChannelName != "dead" {
		t.Errorf("expected 'dead' last, got %s", results[2].ChannelName)
	}
}
