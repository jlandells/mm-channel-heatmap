package cmd

import (
	"fmt"
	"os"

	"github.com/jlandells/mm-channel-heatmap/internal/analysis"
	"github.com/jlandells/mm-channel-heatmap/internal/client"
	"github.com/jlandells/mm-channel-heatmap/internal/config"
	"github.com/jlandells/mm-channel-heatmap/internal/output"
	"github.com/jlandells/mm-channel-heatmap/internal/progress"
	"github.com/spf13/cobra"
)

var (
	cfg      config.Config
	exitCode int
)

var rootCmd = &cobra.Command{
	Use:   "mm-channel-heatmap",
	Short: "Analyse Mattermost channel activity and produce tiered reports",
	Long: `mm-channel-heatmap connects to a Mattermost instance and analyses
channel post activity over a configurable time window. It produces
tiered activity reports to help administrators identify dead, low-activity,
and thriving channels.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          run,
}

func init() {
	defaults := config.DefaultConfig()

	// Connection
	rootCmd.Flags().StringVar(&cfg.URL, "url", "", "Mattermost server URL (env: MM_URL)")
	rootCmd.Flags().StringVar(&cfg.Token, "token", "", "Personal access token (env: MM_TOKEN)")
	rootCmd.Flags().StringVar(&cfg.Username, "username", "", "Username for login auth (env: MM_USERNAME)")

	// Scope
	rootCmd.Flags().StringVar(&cfg.Team, "team", "", "Limit to a specific team (by name)")
	rootCmd.Flags().BoolVar(&cfg.IncludePrivate, "include-private", false, "Include private channels (requires permission)")

	// Time window
	rootCmd.Flags().IntVar(&cfg.Days, "days", defaults.Days, "Number of days to analyse")

	// Thresholds
	rootCmd.Flags().Float64Var(&cfg.ThresholdDead, "threshold-dead", defaults.ThresholdDead, "Posts/day threshold: Dead (0 = exactly zero posts)")
	rootCmd.Flags().Float64Var(&cfg.ThresholdVeryLow, "threshold-very-low", defaults.ThresholdVeryLow, "Posts/day threshold: Very Low")
	rootCmd.Flags().Float64Var(&cfg.ThresholdLow, "threshold-low", defaults.ThresholdLow, "Posts/day threshold: Low")
	rootCmd.Flags().Float64Var(&cfg.ThresholdActive, "threshold-active", defaults.ThresholdActive, "Posts/day threshold: Active")

	// Output
	rootCmd.Flags().StringVar(&cfg.Format, "format", defaults.Format, "Output format: table, csv, json")
	rootCmd.Flags().StringVar(&cfg.Output, "output", "", "Write output to file (enables full list for table format)")
	rootCmd.Flags().BoolVar(&cfg.Full, "full", false, "Show full channel list in table output")
	rootCmd.Flags().StringVar(&cfg.SortBy, "sort-by", defaults.SortBy, "Sort channels by: rating, posts, name")
	rootCmd.Flags().BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging to stderr")

	// Concurrency
	rootCmd.Flags().IntVar(&cfg.Workers, "workers", defaults.Workers, "Number of concurrent API workers")
}

func run(cmd *cobra.Command, args []string) error {
	cfg.ResolveEnv()
	if err := cfg.Validate(); err != nil {
		return err
	}

	mc, err := client.NewClient(cfg.URL, cfg.Token, cfg.Username, cfg.Verbose)
	if err != nil {
		return err
	}

	if cfg.IncludePrivate {
		fmt.Fprintln(os.Stderr, "WARNING: --include-private is set. Private channel names will be visible in the output.")
	}

	prog := progress.New(cfg.Verbose)
	defer prog.Clear()

	results, err := analysis.Run(&cfg, mc, prog)
	if err != nil {
		exitCode = 2
		return err
	}

	prog.Clear()

	if err := output.Write(&cfg, results); err != nil {
		exitCode = 3
		return err
	}

	return nil
}

// Execute runs the root command and returns the exit code.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if exitCode == 0 {
			exitCode = 1
		}
	}
	return exitCode
}
