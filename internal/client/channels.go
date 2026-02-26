package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// FetchChannelsForTeam returns public (and optionally private) channels for a team.
// DMs, GMs, and archived channels are excluded.
func (c *Client) FetchChannelsForTeam(teamID string, includePrivate bool) ([]Channel, error) {
	var channels []Channel

	// Fetch public channels
	err := c.paginateGet(fmt.Sprintf("/teams/%s/channels", teamID), 200, func(body io.Reader) (int, error) {
		var page []Channel
		if err := json.NewDecoder(body).Decode(&page); err != nil {
			return 0, err
		}
		for _, ch := range page {
			if shouldIncludeChannel(ch) {
				channels = append(channels, ch)
			}
		}
		return len(page), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public channels for team %s: %w", teamID, err)
	}

	if includePrivate {
		privErr := c.paginateGet(fmt.Sprintf("/teams/%s/channels/private", teamID), 200, func(body io.Reader) (int, error) {
			var page []Channel
			if err := json.NewDecoder(body).Decode(&page); err != nil {
				return 0, err
			}
			for _, ch := range page {
				if shouldIncludeChannel(ch) {
					channels = append(channels, ch)
				}
			}
			return len(page), nil
		})
		if privErr != nil {
			// 403 on private channels is non-fatal — warn and continue
			fmt.Fprintf(os.Stderr, "WARNING: could not fetch private channels for team %s: %v\n", teamID, privErr)
		}
	}

	return channels, nil
}

// FetchChannelStats returns the stats (including member count) for a channel.
func (c *Client) FetchChannelStats(channelID string) (*ChannelStats, error) {
	var stats ChannelStats
	_, err := c.doJSON(http.MethodGet, fmt.Sprintf("/channels/%s/stats", channelID), &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats for channel %s: %w", channelID, err)
	}
	return &stats, nil
}

func shouldIncludeChannel(ch Channel) bool {
	// Exclude DMs and GMs
	if ch.Type == "D" || ch.Type == "G" {
		return false
	}
	// Exclude archived channels
	if ch.DeleteAt != 0 {
		return false
	}
	return true
}
