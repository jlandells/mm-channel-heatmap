package output

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/jlandells/mm-channel-heatmap/internal/analysis"
)

// WriteCSV writes channel results as CSV.
func WriteCSV(w io.Writer, results *analysis.Results) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{
		"team_name",
		"channel_name",
		"channel_display_name",
		"channel_type",
		"created_at",
		"member_count",
		"post_count",
		"avg_posts_per_day",
		"activity_rating",
		"activity_label",
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, ch := range results.Channels {
		row := []string{
			ch.TeamName,
			ch.ChannelName,
			ch.ChannelDisplayName,
			channelTypeName(ch.ChannelType),
			ch.CreatedAt.Format("2006-01-02"),
			fmt.Sprintf("%d", ch.MemberCount),
			fmt.Sprintf("%d", ch.PostCount),
			fmt.Sprintf("%.2f", ch.PostsPerDay),
			fmt.Sprintf("%d", ch.Rating),
			ch.RatingLabel,
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
