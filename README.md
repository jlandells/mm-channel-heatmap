# mm-channel-heatmap

A command-line tool that analyses Mattermost channel post activity over a configurable time window and categorises every channel into one of five activity tiers — from **Dead** (zero posts) to **Very Active** — helping administrators quickly identify unused channels ripe for cleanup and understand where conversation is concentrated.

## Why you'd use it

If your Mattermost instance has hundreds of channels, manually checking which ones are still alive is impractical. Channels accumulate over time: project channels that outlive their projects, experiment channels nobody remembers, and announcement channels that went quiet months ago. `mm-channel-heatmap` scans all public channels (and optionally private ones), counts posts over the period you choose, and gives you a clear, sortable report so you can make informed decisions about channel hygiene — without clicking through every channel in the System Console.

## Installation

Download the pre-built binary for your platform from the [Releases](https://github.com/jlandells/mm-channel-heatmap/releases) page.

| Platform | Binary |
|----------|--------|
| macOS (Apple Silicon) | `mm-channel-heatmap-darwin-arm64` |
| macOS (Intel) | `mm-channel-heatmap-darwin-amd64` |
| Linux (x86-64) | `mm-channel-heatmap-linux-amd64` |
| Linux (ARM64) | `mm-channel-heatmap-linux-arm64` |

After downloading, make the binary executable:

```bash
chmod +x mm-channel-heatmap-*
```

Move it somewhere on your `PATH` (e.g. `/usr/local/bin`) for convenience.

## Authentication

The tool supports two mutually exclusive authentication methods.

### Personal Access Token (recommended)

Pass a [Personal Access Token](https://docs.mattermost.com/developer/personal-access-tokens.html) via the `--token` flag or the `MM_TOKEN` environment variable:

```bash
mm-channel-heatmap --url https://mattermost.example.com --token YOUR_PAT
```

> **Note:** Personal Access Tokens may be disabled by your Mattermost administrator. If you receive a 401 error, check with your admin or use username/password authentication instead.

### Username and password

Pass your username via `--username` (or `MM_USERNAME`). The password is read from the `MM_PASSWORD` environment variable. If `MM_PASSWORD` is not set, you will be prompted to enter it interactively:

```bash
mm-channel-heatmap --url https://mattermost.example.com --username admin
Password for admin: ▊
```

> **Note:** You must use exactly one authentication method. Supplying both `--token` and `--username` is an error.

## Usage

```
mm-channel-heatmap [flags]
```

### Flag reference

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--url` | `MM_URL` | *(required)* | Mattermost server URL |
| `--token` | `MM_TOKEN` | | Personal access token |
| `--username` | `MM_USERNAME` | | Username for login auth (password via `MM_PASSWORD` or interactive prompt) |
| `--team` | | | Limit to a specific team (by name) |
| `--include-private` | | `false` | Include private channels (requires permission) |
| `--days` | | `30` | Number of days to analyse |
| `--threshold-dead` | | `0` | Posts/day threshold: Dead (0 = exactly zero posts) |
| `--threshold-very-low` | | `0.5` | Posts/day threshold: Very Low |
| `--threshold-low` | | `2` | Posts/day threshold: Low |
| `--threshold-active` | | `10` | Posts/day threshold: Active |
| `--format` | | `table` | Output format: `table`, `csv`, `json` |
| `--output` | | *(stdout)* | Write output to file (enables full list for table format) |
| `--full` | | `false` | Show full channel list in table output |
| `--sort-by` | | `rating` | Sort channels by: `rating`, `posts`, `name` |
| `--verbose` | | `false` | Enable verbose logging to stderr |
| `--workers` | | `5` | Number of concurrent API workers |
| `--version` | | | Print version and exit |

## Examples

**Basic run with token authentication:**

```bash
mm-channel-heatmap --url https://mattermost.example.com --token YOUR_PAT
```

**Basic run with username/password authentication:**

```bash
mm-channel-heatmap --url https://mattermost.example.com --username admin
```

**Using environment variables:**

```bash
export MM_URL=https://mattermost.example.com
export MM_TOKEN=YOUR_PAT
mm-channel-heatmap
```

**CSV output to a file:**

```bash
mm-channel-heatmap --url https://mattermost.example.com --token YOUR_PAT \
  --format csv --output report.csv
```

**Scope to a single team:**

```bash
mm-channel-heatmap --url https://mattermost.example.com --token YOUR_PAT \
  --team engineering
```

**Include private channels:**

```bash
mm-channel-heatmap --url https://mattermost.example.com --token YOUR_PAT \
  --include-private
```

**Show the full channel list:**

```bash
mm-channel-heatmap --url https://mattermost.example.com --token YOUR_PAT \
  --full
```

## Output formats

### Table (default)

The table format prints a multi-section report to the terminal (or to a file with `--output`):

1. **Summary** — server URL, team scope, analysis period, total channels, and activity distribution counts.
2. **Top 5 Most Active Channels** — ranked by post count.
3. **Bottom 5 Least Active Channels** — ranked by post count ascending, including member counts.
4. **Zero-Activity Channels** — every channel with exactly zero posts in the window.
5. **Full Channel List** — shown only when `--full` is passed or output is written to a file. Channels are grouped by team.

Example:

```
=== Channel Activity Report ===

  Server:         https://mattermost.example.com
  Team:           (all teams)
  Period:         last 30 days (since 2026-01-27)
  Total channels: 84
  Scope:          public channels only

  Activity Distribution:
    🔵 Very Active  6
    🟢 Active       18
    🟡 Low          22
    🔴 Very Low     15
    💀 Dead         23

  Top 5 Most Active Channels:
    #  TEAM          CHANNEL        POSTS  POSTS/DAY  RATING
    1  Engineering   general        945    31.5       🔵 Very Active
    2  Engineering   announcements  620    20.7       🔵 Very Active
    3  Marketing     campaigns      412    13.7       🔵 Very Active
    4  Engineering   dev-ops        389    13.0       🔵 Very Active
    5  Marketing     general        301    10.0       🟢 Active

  Bottom 5 Least Active Channels:
    #  TEAM          CHANNEL        POSTS  POSTS/DAY  MEMBERS
    1  Engineering   old-project    0      0.0        12
    2  Marketing     test-channel   0      0.0        3
    3  Engineering   archived-exp   0      0.0        7
    4  Sales         q3-planning    0      0.0        5
    5  Marketing     draft-ideas    0      0.0        2

=== Zero-Activity Channels ===

  TEAM          CHANNEL        TYPE    CREATED     MEMBERS
  Engineering   old-project    Public  2024-03-15  12
  Marketing     test-channel   Public  2025-06-01  3
  Engineering   archived-exp   Public  2024-11-20  7
  ...
```

### CSV

CSV output contains all channels with 10 columns:

`team_name`, `channel_name`, `channel_display_name`, `channel_type`, `created_at`, `member_count`, `post_count`, `avg_posts_per_day`, `activity_rating`, `activity_label`

Example:

```csv
team_name,channel_name,channel_display_name,channel_type,created_at,member_count,post_count,avg_posts_per_day,activity_rating,activity_label
Engineering,general,General,Public,2023-01-10,52,945,31.50,5,Very Active
Engineering,old-project,Old Project,Public,2024-03-15,12,0,0.00,1,Dead
```

### JSON

JSON output contains four top-level keys:

- **`parameters`** — the settings used for the analysis (server URL, team, days, thresholds, etc.)
- **`summary`** — total channel count and activity distribution
- **`channels`** — array of all channel results
- **`notice`** — present only when `--include-private` is set, as a reminder that private channel names are included

Example (abbreviated):

```json
{
  "parameters": {
    "server_url": "https://mattermost.example.com",
    "days": 30,
    "since": "2026-01-27",
    "include_private": false,
    "threshold_dead": 0,
    "threshold_very_low": 0.5,
    "threshold_low": 2,
    "threshold_active": 10
  },
  "summary": {
    "total_channels": 84,
    "distribution": {
      "Dead": 23,
      "Very Low": 15,
      "Low": 22,
      "Active": 18,
      "Very Active": 6
    }
  },
  "channels": [
    {
      "team_name": "Engineering",
      "channel_name": "general",
      "channel_display_name": "General",
      "channel_type": "Public",
      "created_at": "2023-01-10",
      "member_count": 52,
      "post_count": 945,
      "avg_posts_per_day": 31.5,
      "activity_rating": 5,
      "activity_label": "Very Active"
    }
  ]
}
```

## Activity ratings

Each channel is assigned one of five activity tiers based on its average posts per day during the analysis window:

| Rating | Posts/Day | Default Threshold |
|--------|-----------|-------------------|
| Dead | exactly 0 | `--threshold-dead 0` |
| Very Low | > 0 and ≤ 0.5 | `--threshold-very-low 0.5` |
| Low | > 0.5 and ≤ 2 | `--threshold-low 2` |
| Active | > 2 and ≤ 10 | `--threshold-active 10` |
| Very Active | > 10 | *(above active threshold)* |

Thresholds are customisable via the `--threshold-*` flags. Thresholds must remain in ascending order (dead ≤ very-low ≤ low ≤ active).

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Configuration or connection error |
| 2 | Analysis failed |
| 3 | Output writing failed |

## Limitations

- **DMs and Group Messages are never included.** The tool only analyses public channels (and private channels if `--include-private` is set).
- **The report is point-in-time only.** It reflects post counts at the moment of execution. Running the tool again later will produce different numbers as new posts are created and the analysis window shifts.

---

## Contributing

We welcome contributions from the community! Whether it's a bug report, a feature suggestion,
or a pull request, your input is valuable to us. Please feel free to contribute in the
following ways:
- **Issues and Pull Requests**: For specific questions, issues, or suggestions for improvements,
  open an issue or a pull request in this repository.
- **Mattermost Community**: Join the discussion in the
  [Integrations and Apps](https://community.mattermost.com/core/channels/integrations) channel
  on the Mattermost Community server.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

For questions, feedback, or contributions regarding this project, please use the following methods:
- **Issues and Pull Requests**: For specific questions, issues, or suggestions for improvements,
  feel free to open an issue or a pull request in this repository.
- **Mattermost Community**: Join us in the Mattermost Community server, where we discuss all
  things related to extending Mattermost. You can find me in the channel
  [Integrations and Apps](https://community.mattermost.com/core/channels/integrations).
- **Social Media**: Follow and message me on Twitter, where I'm
  [@jlandells](https://twitter.com/jlandells).
