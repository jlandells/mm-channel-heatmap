package output

import (
	"bytes"
	"encoding/csv"
	"testing"
	"time"

	"github.com/jlandells/mm-channel-heatmap/internal/analysis"
)

func TestWriteCSV_HeaderAndRows(t *testing.T) {
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
			{
				TeamName:           "Engineering",
				ChannelName:        "dead-channel",
				ChannelDisplayName: "Dead Channel",
				ChannelType:        "P",
				CreatedAt:          time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
				MemberCount:        5,
				PostCount:          0,
				PostsPerDay:        0,
				Rating:             analysis.RatingDead,
				RatingLabel:        "Dead",
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteCSV(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 3 { // header + 2 rows
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	// Check header
	expectedHeader := []string{
		"team_name", "channel_name", "channel_display_name", "channel_type",
		"created_at", "member_count", "post_count", "avg_posts_per_day",
		"activity_rating", "activity_label",
	}
	for i, h := range expectedHeader {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// Check first data row
	if records[1][0] != "Engineering" {
		t.Errorf("row 1 team = %q, want %q", records[1][0], "Engineering")
	}
	if records[1][1] != "general" {
		t.Errorf("row 1 channel = %q, want %q", records[1][1], "general")
	}
	if records[1][3] != "Public" {
		t.Errorf("row 1 type = %q, want %q", records[1][3], "Public")
	}
}

func TestWriteCSV_EmptyResults(t *testing.T) {
	results := &analysis.Results{
		Channels: []analysis.ChannelResult{},
	}

	var buf bytes.Buffer
	if err := WriteCSV(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 1 { // header only
		t.Fatalf("expected 1 record (header), got %d", len(records))
	}
}
