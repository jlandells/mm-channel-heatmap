package output

import (
	"fmt"
	"io"
	"os"

	"github.com/jlandells/mm-channel-heatmap/internal/analysis"
	"github.com/jlandells/mm-channel-heatmap/internal/config"
)

// Write dispatches output to the appropriate formatter and destination.
func Write(cfg *config.Config, results *analysis.Results) error {
	var w io.Writer = os.Stdout
	showFull := cfg.Full

	if cfg.Output != "" {
		f, err := os.Create(cfg.Output)
		if err != nil {
			return fmt.Errorf("failed to create output file %q: %w", cfg.Output, err)
		}
		defer f.Close()
		w = f
		showFull = true // always show full list when writing to file
	}

	switch cfg.Format {
	case "csv":
		return WriteCSV(w, results)
	case "json":
		return WriteJSON(w, cfg, results)
	case "table":
		return WriteTable(w, cfg, results, showFull)
	default:
		return fmt.Errorf("unknown format %q", cfg.Format)
	}
}
