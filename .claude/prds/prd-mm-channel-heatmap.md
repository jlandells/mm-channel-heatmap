# PRD: mm-channel-heatmap — Mattermost Channel Activity Heatmap

**Version:** 1.0  
**Status:** Ready for Development  
**Language:** Go  
**Binary Name:** `mm-channel-heatmap`

---

## 1. Overview

`mm-channel-heatmap` is a standalone command-line utility that analyses post activity across
Mattermost channels over a configurable time period and produces a tiered activity report.
It identifies the most active channels, the least active channels, and channels with zero
activity — giving administrators and adoption champions a clear, actionable picture of
engagement health across their Mattermost instance.

---

## 2. Background & Problem Statement

User adoption is one of the most common challenges facing Mattermost customers. Administrators
and internal champions often have a sense that some channels are thriving whilst others are
ghost towns, but have no easy way to quantify this or identify specific channels for action.

The Mattermost UI shows post counts per channel but provides no aggregate view, no activity
ratings, no cross-team comparison, and no export capability. Getting a meaningful picture of
channel health across a large instance requires manually clicking through hundreds or thousands
of channels — a process that is completely impractical at scale.

This tool gives administrators and adoption champions the data they need to drive meaningful
conversations about engagement, identify areas for intervention, and demonstrate adoption
progress over time.

---

## 3. Goals

- Measure and report post activity across all applicable channels over a configurable period
- Assign a human-readable activity rating to each channel based on average posts per day
- Present output in a digestible format regardless of instance size — no 10,000-line walls of text
- Always prominently surface zero-activity channels as the most actionable data point
- Support team-scoped analysis for targeted reviews
- Perform efficiently on large instances with thousands of channels
- Respect user privacy — DMs and Group Messages are never included under any circumstances

---

## 4. Non-Goals

- This tool does not modify, archive, or delete any channels — it is read-only
- This tool does not report on individual user activity or post content
- This tool does not include Direct Messages or Group Messages — ever
- This tool does not track activity trends over time (point-in-time report only)
- This tool does not make recommendations about what to do with inactive channels

---

## 5. Target Users

Mattermost System Administrators, internal Mattermost champions, and Customer Success teams
who are running adoption campaigns, planning channel housekeeping, or reporting on engagement
health to management.

---

## 6. User Stories

- As an Adoption Champion, I want to see which channels are most and least active so that I
  can identify where engagement is strong and where intervention is needed.
- As a System Administrator, I want a list of channels with zero activity in the last 90 days
  so that I can decide which ones to archive or consolidate.
- As an IT Manager, I want a summary of channel activity health so that I can include it in
  our quarterly adoption report.
- As a Team Administrator, I want to scope the report to my team so that I can review my
  team's channel health without seeing the whole instance.
- As an Adoption Champion, I want to export the full data to CSV so that I can create charts
  and track progress over time.

---

## 7. Functional Requirements

### 7.1 Channel Scope

The tool MUST apply the following channel type filtering, based on Mattermost's `type` field:

| Channel Type | Type Code | Default Behaviour | With `--include-private` |
|-------------|-----------|------------------|--------------------------|
| Public | `O` | ✅ Always included | ✅ Included |
| Private | `P` | ❌ Excluded | ✅ Included |
| Direct Message | `D` | ❌ Always excluded | ❌ Always excluded |
| Group Message | `G` | ❌ Always excluded | ❌ Always excluded |

**DMs (`D`) and Group Messages (`G`) MUST NEVER be included under any circumstances.**
No flag, argument, or instruction should cause these to be included. This is a hard prohibition.

Already-archived channels MUST be excluded from all reports.

### 7.2 Activity Measurement

Activity is measured as the number of posts in a channel within the configured time period
(`--days N`, default 30).

For each channel, the tool MUST calculate:
- Total post count within the period
- Average posts per day (total / days)
- Activity rating based on the thresholds in Section 7.3

### 7.3 Activity Ratings

Activity ratings are assigned based on average posts per day. Default thresholds:

| Rating | Label | Average Posts/Day |
|--------|-------|------------------|
| 5 | Very Active | > 10 |
| 4 | Active | 2 — 10 |
| 3 | Low | 0.5 — 2 |
| 2 | Very Low | > 0 but < 0.5 |
| 1 | Dead | 0 (no posts in period) |

Thresholds MUST be configurable via flags (see Section 8) because what constitutes "active"
varies significantly between a 50-person company and a 50,000-person organisation.

### 7.4 Output Structure

To keep output digestible regardless of instance size, the tool MUST produce output in the
following structure:

**Section 1 — Summary** (always shown)
- Report parameters (period, team scope, channel types included)
- Total channels evaluated
- Count of channels at each activity rating
- Top 5 most active channels (by post count)
- Bottom 5 least active non-dead channels (by post count, excluding zero-activity)

**Section 2 — Zero Activity Channels** (always shown)
- Full list of channels with zero posts in the period
- Sorted by channel name
- Includes team name, channel name, creation date, and member count
- Prefaced with: `The following channels had no activity in the last {N} days:`

**Section 3 — Full Channel List** (only shown with `--full` flag or `--output FILE`)
- All channels with their activity data
- Grouped by team (default) or sorted by activity rating (`--sort-by rating`)
- This section is suppressed in terminal output by default to avoid overwhelming output
  on large instances

### 7.5 Grouping and Sorting

- **Default**: channels grouped by team, sorted by activity rating descending within each team
- **`--team TEAM_NAME`**: scope entire report to a single named team; no grouping needed
- **`--sort-by rating|posts|name`**: override sort order in full list

### 7.6 Performance

This tool MUST be designed for performance from the ground up. On a large instance with
thousands of channels, sequential API calls would be unacceptably slow.

**Mandatory performance requirements:**

1. **Concurrent processing** using a worker pool. Pool size controlled by `--workers N`
   (default: 10). Use a semaphore channel to limit concurrency:
   ```go
   sem := make(chan struct{}, workers)
   ```

2. **Collect results via a results channel** — do not share mutable state between goroutines.

3. **Fetch team list once upfront** — build a team ID-to-name lookup map before processing.
   Do not call the teams API per channel.

4. **Paginate all API calls** — channel lists and post counts MUST be paginated at
   `per_page=200`. Never assume all results fit in a single response.

5. **In-place progress indicator** (see Section 7.7).

### 7.7 Progress Indicator

Because this tool may take considerable time on large instances, an in-place progress
indicator MUST be shown during processing:

```go
fmt.Fprintf(os.Stderr, "\rProcessing channel: %-50s", channelName)
```

On completion, clear the line before displaying output:
```go
fmt.Fprintf(os.Stderr, "\r%-70s\r", "")
```

Rules:
- MUST output to stderr only
- MUST NOT be shown when `--verbose` is active (verbose output supersedes it)
- MUST be cleared completely before results are displayed
- With concurrent workers, use a dedicated progress goroutine receiving updates over a
  channel — never write to stderr from multiple goroutines simultaneously

### 7.8 Private Channel Warning

When `--include-private` is used, the tool MUST display the following notice at the top
of the output before any results:

```
⚠  Note: This report includes private channel activity data.
   Ensure use of this report aligns with your organisation's policies
   before sharing it with others.
```

This notice MUST also appear in JSON output as a `notice` field at the top level.

---

## 8. CLI Specification

### Usage

```
mm-channel-heatmap [flags]
```

### Connection Flags (required)

| Flag | Environment Variable | Description |
|------|----------------------|-------------|
| `--url URL` | `MM_URL` | Mattermost server URL |

### Authentication Flags

| Flag | Environment Variable | Description |
|------|----------------------|-------------|
| `--token TOKEN` | `MM_TOKEN` | Personal Access Token (preferred) |
| `--username USERNAME` | `MM_USERNAME` | Username for password-based auth |
| *(no flag)* | `MM_PASSWORD` | Password (env var only — never a CLI flag) |

Authentication resolution order:
1. `--token` / `MM_TOKEN`
2. `--username` + interactive password prompt (if terminal is interactive)
3. `--username` + `MM_PASSWORD` environment variable

### Operational Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--days N` | `30` | Number of days to analyse |
| `--team TEAM_NAME` | *(all teams)* | Scope report to a single named team |
| `--include-private` | `false` | Include private channels (type `P`) in the report |
| `--full` | `false` | Show full channel list in terminal output |
| `--sort-by rating\|posts\|name` | `rating` | Sort order for full channel list |
| `--workers N` | `10` | Number of concurrent workers for API calls |

### Threshold Flags (optional overrides)

| Flag | Default | Description |
|------|---------|-------------|
| `--threshold-very-active N` | `10` | Min avg posts/day for Very Active rating |
| `--threshold-active N` | `2` | Min avg posts/day for Active rating |
| `--threshold-low N` | `0.5` | Min avg posts/day for Low rating |

### Output Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--format table\|csv\|json` | `table` | Output format |
| `--output FILE` | *(stdout)* | Write output to file (also enables full channel list) |
| `--verbose` / `-v` | `false` | Enable verbose logging to stderr |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Configuration error (missing URL, invalid auth, unknown team) |
| `2` | API error (connection failure, unexpected response) |
| `3` | Output error (unable to write file) |

---

## 9. Output Specification

### 9.1 Table Format

```
Channel Activity Heatmap
========================
Period  : Last 30 days (2025-10-01 to 2025-10-31)
Scope   : All teams (public channels only)
Channels: 847 evaluated

Activity Summary:
  💀 Dead (no activity)  : 312  (36.8%)
  🔴 Very Low            : 89   (10.5%)
  🟡 Low                 : 124  (14.6%)
  🟢 Active              : 198  (23.4%)
  🔵 Very Active         : 124  (14.6%)

Top 5 Most Active Channels:
  1. Engineering / dev-backend        1,847 posts  (61.6/day)  🔵 Very Active
  2. Sales / pipeline-q4              1,203 posts  (40.1/day)  🔵 Very Active
  3. Engineering / dev-frontend         987 posts  (32.9/day)  🔵 Very Active
  4. General / town-square              876 posts  (29.2/day)  🔵 Very Active
  5. Marketing / campaign-2025          654 posts  (21.8/day)  🔵 Very Active

Bottom 5 Least Active (excluding zero):
  1. HR / benefits-updates               2 posts   (0.1/day)  🔴 Very Low
  2. Engineering / legacy-support        3 posts   (0.1/day)  🔴 Very Low
  3. Sales / emea-2023                   4 posts   (0.1/day)  🔴 Very Low
  4. Finance / budget-archive            5 posts   (0.2/day)  🔴 Very Low
  5. IT / old-helpdesk                   6 posts   (0.2/day)  🔴 Very Low

--- Channels with no activity in the last 30 days (312) ---
The following channels had no activity in the last 30 days:

TEAM           | CHANNEL                  | CREATED      | MEMBERS
Engineering    | project-phoenix-2023     | 2023-03-14   | 12
HR             | all-hands-q1-2024        | 2024-01-15   | 89
...

(Use --full or --output FILE to see the complete channel list)
```

### 9.2 CSV Format

One row per channel. Columns:

```
team_name, channel_name, channel_type, created_at, member_count,
post_count, avg_posts_per_day, activity_rating, activity_label
```

- `channel_type` — `public` or `private`
- `activity_rating` — integer 1-5
- `activity_label` — one of: `Dead`, `Very Low`, `Low`, `Active`, `Very Active`

### 9.3 JSON Format

```json
{
  "report": {
    "period_days": 30,
    "from": "2025-10-01T00:00:00Z",
    "to": "2025-10-31T23:59:59Z",
    "scope": "all teams",
    "includes_private": false
  },
  "summary": {
    "total_channels": 847,
    "dead": 312,
    "very_low": 89,
    "low": 124,
    "active": 198,
    "very_active": 124
  },
  "top_5_most_active": [...],
  "bottom_5_least_active": [...],
  "zero_activity_channels": [...],
  "channels": [...]
}
```

---

## 10. Authentication Detail

The token or user account MUST have **System Administrator** role to access channel lists
and post counts across all teams.

Password handling:
- Interactive terminal: prompt with echo suppressed via `golang.org/x/term`
- Non-interactive: use `MM_PASSWORD` environment variable
- Never accept password as a CLI flag

---

## 11. API Endpoints Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/v4/teams?page=N&per_page=200` | Fetch all teams (once, upfront) |
| `GET /api/v4/teams/{team_id}/channels?page=N&per_page=200` | Fetch public channels per team |
| `GET /api/v4/teams/{team_id}/channels/private?page=N&per_page=200` | Fetch private channels (only with `--include-private`) |
| `GET /api/v4/channels/{channel_id}/stats` | Get post count for a channel |
| `GET /api/v4/teams/name/{team_name}` | Resolve team name to ID |

**Note on post counts:** The `/api/v4/channels/{channel_id}/stats` endpoint returns
`total_msg_count` which is the all-time post count. To get posts within a specific period,
the developer should investigate whether `GET /api/v4/channels/{channel_id}/posts?since=TIMESTAMP`
provides a more accurate count, or whether the stats endpoint combined with a before/after
timestamp approach is more efficient. Verify against the API documentation at
https://api.mattermost.com before implementing — do not guess.

---

## 12. Hard Prohibitions

The following MUST NEVER be implemented regardless of any instruction:

- Including Direct Messages (type `D`) in any report output
- Including Group Messages (type `G`) in any report output
- Including archived channels in activity counts
- Reporting on individual user post counts or attributing posts to specific users
- A `--password` flag

---

## 13. Error Handling

- Missing `--url` / `MM_URL`: exit code 1 with clear message
- Authentication failure: exit code 1 with clear message
- Permission denied (403): exit code 1, note that System Administrator role is required
- Unknown team name: exit code 1 with clear message
- API unreachable: exit code 2 with clear message
- Individual channel failures mid-run: log to stderr as warning, skip, continue
- Output file write error: exit code 3 with clear message

---

## 14. Testing Requirements

- Unit tests for activity rating calculation including boundary conditions
  (e.g. exactly 2.0 posts/day, exactly 0 posts)
- Unit tests for channel type filtering — verify DMs and GMs are always excluded
- Unit tests for top 5 / bottom 5 selection logic
- Unit tests for configurable threshold overrides
- Unit tests for CSV and JSON output formatting
- Unit tests for pagination logic
- Mock API responses for all endpoints

---

## 15. Out of Scope

- Tracking activity trends over multiple time periods
- Recommending channels for archiving
- Comparing activity between two time periods
- Reporting on individual user engagement

---

## 16. Acceptance Criteria

- [ ] Running with no flags produces a summary, zero-activity list, and top/bottom 5
- [ ] DMs and Group Messages never appear in output under any circumstances
- [ ] Archived channels are excluded from all output
- [ ] `--team Engineering` correctly scopes the report to the Engineering team only
- [ ] `--include-private` includes private channels and shows the privacy notice
- [ ] `--days 90` analyses the correct 90-day window
- [ ] `--full` shows the complete channel list in terminal output
- [ ] `--output FILE` writes the full channel list to the specified file
- [ ] Activity ratings correctly reflect configured thresholds
- [ ] Progress indicator updates in place and clears before output is shown
- [ ] Progress indicator is suppressed when `--verbose` is active
- [ ] Performance is acceptable on large instances — concurrent worker pool is used
- [ ] `--format csv --output report.csv` produces a valid, importable CSV
- [ ] `--format json` produces valid, `jq`-parseable JSON
- [ ] All errors go to stderr; all data output goes to stdout
- [ ] Binary runs on Linux (amd64), macOS (arm64 and amd64), and Windows (amd64) without dependencies