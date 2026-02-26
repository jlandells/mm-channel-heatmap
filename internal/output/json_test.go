package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/jlandells/mm-channel-heatmap/internal/analysis"
	"github.com/jlandells/mm-channel-heatmap/internal/config"
)

func TestWriteJSON_ValidStructure(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.URL = "https://mm.example.com"

	results := &analysis.Results{
		Channels: []analysis.ChannelResult{
			{
				TeamName:           "Engineering",
				ChannelName:        "general",
				ChannelDisplayName: "General",
				ChannelType:        "O",
				CreatedAt:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				MemberCount:        50,
				PostCount:          100,
				PostsPerDay:        3.33,
				Rating:             analysis.RatingActive,
				RatingLabel:        "Active",
			},
		},
		Since:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Days:         30,
		Distribution: map[analysis.Rating]int{analysis.RatingActive: 1},
	}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, &cfg, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, buf.String())
	}

	// Check top-level keys
	for _, key := range []string{"parameters", "summary", "channels"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	// Check channels array
	channels, ok := parsed["channels"].([]interface{})
	if !ok {
		t.Fatal("channels is not an array")
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}

	ch := channels[0].(map[string]interface{})
	if ch["channel_name"] != "general" {
		t.Errorf("channel_name = %v, want %q", ch["channel_name"], "general")
	}
}

func TestWriteJSON_PrivateNotice(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.URL = "https://mm.example.com"
	cfg.IncludePrivate = true

	results := &analysis.Results{
		Channels:     []analysis.ChannelResult{},
		Since:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Days:         30,
		Distribution: map[analysis.Rating]int{},
	}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, &cfg, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	notice, ok := parsed["notice"].(string)
	if !ok || notice == "" {
		t.Error("expected non-empty notice when include-private is set")
	}
}

func TestWriteJSON_NoNoticeWhenPublicOnly(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.URL = "https://mm.example.com"
	cfg.IncludePrivate = false

	results := &analysis.Results{
		Channels:     []analysis.ChannelResult{},
		Since:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Days:         30,
		Distribution: map[analysis.Rating]int{},
	}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, &cfg, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(buf.Bytes(), &parsed)

	if _, ok := parsed["notice"]; ok {
		t.Error("did not expect notice when include-private is false")
	}
}
