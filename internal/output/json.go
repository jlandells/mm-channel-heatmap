package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jlandells/mm-channel-heatmap/internal/analysis"
	"github.com/jlandells/mm-channel-heatmap/internal/config"
)

type jsonOutput struct {
	Parameters jsonParameters    `json:"parameters"`
	Summary    jsonSummary       `json:"summary"`
	Channels   []jsonChannel     `json:"channels"`
	Notice     string            `json:"notice,omitempty"`
}

type jsonParameters struct {
	ServerURL      string  `json:"server_url"`
	Team           string  `json:"team,omitempty"`
	Days           int     `json:"days"`
	Since          string  `json:"since"`
	IncludePrivate bool    `json:"include_private"`
	ThresholdDead  float64 `json:"threshold_dead"`
	ThresholdVLow  float64 `json:"threshold_very_low"`
	ThresholdLow   float64 `json:"threshold_low"`
	ThresholdAct   float64 `json:"threshold_active"`
}

type jsonSummary struct {
	TotalChannels int                 `json:"total_channels"`
	Distribution  map[string]int      `json:"distribution"`
}

type jsonChannel struct {
	TeamName           string  `json:"team_name"`
	ChannelName        string  `json:"channel_name"`
	ChannelDisplayName string  `json:"channel_display_name"`
	ChannelType        string  `json:"channel_type"`
	CreatedAt          string  `json:"created_at"`
	MemberCount        int     `json:"member_count"`
	PostCount          int     `json:"post_count"`
	AvgPostsPerDay     float64 `json:"avg_posts_per_day"`
	ActivityRating     int     `json:"activity_rating"`
	ActivityLabel      string  `json:"activity_label"`
}

// WriteJSON writes channel results as structured JSON.
func WriteJSON(w io.Writer, cfg *config.Config, results *analysis.Results) error {
	dist := make(map[string]int)
	for r, count := range results.Distribution {
		dist[r.Label()] = count
	}

	channels := make([]jsonChannel, 0, len(results.Channels))
	for _, ch := range results.Channels {
		channels = append(channels, jsonChannel{
			TeamName:           ch.TeamName,
			ChannelName:        ch.ChannelName,
			ChannelDisplayName: ch.ChannelDisplayName,
			ChannelType:        channelTypeName(ch.ChannelType),
			CreatedAt:          ch.CreatedAt.Format("2006-01-02"),
			MemberCount:        ch.MemberCount,
			PostCount:          ch.PostCount,
			AvgPostsPerDay:     ch.PostsPerDay,
			ActivityRating:     int(ch.Rating),
			ActivityLabel:      ch.RatingLabel,
		})
	}

	out := jsonOutput{
		Parameters: jsonParameters{
			ServerURL:      cfg.URL,
			Team:           results.TeamScoped,
			Days:           results.Days,
			Since:          results.Since.Format("2006-01-02"),
			IncludePrivate: cfg.IncludePrivate,
			ThresholdDead:  cfg.ThresholdDead,
			ThresholdVLow:  cfg.ThresholdVeryLow,
			ThresholdLow:   cfg.ThresholdLow,
			ThresholdAct:   cfg.ThresholdActive,
		},
		Summary: jsonSummary{
			TotalChannels: len(results.Channels),
			Distribution:  dist,
		},
		Channels: channels,
	}

	if cfg.IncludePrivate {
		out.Notice = "This report includes private channel names. Handle with care."
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
