package analysis

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jlandells/mm-channel-heatmap/internal/client"
	"github.com/jlandells/mm-channel-heatmap/internal/config"
	"github.com/jlandells/mm-channel-heatmap/internal/progress"
)

// Rating represents a channel activity tier (1–5).
type Rating int

const (
	RatingDead     Rating = 1
	RatingVeryLow  Rating = 2
	RatingLow      Rating = 3
	RatingActive   Rating = 4
	RatingVeryHigh Rating = 5
)

// RatingLabel returns a human-readable label for the rating.
func (r Rating) Label() string {
	switch r {
	case RatingDead:
		return "Dead"
	case RatingVeryLow:
		return "Very Low"
	case RatingLow:
		return "Low"
	case RatingActive:
		return "Active"
	case RatingVeryHigh:
		return "Very Active"
	default:
		return "Unknown"
	}
}

// ChannelResult holds the analysis results for a single channel.
type ChannelResult struct {
	TeamName           string
	ChannelName        string
	ChannelDisplayName string
	ChannelType        string
	CreatedAt          time.Time
	MemberCount        int
	PostCount          int
	PostsPerDay        float64
	Rating             Rating
	RatingLabel        string
}

// Results holds all analysis output.
type Results struct {
	Channels    []ChannelResult
	Top5        []ChannelResult
	Bottom5     []ChannelResult
	Since       time.Time
	Days        int
	TeamScoped  string
	Distribution map[Rating]int
}

// AssignRating determines the activity tier for a given posts-per-day value.
func AssignRating(postsPerDay float64, cfg *config.Config) Rating {
	if postsPerDay <= cfg.ThresholdDead {
		return RatingDead
	}
	if postsPerDay <= cfg.ThresholdVeryLow {
		return RatingVeryLow
	}
	if postsPerDay <= cfg.ThresholdLow {
		return RatingLow
	}
	if postsPerDay <= cfg.ThresholdActive {
		return RatingActive
	}
	return RatingVeryHigh
}

// Run performs the full analysis: fetches teams/channels, counts posts concurrently,
// and returns sorted, rated results.
func Run(cfg *config.Config, mc *client.Client, prog *progress.Progress) (*Results, error) {
	since := time.Now().AddDate(0, 0, -cfg.Days)
	days := float64(cfg.Days)

	// Resolve teams
	var teams []client.Team
	if cfg.Team != "" {
		team, err := mc.ResolveTeamByName(cfg.Team)
		if err != nil {
			return nil, fmt.Errorf("could not resolve team %q: %w", cfg.Team, err)
		}
		teams = []client.Team{*team}
	} else {
		var err error
		teams, err = mc.FetchAllTeams()
		if err != nil {
			return nil, fmt.Errorf("could not fetch teams: %w", err)
		}
	}

	// Build team name map
	teamNames := make(map[string]string)
	for _, t := range teams {
		teamNames[t.ID] = t.DisplayName
		if t.DisplayName == "" {
			teamNames[t.ID] = t.Name
		}
	}

	// Fetch channels across all teams
	var allChannels []client.Channel
	for _, t := range teams {
		channels, err := mc.FetchChannelsForTeam(t.ID, cfg.IncludePrivate)
		if err != nil {
			return nil, fmt.Errorf("could not fetch channels for team %s: %w", t.Name, err)
		}
		allChannels = append(allChannels, channels...)
	}

	if len(allChannels) == 0 {
		return &Results{
			Since:        since,
			Days:         cfg.Days,
			TeamScoped:   cfg.Team,
			Distribution: make(map[Rating]int),
		}, nil
	}

	prog.SetTotal(len(allChannels))

	// Worker pool with semaphore
	type workerResult struct {
		idx    int
		result ChannelResult
		err    error
	}

	resultCh := make(chan workerResult, len(allChannels))
	sem := make(chan struct{}, cfg.Workers)
	var wg sync.WaitGroup

	for i, ch := range allChannels {
		wg.Add(1)
		go func(idx int, ch client.Channel) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			prog.Update(ch.DisplayName)

			postCount, err := mc.CountPostsInWindow(ch.ID, since)
			if err != nil {
				resultCh <- workerResult{idx: idx, err: fmt.Errorf("channel %s: %w", ch.Name, err)}
				return
			}

			stats, err := mc.FetchChannelStats(ch.ID)
			memberCount := 0
			if err == nil {
				memberCount = stats.MemberCount
			}

			postsPerDay := float64(postCount) / days
			rating := AssignRating(postsPerDay, cfg)

			resultCh <- workerResult{
				idx: idx,
				result: ChannelResult{
					TeamName:           teamNames[ch.TeamID],
					ChannelName:        ch.Name,
					ChannelDisplayName: ch.DisplayName,
					ChannelType:        ch.Type,
					CreatedAt:          time.UnixMilli(ch.CreateAt),
					MemberCount:        memberCount,
					PostCount:          postCount,
					PostsPerDay:        postsPerDay,
					Rating:             rating,
					RatingLabel:        rating.Label(),
				},
			}
		}(i, ch)
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	results := make([]ChannelResult, 0, len(allChannels))
	var firstErr error
	for wr := range resultCh {
		if wr.err != nil {
			if firstErr == nil {
				firstErr = wr.err
			}
			continue
		}
		results = append(results, wr.result)
	}

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	// Sort results
	SortResults(results, cfg.SortBy)

	// Distribution
	dist := make(map[Rating]int)
	for _, r := range results {
		dist[r.Rating]++
	}

	// Top 5 / Bottom 5
	top5 := selectTop(results, 5)
	bottom5 := selectBottom(results, 5)

	return &Results{
		Channels:     results,
		Top5:         top5,
		Bottom5:      bottom5,
		Since:        since,
		Days:         cfg.Days,
		TeamScoped:   cfg.Team,
		Distribution: dist,
	}, nil
}

// SortResults sorts channel results by the specified field.
func SortResults(results []ChannelResult, sortBy string) {
	switch sortBy {
	case "posts":
		sort.Slice(results, func(i, j int) bool {
			return results[i].PostCount > results[j].PostCount
		})
	case "name":
		sort.Slice(results, func(i, j int) bool {
			if results[i].TeamName != results[j].TeamName {
				return results[i].TeamName < results[j].TeamName
			}
			return results[i].ChannelName < results[j].ChannelName
		})
	default: // "rating"
		sort.Slice(results, func(i, j int) bool {
			if results[i].Rating != results[j].Rating {
				return results[i].Rating > results[j].Rating
			}
			return results[i].PostCount > results[j].PostCount
		})
	}
}

func selectTop(results []ChannelResult, n int) []ChannelResult {
	// Sort by posts descending for top selection
	sorted := make([]ChannelResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PostCount > sorted[j].PostCount
	})
	if len(sorted) < n {
		n = len(sorted)
	}
	return sorted[:n]
}

func selectBottom(results []ChannelResult, n int) []ChannelResult {
	// Sort by posts ascending for bottom selection
	sorted := make([]ChannelResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PostCount < sorted[j].PostCount
	})
	if len(sorted) < n {
		n = len(sorted)
	}
	return sorted[:n]
}
