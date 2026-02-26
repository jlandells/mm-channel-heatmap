package output

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/jlandells/mm-channel-heatmap/internal/analysis"
	"github.com/jlandells/mm-channel-heatmap/internal/config"
)

func ratingEmoji(r analysis.Rating) string {
	switch r {
	case analysis.RatingDead:
		return "\U0001F480" // 💀
	case analysis.RatingVeryLow:
		return "\U0001F534" // 🔴
	case analysis.RatingLow:
		return "\U0001F7E1" // 🟡
	case analysis.RatingActive:
		return "\U0001F7E2" // 🟢
	case analysis.RatingVeryHigh:
		return "\U0001F535" // 🔵
	default:
		return "?"
	}
}

// WriteTable renders the multi-section table output.
func WriteTable(w io.Writer, cfg *config.Config, results *analysis.Results, showFull bool) error {
	// Section 1: Summary
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "=== Channel Activity Report ===")
	fmt.Fprintln(w, "")

	fmt.Fprintf(w, "  Server:         %s\n", cfg.URL)
	if results.TeamScoped != "" {
		fmt.Fprintf(w, "  Team:           %s\n", results.TeamScoped)
	} else {
		fmt.Fprintf(w, "  Team:           (all teams)\n")
	}
	fmt.Fprintf(w, "  Period:         last %d days (since %s)\n", results.Days, results.Since.Format("2006-01-02"))
	fmt.Fprintf(w, "  Total channels: %d\n", len(results.Channels))
	if cfg.IncludePrivate {
		fmt.Fprintf(w, "  Scope:          public + private channels\n")
	} else {
		fmt.Fprintf(w, "  Scope:          public channels only\n")
	}
	fmt.Fprintln(w, "")

	// Activity distribution
	fmt.Fprintln(w, "  Activity Distribution:")
	ratings := []analysis.Rating{
		analysis.RatingVeryHigh,
		analysis.RatingActive,
		analysis.RatingLow,
		analysis.RatingVeryLow,
		analysis.RatingDead,
	}
	for _, r := range ratings {
		count := results.Distribution[r]
		fmt.Fprintf(w, "    %s %-12s %d\n", ratingEmoji(r), r.Label(), count)
	}
	fmt.Fprintln(w, "")

	// Top 5 most active
	if len(results.Top5) > 0 {
		fmt.Fprintln(w, "  Top 5 Most Active Channels:")
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "    %s\t%s\t%s\t%s\t%s\t%s\n", "#", "TEAM", "CHANNEL", "POSTS", "POSTS/DAY", "RATING")
		for i, ch := range results.Top5 {
			fmt.Fprintf(tw, "    %d\t%s\t%s\t%d\t%.1f\t%s %s\n",
				i+1, ch.TeamName, channelLabel(ch), ch.PostCount, ch.PostsPerDay, ratingEmoji(ch.Rating), ch.RatingLabel)
		}
		tw.Flush()
		fmt.Fprintln(w, "")
	}

	// Bottom 5 least active
	if len(results.Bottom5) > 0 {
		fmt.Fprintln(w, "  Bottom 5 Least Active Channels:")
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "    %s\t%s\t%s\t%s\t%s\t%s\n", "#", "TEAM", "CHANNEL", "POSTS", "POSTS/DAY", "MEMBERS")
		for i, ch := range results.Bottom5 {
			fmt.Fprintf(tw, "    %d\t%s\t%s\t%d\t%.1f\t%d\n",
				i+1, ch.TeamName, channelLabel(ch), ch.PostCount, ch.PostsPerDay, ch.MemberCount)
		}
		tw.Flush()
		fmt.Fprintln(w, "")
	}

	// Section 2: Zero-activity channels
	deadChannels := filterByRating(results.Channels, analysis.RatingDead)
	if len(deadChannels) > 0 {
		fmt.Fprintln(w, "=== Zero-Activity Channels ===")
		fmt.Fprintln(w, "")
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\n", "TEAM", "CHANNEL", "TYPE", "CREATED", "MEMBERS")
		for _, ch := range deadChannels {
			fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%d\n",
				ch.TeamName, channelLabel(ch), channelTypeName(ch.ChannelType), ch.CreatedAt.Format("2006-01-02"), ch.MemberCount)
		}
		tw.Flush()
		fmt.Fprintln(w, "")
	}

	// Section 3: Full channel list (only with --full or --output)
	if showFull && len(results.Channels) > 0 {
		fmt.Fprintln(w, "=== Full Channel List ===")
		fmt.Fprintln(w, "")

		// Group by team
		teamGroups := groupByTeam(results.Channels)
		for _, group := range teamGroups {
			fmt.Fprintf(w, "  Team: %s\n", group.teamName)
			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "    %s\t%s\t%s\t%s\t%s\t%s\n", "CHANNEL", "TYPE", "POSTS", "POSTS/DAY", "MEMBERS", "RATING")
			for _, ch := range group.channels {
				fmt.Fprintf(tw, "    %s\t%s\t%d\t%.1f\t%d\t%s %s\n",
					channelLabel(ch), channelTypeName(ch.ChannelType), ch.PostCount, ch.PostsPerDay, ch.MemberCount, ratingEmoji(ch.Rating), ch.RatingLabel)
			}
			tw.Flush()
			fmt.Fprintln(w, "")
		}
	}

	return nil
}

func channelLabel(ch analysis.ChannelResult) string {
	if ch.ChannelDisplayName != "" && ch.ChannelDisplayName != ch.ChannelName {
		return ch.ChannelDisplayName
	}
	return ch.ChannelName
}

func channelTypeName(t string) string {
	switch t {
	case "O":
		return "Public"
	case "P":
		return "Private"
	default:
		return t
	}
}

func filterByRating(channels []analysis.ChannelResult, rating analysis.Rating) []analysis.ChannelResult {
	var out []analysis.ChannelResult
	for _, ch := range channels {
		if ch.Rating == rating {
			out = append(out, ch)
		}
	}
	return out
}

type teamGroup struct {
	teamName string
	channels []analysis.ChannelResult
}

func groupByTeam(channels []analysis.ChannelResult) []teamGroup {
	order := make([]string, 0)
	groups := make(map[string][]analysis.ChannelResult)

	for _, ch := range channels {
		if _, exists := groups[ch.TeamName]; !exists {
			order = append(order, ch.TeamName)
		}
		groups[ch.TeamName] = append(groups[ch.TeamName], ch)
	}

	result := make([]teamGroup, 0, len(order))
	for _, name := range order {
		result = append(result, teamGroup{teamName: name, channels: groups[name]})
	}
	return result
}
