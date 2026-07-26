package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/grokify/gogit"
	"github.com/grokify/gogithub/profile"
	"github.com/grokify/structured-changelog/changelog"
	"github.com/plexusone/devfolio/quarterly"
	"github.com/plexusone/omnidevx-core/report"
	"github.com/spf13/cobra"
)

var quarterlyReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a quarterly developer report",
	Long: `Generate a quarterly developer report joining GitHub stats, git commit analytics,
changelog highlights, and token spend data.

Example:
  devfolio quarterly report --username grokify --year 2026 --quarter 2 \
    --repos ~/go/src/github.com/grokify \
    --stats ~/go/src/github.com/grokify/grokify/stats \
    --output quarterly-report.html`,
	RunE: runQuarterlyReport,
}

var (
	qrUsername    string
	qrYear        int
	qrQuarter     int
	qrSince       string
	qrUntil       string
	qrRepoRoot    string
	qrStatsDir    string
	qrStatsFile   string
	qrEventsDir   string
	qrOutput      string
	qrFormat      string
	qrChartEngine string
)

func init() {
	quarterlyCmd.AddCommand(quarterlyReportCmd)

	quarterlyReportCmd.Flags().StringVar(&qrUsername, "username", "", "GitHub username")
	quarterlyReportCmd.Flags().IntVar(&qrYear, "year", 0, "Year (e.g., 2026)")
	quarterlyReportCmd.Flags().IntVar(&qrQuarter, "quarter", 0, "Quarter (1-4)")
	quarterlyReportCmd.Flags().StringVar(&qrSince, "since", "", "Start date (YYYY-MM-DD), overrides quarter")
	quarterlyReportCmd.Flags().StringVar(&qrUntil, "until", "", "End date (YYYY-MM-DD), overrides quarter")
	quarterlyReportCmd.Flags().StringVar(&qrRepoRoot, "repos", "", "Root directory containing git repos")
	quarterlyReportCmd.Flags().StringVar(&qrStatsDir, "stats", "", "Directory with gogithub profile stats (report.json)")
	quarterlyReportCmd.Flags().StringVar(&qrStatsFile, "stats-file", "", "Direct path to quarterly stats JSON file")
	quarterlyReportCmd.Flags().StringVar(&qrEventsDir, "events", "", "omnidevx events directory for token spend")
	quarterlyReportCmd.Flags().StringVar(&qrOutput, "output", "quarterly-report.html", "Output file path")
	quarterlyReportCmd.Flags().StringVar(&qrFormat, "format", "html", "Output format: html, json, dashboard")
	quarterlyReportCmd.Flags().StringVar(&qrChartEngine, "chart-engine", "svg", "Chart engine: svg (self-contained), echarts (CDN)")

	quarterlyReportCmd.MarkFlagRequired("username")
}

func runQuarterlyReport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine date range
	var since, until time.Time
	var label string

	if qrSince != "" && qrUntil != "" {
		var err error
		since, err = time.Parse("2006-01-02", qrSince)
		if err != nil {
			return fmt.Errorf("invalid --since date: %w", err)
		}
		until, err = time.Parse("2006-01-02", qrUntil)
		if err != nil {
			return fmt.Errorf("invalid --until date: %w", err)
		}
		until = until.AddDate(0, 0, 1) // Make end date inclusive
		label = fmt.Sprintf("%s to %s", qrSince, qrUntil)
	} else if qrYear > 0 && qrQuarter > 0 {
		since, until = quarterly.QuarterDates(qrYear, qrQuarter)
		label = fmt.Sprintf("Q%d %d", qrQuarter, qrYear)
	} else {
		return fmt.Errorf("must specify either --year/--quarter or --since/--until")
	}

	// Discover repos with .git directories
	var repoPaths []string
	if qrRepoRoot != "" {
		repoPaths = discoverRepos(qrRepoRoot)
		fmt.Printf("Found %d repositories in %s\n", len(repoPaths), qrRepoRoot)
	}

	// Build report with custom date range
	report := buildCustomReport(ctx, qrUsername, since, until, label, repoPaths)

	// Load stats from direct file if provided
	if qrStatsFile != "" {
		stats, err := loadQuarterlyStatsFile(qrStatsFile)
		if err != nil {
			fmt.Printf("Warning: could not load stats file: %v\n", err)
		} else {
			report.GitHubStats = stats
		}
	}

	// Load token spend from events
	if qrEventsDir != "" {
		tokenSpend, err := loadTokenSpendFromEvents(qrEventsDir, since, until)
		if err != nil {
			fmt.Printf("Warning: could not load events: %v\n", err)
		} else {
			report.TokenSpend = tokenSpend
			fmt.Printf("Loaded token data: %d input, %d output tokens\n",
				tokenSpend.TotalInputTokens, tokenSpend.TotalOutputTokens)
		}
	}

	switch qrFormat {
	case "json":
		return writeJSON(report, qrOutput)
	case "html":
		return writeHTML(report, qrOutput)
	case "dashboard":
		return writeDashboard(report, qrOutput)
	default:
		return fmt.Errorf("unknown format: %s", qrFormat)
	}
}

func discoverRepos(root string) []string {
	var repos []string
	entries, err := os.ReadDir(root)
	if err != nil {
		return repos
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(root, e.Name())
		gitDir := filepath.Join(path, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			repos = append(repos, path)
		}
	}
	return repos
}

func writeJSON(report *quarterly.Report, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func writeHTML(report *quarterly.Report, path string) error {
	html := generateHTML(report)
	return os.WriteFile(path, []byte(html), 0644)
}

func writeDashboard(report *quarterly.Report, path string) error {
	dash, err := quarterly.ExportDashboard(report)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(dash, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func generateHTML(r *quarterly.Report) string {
	return generateHTMLWithEngine(r, qrChartEngine)
}

func generateHTMLWithEngine(r *quarterly.Report, chartEngine string) string {
	var sb strings.Builder

	useECharts := chartEngine == "echarts"

	// CSS styles
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Quarterly Report - ` + r.Username + ` - ` + r.Label + `</title>
<style>
:root {
  --bg: #fafafa;
  --bg-card: #ffffff;
  --text: #1a1a1a;
  --text-muted: #666666;
  --accent: #2563eb;
  --accent-light: #dbeafe;
  --border: #e5e7eb;
  --success: #059669;
  --warning: #d97706;
}
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #0f0f0f;
    --bg-card: #1a1a1a;
    --text: #f5f5f5;
    --text-muted: #a0a0a0;
    --accent: #3b82f6;
    --accent-light: #1e3a5f;
    --border: #2a2a2a;
    --success: #10b981;
    --warning: #f59e0b;
  }
}
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  background: var(--bg);
  color: var(--text);
  line-height: 1.6;
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}
h1 { font-size: 2rem; font-weight: 600; margin-bottom: 0.5rem; }
h2 { font-size: 1.25rem; font-weight: 600; margin: 2rem 0 1rem; color: var(--text); border-bottom: 2px solid var(--accent); padding-bottom: 0.5rem; }
h3 { font-size: 1rem; font-weight: 600; margin: 1rem 0 0.5rem; }
.subtitle { color: var(--text-muted); font-size: 1rem; margin-bottom: 2rem; }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.25rem;
}
.card-title { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); margin-bottom: 0.25rem; }
.card-value { font-size: 1.75rem; font-weight: 700; font-variant-numeric: tabular-nums; }
.card-detail { font-size: 0.875rem; color: var(--text-muted); margin-top: 0.25rem; }
.category-list { list-style: none; }
.category-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border);
}
.category-item:last-child { border-bottom: none; }
.category-bar {
  height: 8px;
  background: var(--accent);
  border-radius: 4px;
  min-width: 4px;
}
.category-name { flex: 1; font-weight: 500; }
.category-pct { font-variant-numeric: tabular-nums; color: var(--text-muted); min-width: 50px; text-align: right; }
.category-count { font-variant-numeric: tabular-nums; min-width: 60px; text-align: right; }
.highlight-list { list-style: none; }
.highlight-item { padding: 0.75rem 0; border-bottom: 1px solid var(--border); }
.highlight-item:last-child { border-bottom: none; }
.highlight-project { font-weight: 600; color: var(--accent); }
.highlight-version { font-size: 0.875rem; color: var(--text-muted); margin-left: 0.5rem; }
.highlight-entries { margin-top: 0.5rem; padding-left: 1rem; }
.highlight-entry { font-size: 0.875rem; color: var(--text-muted); }
.ai-stats { margin-top: 1rem; }
.ai-bar { display: flex; height: 24px; border-radius: 4px; overflow: hidden; margin: 0.5rem 0; }
.ai-bar-segment { display: flex; align-items: center; justify-content: center; font-size: 0.75rem; font-weight: 500; color: white; min-width: 40px; }
.model-list { margin-top: 1rem; }
.model-item { display: flex; justify-content: space-between; padding: 0.25rem 0; font-size: 0.875rem; }
.model-name { font-family: ui-monospace, monospace; }
.model-table { width: 100%; border-collapse: collapse; font-size: 0.8125rem; margin-top: 0.5rem; }
.model-table th { text-align: left; padding: 0.5rem 0.75rem; border-bottom: 2px solid var(--border); font-weight: 600; color: var(--text-muted); font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.03em; }
.model-table th.num { text-align: right; }
.model-table td { padding: 0.5rem 0.75rem; border-bottom: 1px solid var(--border); }
.model-table td.num { text-align: right; font-variant-numeric: tabular-nums; }
.model-table td.model { font-family: ui-monospace, monospace; font-size: 0.75rem; }
.model-table tr:last-child td { border-bottom: none; }
.model-table tr:hover { background: var(--accent-light); }
.donut-row { display: flex; gap: 2rem; margin-top: 1.5rem; flex-wrap: wrap; }
.donut-chart { flex: 1; min-width: 280px; padding: 1rem; background: var(--bg-card); border-radius: 12px; border: 1px solid var(--border); }
.donut-chart h4 { margin-bottom: 1rem; font-size: 0.9375rem; font-weight: 600; color: var(--text); text-align: center; }
.donut-container { position: relative; width: 70%; max-width: 280px; aspect-ratio: 1; margin: 0 auto; filter: drop-shadow(0 4px 12px rgba(0,0,0,0.1)); }
.donut-svg { transform: rotate(-90deg); width: 100%; height: 100%; }
.donut-segment { transition: opacity 0.2s; }
.donut-segment:hover { opacity: 0.8; }
.donut-center { position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); text-align: center; background: var(--bg-card); border-radius: 50%; width: 45%; height: 45%; display: flex; flex-direction: column; align-items: center; justify-content: center; box-shadow: inset 0 2px 8px rgba(0,0,0,0.06); }
.donut-center-value { font-size: 1.5rem; font-weight: 700; color: var(--text); }
.donut-center-label { font-size: 0.6875rem; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; margin-top: 0.125rem; }
.donut-legend { margin-top: 1.25rem; display: flex; flex-wrap: wrap; gap: 0.625rem 1.25rem; justify-content: center; }
.donut-legend-item { display: flex; align-items: center; gap: 0.5rem; font-size: 0.8125rem; color: var(--text-muted); }
.donut-legend-item:hover { color: var(--text); }
.donut-legend-color { width: 12px; height: 12px; border-radius: 3px; box-shadow: 0 1px 3px rgba(0,0,0,0.15); }
.echarts-container { width: 100%; height: 340px; }
.stacked-chart { margin-top: 1.5rem; padding: 1rem; background: var(--bg-card); border-radius: 12px; border: 1px solid var(--border); }
.stacked-chart h4 { margin-bottom: 1rem; font-size: 0.9375rem; font-weight: 600; color: var(--text); }
.stacked-chart-container { display: flex; align-items: flex-end; gap: 0.5rem; height: 250px; padding: 0 0.5rem 2rem; }
.stacked-bar { flex: 1; min-width: 60px; height: 100%; display: flex; flex-direction: column; justify-content: flex-end; }
.stacked-bar-inner { display: flex; flex-direction: column-reverse; border-radius: 4px 4px 0 0; overflow: hidden; }
.stacked-bar-segment { transition: opacity 0.2s; position: relative; min-height: 3px; }
.stacked-bar-segment:hover { opacity: 0.85; }
.stacked-bar-segment:hover::after { content: attr(data-tooltip); position: absolute; bottom: 100%; left: 50%; transform: translateX(-50%); background: var(--text); color: var(--bg); padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.6875rem; white-space: nowrap; z-index: 10; pointer-events: none; }
.stacked-bar-label { text-align: center; font-size: 0.625rem; color: var(--text-muted); margin-top: 0.375rem; word-break: break-all; line-height: 1.2; position: absolute; bottom: -1.75rem; left: 0; right: 0; }
.stacked-legend { display: flex; gap: 1rem; justify-content: center; margin-top: 1rem; flex-wrap: wrap; }
.stacked-legend-item { display: flex; align-items: center; gap: 0.375rem; font-size: 0.75rem; color: var(--text-muted); }
.stacked-legend-color { width: 12px; height: 12px; border-radius: 2px; }
footer { margin-top: 3rem; padding-top: 1rem; border-top: 1px solid var(--border); font-size: 0.75rem; color: var(--text-muted); }

@media print {
  body { padding: 1.5cm; max-width: 100%; font-size: 11pt; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  h1 { font-size: 18pt; }
  h2 { font-size: 14pt; page-break-after: avoid; }
  h3 { font-size: 12pt; }
  .card { break-inside: avoid; }
  .grid { grid-template-columns: repeat(3, 1fr); }
  .donut-row { page-break-inside: avoid; }
  .donut-chart { min-width: 200px; }
  .stacked-chart { page-break-inside: avoid; }
  .model-table { font-size: 9pt; }
  .model-table th, .model-table td { padding: 0.25rem 0.5rem; }
  button { display: none; }
  footer { page-break-before: avoid; }
  a { color: var(--accent); text-decoration: none; }
}
</style>
`)

	if useECharts {
		sb.WriteString(`<script src="https://cdn.jsdelivr.net/npm/echarts@5.5.0/dist/echarts.min.js"></script>
`)
	}

	sb.WriteString(`</head>
<body>
`)

	// Header
	sb.WriteString(fmt.Sprintf(`<h1>%s</h1>`, r.Label))
	sb.WriteString(fmt.Sprintf(`<p class="subtitle">Developer Report for <strong>%s</strong> &middot; %s to %s</p>`,
		r.Username, r.Since.Format("Jan 2, 2006"), r.Until.Add(-1).Format("Jan 2, 2006")))

	// Summary stats
	sb.WriteString(`<h2>Summary</h2>`)
	sb.WriteString(`<div class="grid">`)

	if r.GitHubStats != nil {
		sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Commits</div><div class="card-value">%s</div></div>`,
			formatNumber(r.GitHubStats.Commits)))
		sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Net Additions</div><div class="card-value">+%s</div><div class="card-detail">+%s / -%s</div></div>`,
			formatNumber(r.GitHubStats.NetAdditions), formatNumber(r.GitHubStats.Additions), formatNumber(r.GitHubStats.Deletions)))
		sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Releases</div><div class="card-value">%d</div></div>`,
			r.GitHubStats.Releases))
		sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Repos</div><div class="card-value">%d</div><div class="card-detail">%d contributed, %d created</div></div>`,
			r.GitHubStats.RepoCountContributed+r.GitHubStats.RepoCountCreated,
			r.GitHubStats.RepoCountContributed, r.GitHubStats.RepoCountCreated))
		sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">PRs</div><div class="card-value">%d</div></div>`,
			r.GitHubStats.PRs))
		sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Reviews</div><div class="card-value">%d</div></div>`,
			r.GitHubStats.Reviews))
	}

	sb.WriteString(`</div>`)

	// Commit categories
	if r.CommitStats != nil && len(r.CommitStats.ByCategory) > 0 {
		sb.WriteString(`<h2>Commit Categories</h2>`)
		sb.WriteString(`<div class="card">`)

		// Sort by commits descending
		categories := r.CommitStats.CategoryBreakdown()
		maxCommits := 0
		for _, cat := range categories {
			if cat.Commits > maxCommits {
				maxCommits = cat.Commits
			}
		}

		sb.WriteString(`<ul class="category-list">`)
		for _, cat := range categories {
			pct := 0.0
			if r.CommitStats.TotalStats.Commits > 0 {
				pct = float64(cat.Commits) / float64(r.CommitStats.TotalStats.Commits) * 100
			}
			barWidth := float64(cat.Commits) / float64(maxCommits) * 100
			sb.WriteString(fmt.Sprintf(`<li class="category-item">
				<span class="category-name">%s</span>
				<span class="category-bar" style="width: %.0f%%"></span>
				<span class="category-pct">%.1f%%</span>
				<span class="category-count">%d</span>
			</li>`, cat.Category, barWidth, pct, cat.Commits))
		}
		sb.WriteString(`</ul>`)
		sb.WriteString(`</div>`)

		// LOC by category (parallel chart)
		sb.WriteString(`<h2>Lines of Code by Category</h2>`)
		sb.WriteString(`<div class="card">`)

		// Calculate total LOC and max for scaling
		totalLOC := 0
		maxLOC := 0
		for _, cat := range categories {
			loc := cat.Insertions + cat.Deletions
			totalLOC += loc
			if loc > maxLOC {
				maxLOC = loc
			}
		}

		// Sort by LOC descending
		type locEntry struct {
			category   string
			loc        int
			insertions int
			deletions  int
		}
		var locCategories []locEntry
		for _, cat := range categories {
			locCategories = append(locCategories, locEntry{
				category:   cat.Category,
				loc:        cat.Insertions + cat.Deletions,
				insertions: cat.Insertions,
				deletions:  cat.Deletions,
			})
		}
		sort.Slice(locCategories, func(i, j int) bool {
			return locCategories[i].loc > locCategories[j].loc
		})

		sb.WriteString(`<ul class="category-list">`)
		for _, cat := range locCategories {
			pct := 0.0
			if totalLOC > 0 {
				pct = float64(cat.loc) / float64(totalLOC) * 100
			}
			barWidth := float64(cat.loc) / float64(maxLOC) * 100
			sb.WriteString(fmt.Sprintf(`<li class="category-item">
				<span class="category-name">%s</span>
				<span class="category-bar" style="width: %.0f%%"></span>
				<span class="category-pct">%.1f%%</span>
				<span class="category-count">%s</span>
			</li>`, cat.category, barWidth, pct, formatNumber(cat.loc)))
		}
		sb.WriteString(`</ul>`)
		sb.WriteString(fmt.Sprintf(`<p style="margin-top: 0.5rem; font-size: 0.75rem; color: var(--text-muted);">Total: %s lines (insertions + deletions)</p>`, formatNumber(totalLOC)))
		sb.WriteString(`</div>`)

		// AI-assisted stats
		if r.CommitStats.AIStats.AIAssistedCount > 0 {
			sb.WriteString(`<h2>AI-Assisted Development</h2>`)
			sb.WriteString(`<div class="card">`)
			sb.WriteString(fmt.Sprintf(`<div class="card-title">AI-Assisted Commits</div>`))
			sb.WriteString(fmt.Sprintf(`<div class="card-value">%d <span style="font-size: 1rem; font-weight: normal; color: var(--text-muted)">(%.1f%%)</span></div>`,
				r.CommitStats.AIStats.AIAssistedCount, r.CommitStats.AIStats.AIAssistedPct))

			// Per-tool breakdown
			if len(r.CommitStats.AIStats.ByTool) > 0 {
				sb.WriteString(`<div class="ai-stats">`)
				sb.WriteString(`<h3>By Tool</h3>`)
				sb.WriteString(`<div class="ai-bar">`)
				colors := []string{"#2563eb", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6"}
				i := 0
				for tool, stats := range r.CommitStats.AIStats.ByTool {
					pct := float64(stats.Commits) / float64(r.CommitStats.AIStats.AIAssistedCount) * 100
					color := colors[i%len(colors)]
					sb.WriteString(fmt.Sprintf(`<div class="ai-bar-segment" style="width: %.1f%%; background: %s" title="%s: %d commits">%s</div>`,
						pct, color, tool, stats.Commits, tool))
					i++
				}
				sb.WriteString(`</div>`)
				sb.WriteString(`</div>`)
			}

			// Per-model breakdown
			if len(r.CommitStats.AIStats.ByModel) > 0 {
				sb.WriteString(`<div class="model-list">`)
				sb.WriteString(`<h3>By Model</h3>`)

				// Sort by commits
				type modelEntry struct {
					key   string
					stats gogit.ModelStats
				}
				var models []modelEntry
				for k, v := range r.CommitStats.AIStats.ByModel {
					models = append(models, modelEntry{k, v})
				}
				sort.Slice(models, func(i, j int) bool {
					return models[i].stats.Commits > models[j].stats.Commits
				})

				for _, m := range models {
					sb.WriteString(fmt.Sprintf(`<div class="model-item"><span class="model-name">%s</span><span>%d commits</span></div>`,
						m.key, m.stats.Commits))
				}
				sb.WriteString(`</div>`)
			}

			sb.WriteString(`</div>`)
		}
	}

	// Highlights
	if r.Highlights != nil && len(r.Highlights.ByRepo) > 0 {
		sb.WriteString(`<h2>Project Highlights</h2>`)
		sb.WriteString(fmt.Sprintf(`<p class="subtitle">%d releases across %d projects</p>`,
			len(r.Highlights.TopReleases), r.Highlights.RepoCount))

		// Build repo stats lookup for commits/LOC
		repoStats := make(map[string]gogit.RepoCommitStats)
		if r.CommitStats != nil {
			for _, rs := range r.CommitStats.ByRepo {
				// Extract repo name from path
				name := filepath.Base(rs.Path)
				repoStats[name] = rs
			}
		}

		// Sort releases by LOC (insertions + deletions) if we have stats, otherwise by highlight count
		type rankedRelease struct {
			ranked  changelog.RankedRelease
			commits int
			loc     int
		}
		var releases []rankedRelease
		for _, rr := range r.Highlights.TopReleases {
			rrel := rankedRelease{ranked: rr}
			if rs, ok := repoStats[rr.Project]; ok {
				rrel.commits = rs.TotalStats.Commits
				rrel.loc = rs.TotalStats.Insertions + rs.TotalStats.Deletions
			}
			releases = append(releases, rrel)
		}
		sort.Slice(releases, func(i, j int) bool {
			return releases[i].loc > releases[j].loc
		})

		// Two-column grid
		sb.WriteString(`<div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem;">`)

		maxReleases := 10
		for i, rel := range releases {
			if i >= maxReleases {
				break
			}
			sb.WriteString(`<div class="card" style="padding: 1rem;">`)
			sb.WriteString(fmt.Sprintf(`<div style="display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 0.5rem;">`))
			sb.WriteString(fmt.Sprintf(`<span class="highlight-project" style="font-size: 1rem;">%s</span>`, rel.ranked.Project))
			sb.WriteString(fmt.Sprintf(`<span class="highlight-version">%s</span>`, rel.ranked.Release.Version))
			sb.WriteString(`</div>`)

			// Stats row
			if rel.commits > 0 || rel.loc > 0 {
				sb.WriteString(`<div style="font-size: 0.75rem; color: var(--text-muted); margin-bottom: 0.5rem;">`)
				if rel.commits > 0 {
					sb.WriteString(fmt.Sprintf(`%d commits`, rel.commits))
				}
				if rel.loc > 0 {
					if rel.commits > 0 {
						sb.WriteString(` · `)
					}
					sb.WriteString(fmt.Sprintf(`%s LOC`, formatNumber(rel.loc)))
				}
				sb.WriteString(`</div>`)
			}

			sb.WriteString(`<div class="highlight-entries" style="font-size: 0.8125rem;">`)
			for _, h := range rel.ranked.Release.Highlights {
				sb.WriteString(fmt.Sprintf(`<div class="highlight-entry">★ %s</div>`, h.Description))
			}
			maxAdded := 2
			for j, a := range rel.ranked.Release.Added {
				if j >= maxAdded {
					remaining := len(rel.ranked.Release.Added) - maxAdded
					if remaining > 0 {
						sb.WriteString(fmt.Sprintf(`<div class="highlight-entry" style="color: var(--text-muted);">+ %d more...</div>`, remaining))
					}
					break
				}
				sb.WriteString(fmt.Sprintf(`<div class="highlight-entry">+ %s</div>`, a.Description))
			}
			sb.WriteString(`</div>`)
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	}

	// Token spend
	if r.TokenSpend != nil && (r.TokenSpend.TotalInputTokens > 0 || r.TokenSpend.TotalOutputTokens > 0) {
		sb.WriteString(`<h2>Token Spend</h2>`)

		if len(r.TokenSpend.ByModel) > 0 {
			// Build sorted model data for charts and list
			type modelData struct {
				name   string
				tokens int64
				cost   float64
			}
			var models []modelData
			var totalTokens int64
			var totalCost float64
			for model, mt := range r.TokenSpend.ByModel {
				if model == "" || (mt.InputTokens == 0 && mt.OutputTokens == 0) {
					continue
				}
				// Include all token types: input, output, cache read, cache write
				tokens := mt.InputTokens + mt.OutputTokens + mt.CacheReadTokens + mt.CacheCreationTokens
				models = append(models, modelData{model, tokens, mt.CostUSD})
				totalTokens += tokens
				totalCost += mt.CostUSD
			}
			sort.Slice(models, func(i, j int) bool {
				return modelSortKey(models[i].name) < modelSortKey(models[j].name)
			})

			// Donut charts first
			colors := []string{"#2563eb", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6", "#ec4899", "#06b6d4", "#84cc16"}

			if useECharts {
				// ECharts version
				sb.WriteString(`<div class="donut-row">`)

				// Tokens donut
				sb.WriteString(`<div class="donut-chart">`)
				sb.WriteString(`<h4>Tokens by Model</h4>`)
				sb.WriteString(`<div id="tokens-chart" class="echarts-container"></div>`)
				sb.WriteString(`</div>`)

				// Cost donut
				sb.WriteString(`<div class="donut-chart">`)
				sb.WriteString(`<h4>Cost by Model</h4>`)
				sb.WriteString(`<div id="cost-chart" class="echarts-container"></div>`)
				sb.WriteString(`</div>`)

				sb.WriteString(`</div>`) // donut-row

				// ECharts initialization script
				sb.WriteString(`<script>
document.addEventListener('DOMContentLoaded', function() {
  var colors = [`)
				for i, c := range colors {
					if i > 0 {
						sb.WriteString(",")
					}
					sb.WriteString(fmt.Sprintf(`'%s'`, c))
				}
				sb.WriteString(`];

  var tokensData = [`)
				for i, m := range models {
					if i > 0 {
						sb.WriteString(",")
					}
					sb.WriteString(fmt.Sprintf(`{value: %d, name: '%s'}`, m.tokens, m.name))
				}
				sb.WriteString(`];

  var costData = [`)
				for i, m := range models {
					if i > 0 {
						sb.WriteString(",")
					}
					sb.WriteString(fmt.Sprintf(`{value: %.2f, name: '%s'}`, m.cost, m.name))
				}
				sb.WriteString(`];

  var baseOption = {
    color: colors,
    tooltip: { show: false },
    legend: {
      orient: 'horizontal',
      bottom: 10,
      top: 'auto',
      itemGap: 12,
      itemWidth: 14,
      itemHeight: 14,
      textStyle: {
        color: getComputedStyle(document.body).getPropertyValue('--text-muted').trim() || '#666',
        fontSize: 11
      }
    },
    series: [{
      type: 'pie',
      radius: ['45%', '75%'],
      center: ['50%', '42%'],
      avoidLabelOverlap: true,
      itemStyle: {
        borderRadius: 4,
        borderColor: 'transparent',
        borderWidth: 0
      },
      label: { show: false },
      emphasis: {
        label: {
          show: true,
          fontSize: 12,
          fontWeight: 'bold',
          formatter: '{b}\n{d}%',
          color: getComputedStyle(document.body).getPropertyValue('--text').trim() || '#1a1a1a',
          textBorderColor: getComputedStyle(document.body).getPropertyValue('--bg-card').trim() || '#fff',
          textBorderWidth: 2,
          lineHeight: 16
        },
        itemStyle: { shadowBlur: 8, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.2)' }
      },
      labelLine: { show: false },
      data: []
    }]
  };

  var totalTokens = tokensData.reduce(function(sum, d) { return sum + d.value; }, 0);
  var totalCostVal = costData.reduce(function(sum, d) { return sum + d.value; }, 0);

  function formatTokens(n) {
    if (n >= 1000000) return (n/1000000).toFixed(1) + 'M';
    if (n >= 1000) return (n/1000).toFixed(1) + 'K';
    return n.toString();
  }

  var tokensChart = echarts.init(document.getElementById('tokens-chart'));
  var tokensOption = JSON.parse(JSON.stringify(baseOption));
  tokensOption.series[0].data = tokensData;
  tokensOption.tooltip.formatter = function(p) { return p.name + ': ' + (p.value/1000000).toFixed(1) + 'M (' + p.percent.toFixed(1) + '%)'; };
  tokensOption.graphic = [{
    type: 'group',
    left: 'center',
    top: '38%',
    children: [
      { type: 'text', style: { text: formatTokens(totalTokens), fontSize: 22, fontWeight: 'bold', fill: getComputedStyle(document.body).getPropertyValue('--text').trim() || '#1a1a1a', textAlign: 'center' }, left: 'center', top: 0 },
      { type: 'text', style: { text: 'tokens', fontSize: 11, fill: getComputedStyle(document.body).getPropertyValue('--text-muted').trim() || '#666', textAlign: 'center', textTransform: 'uppercase' }, left: 'center', top: 26 }
    ]
  }];
  tokensChart.setOption(tokensOption);

  var costChart = echarts.init(document.getElementById('cost-chart'));
  var costOption = JSON.parse(JSON.stringify(baseOption));
  costOption.series[0].data = costData;
  costOption.tooltip.formatter = function(p) { return p.name + ': $' + p.value.toFixed(2) + ' (' + p.percent.toFixed(1) + '%)'; };
  costOption.graphic = [{
    type: 'group',
    left: 'center',
    top: '38%',
    children: [
      { type: 'text', style: { text: '$' + Math.round(totalCostVal), fontSize: 22, fontWeight: 'bold', fill: getComputedStyle(document.body).getPropertyValue('--text').trim() || '#1a1a1a', textAlign: 'center' }, left: 'center', top: 0 },
      { type: 'text', style: { text: 'total', fontSize: 11, fill: getComputedStyle(document.body).getPropertyValue('--text-muted').trim() || '#666', textAlign: 'center', textTransform: 'uppercase' }, left: 'center', top: 26 }
    ]
  }];
  costChart.setOption(costOption);

  window.addEventListener('resize', function() {
    tokensChart.resize();
    costChart.resize();
  });
});
</script>
`)
			} else {
				// SVG version
				sb.WriteString(`<div class="donut-row">`)

				// Tokens donut
				sb.WriteString(`<div class="donut-chart">`)
				sb.WriteString(`<h4>Tokens by Model</h4>`)
				sb.WriteString(`<div class="donut-container">`)
				// r=70, circumference = 2*pi*70 = 439.82
				const radius = 70.0
				const circumference = 439.82
				const strokeWidth = 28.0
				const gapPct = 0.8 // small gap between segments

				sb.WriteString(`<svg class="donut-svg" width="200" height="200" viewBox="0 0 200 200">`)
				sb.WriteString(`<defs>`)
				for i, m := range models {
					color := colors[i%len(colors)]
					sb.WriteString(fmt.Sprintf(`<linearGradient id="grad-tok-%d" x1="0%%" y1="0%%" x2="100%%" y2="100%%"><stop offset="0%%" style="stop-color:%s;stop-opacity:1"/><stop offset="100%%" style="stop-color:%s;stop-opacity:0.7"/></linearGradient>`,
						i, color, color))
					_ = m
				}
				sb.WriteString(`</defs>`)
				offset := 0.0
				for i, m := range models {
					pct := float64(m.tokens) / float64(totalTokens) * 100
					displayPct := pct - gapPct
					if displayPct < 0.5 {
						displayPct = pct
					}
					strokeDash := displayPct / 100 * circumference
					strokeOffset := -offset / 100 * circumference
					sb.WriteString(fmt.Sprintf(`<circle class="donut-segment" cx="100" cy="100" r="%.0f" fill="none" stroke="url(#grad-tok-%d)" stroke-width="%.0f" stroke-linecap="round" stroke-dasharray="%.2f %.2f" stroke-dashoffset="%.2f"/>`,
						radius, i, strokeWidth, strokeDash, circumference, strokeOffset))
					offset += pct
				}
				sb.WriteString(`</svg>`)
				sb.WriteString(fmt.Sprintf(`<div class="donut-center"><div class="donut-center-value">%s</div><div class="donut-center-label">tokens</div></div>`,
					formatNumber(int(totalTokens))))
				sb.WriteString(`</div>`)
				sb.WriteString(`<div class="donut-legend">`)
				for i, m := range models {
					pct := float64(m.tokens) / float64(totalTokens) * 100
					color := colors[i%len(colors)]
					sb.WriteString(fmt.Sprintf(`<div class="donut-legend-item"><span class="donut-legend-color" style="background:%s"></span>%s (%.1f%%)</div>`,
						color, m.name, pct))
				}
				sb.WriteString(`</div>`)
				sb.WriteString(`</div>`)

				// Cost donut
				sb.WriteString(`<div class="donut-chart">`)
				sb.WriteString(`<h4>Cost by Model</h4>`)
				sb.WriteString(`<div class="donut-container">`)
				sb.WriteString(`<svg class="donut-svg" width="200" height="200" viewBox="0 0 200 200">`)
				sb.WriteString(`<defs>`)
				for i, m := range models {
					color := colors[i%len(colors)]
					sb.WriteString(fmt.Sprintf(`<linearGradient id="grad-cost-%d" x1="0%%" y1="0%%" x2="100%%" y2="100%%"><stop offset="0%%" style="stop-color:%s;stop-opacity:1"/><stop offset="100%%" style="stop-color:%s;stop-opacity:0.7"/></linearGradient>`,
						i, color, color))
					_ = m
				}
				sb.WriteString(`</defs>`)
				offset = 0.0
				for i, m := range models {
					pct := m.cost / totalCost * 100
					displayPct := pct - gapPct
					if displayPct < 0.5 {
						displayPct = pct
					}
					strokeDash := displayPct / 100 * circumference
					strokeOffset := -offset / 100 * circumference
					sb.WriteString(fmt.Sprintf(`<circle class="donut-segment" cx="100" cy="100" r="%.0f" fill="none" stroke="url(#grad-cost-%d)" stroke-width="%.0f" stroke-linecap="round" stroke-dasharray="%.2f %.2f" stroke-dashoffset="%.2f"/>`,
						radius, i, strokeWidth, strokeDash, circumference, strokeOffset))
					offset += pct
				}
				sb.WriteString(`</svg>`)
				sb.WriteString(fmt.Sprintf(`<div class="donut-center"><div class="donut-center-value">$%.0f</div><div class="donut-center-label">total</div></div>`,
					totalCost))
				sb.WriteString(`</div>`)
				sb.WriteString(`<div class="donut-legend">`)
				for i, m := range models {
					pct := m.cost / totalCost * 100
					color := colors[i%len(colors)]
					sb.WriteString(fmt.Sprintf(`<div class="donut-legend-item"><span class="donut-legend-color" style="background:%s"></span>%s (%.1f%%)</div>`,
						color, m.name, pct))
				}
				sb.WriteString(`</div>`)
				sb.WriteString(`</div>`)

				sb.WriteString(`</div>`) // donut-row
			}

			// Category donut charts (Input/Output/Cache Read/Cache Write)
			// Calculate totals by category across all models
			var catInput, catOutput, catCacheRead, catCacheWrite int64
			var costInput, costOutput, costCacheRead, costCacheWrite float64
			for _, mt := range r.TokenSpend.ByModel {
				catInput += mt.InputTokens
				catOutput += mt.OutputTokens
				catCacheRead += mt.CacheReadTokens
				catCacheWrite += mt.CacheCreationTokens

				pricing, _ := report.LookupPricing(mt.Model)
				costInput += float64(mt.InputTokens) / 1e6 * pricing.InputPerMillion
				costOutput += float64(mt.OutputTokens) / 1e6 * pricing.OutputPerMillion
				costCacheRead += float64(mt.CacheReadTokens) / 1e6 * pricing.CacheReadPerMillion
				costCacheWrite += float64(mt.CacheCreationTokens) / 1e6 * pricing.CacheCreationPerMillion
			}
			catTotalTokens := catInput + catOutput + catCacheRead + catCacheWrite
			catTotalCost := costInput + costOutput + costCacheRead + costCacheWrite

			catColors := []string{"#3b82f6", "#10b981", "#f59e0b", "#ef4444"}
			catLabels := []string{"Input", "Output", "Cache Read", "Cache Write"}
			catTokens := []int64{catInput, catOutput, catCacheRead, catCacheWrite}
			catCosts := []float64{costInput, costOutput, costCacheRead, costCacheWrite}

			if useECharts {
				// ECharts category donuts
				sb.WriteString(`<div class="donut-row">`)
				sb.WriteString(`<div class="donut-chart"><h4>Tokens by Category</h4><div id="cat-tokens-chart" class="echarts-container"></div></div>`)
				sb.WriteString(`<div class="donut-chart"><h4>Cost by Category</h4><div id="cat-cost-chart" class="echarts-container"></div></div>`)
				sb.WriteString(`</div>`)

				sb.WriteString(`<script>
document.addEventListener('DOMContentLoaded', function() {
  var catColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444'];
  var catLabels = ['Input', 'Output', 'Cache Read', 'Cache Write'];
`)
				sb.WriteString(fmt.Sprintf(`  var catTokens = [%d, %d, %d, %d];
`, catInput, catOutput, catCacheRead, catCacheWrite))
				sb.WriteString(fmt.Sprintf(`  var catCosts = [%.2f, %.2f, %.2f, %.2f];
`, costInput, costOutput, costCacheRead, costCacheWrite))
				sb.WriteString(`
  var catTokensData = catLabels.map(function(l, i) { return {value: catTokens[i], name: l}; });
  var catCostsData = catLabels.map(function(l, i) { return {value: catCosts[i], name: l}; });

  var catBaseOption = {
    color: catColors,
    tooltip: { show: false },
    legend: { orient: 'horizontal', bottom: 10, itemGap: 12, textStyle: { fontSize: 11, color: getComputedStyle(document.body).getPropertyValue('--text-muted').trim() || '#666' } },
    series: [{
      type: 'pie', radius: ['45%', '75%'], center: ['50%', '42%'],
      itemStyle: { borderRadius: 4, borderColor: 'transparent', borderWidth: 0 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 12, fontWeight: 'bold', formatter: '{b}\n{d}%', color: getComputedStyle(document.body).getPropertyValue('--text').trim(), textBorderColor: getComputedStyle(document.body).getPropertyValue('--bg-card').trim(), textBorderWidth: 2, lineHeight: 16 } },
      data: []
    }]
  };

  var catTokensChart = echarts.init(document.getElementById('cat-tokens-chart'));
  var catTokensOpt = JSON.parse(JSON.stringify(catBaseOption));
  catTokensOpt.series[0].data = catTokensData;
`)
				sb.WriteString(fmt.Sprintf(`  catTokensOpt.graphic = [{ type: 'group', left: 'center', top: '38%%', children: [
    { type: 'text', style: { text: '%s', fontSize: 22, fontWeight: 'bold', fill: getComputedStyle(document.body).getPropertyValue('--text').trim(), textAlign: 'center' }, left: 'center', top: 0 },
    { type: 'text', style: { text: 'tokens', fontSize: 11, fill: getComputedStyle(document.body).getPropertyValue('--text-muted').trim(), textAlign: 'center' }, left: 'center', top: 26 }
  ]}];
`, formatNumber(int(catTotalTokens))))
				sb.WriteString(`  catTokensChart.setOption(catTokensOpt);

  var catCostChart = echarts.init(document.getElementById('cat-cost-chart'));
  var catCostOpt = JSON.parse(JSON.stringify(catBaseOption));
  catCostOpt.series[0].data = catCostsData;
`)
				sb.WriteString(fmt.Sprintf(`  catCostOpt.graphic = [{ type: 'group', left: 'center', top: '38%%', children: [
    { type: 'text', style: { text: '$%.0f', fontSize: 22, fontWeight: 'bold', fill: getComputedStyle(document.body).getPropertyValue('--text').trim(), textAlign: 'center' }, left: 'center', top: 0 },
    { type: 'text', style: { text: 'total', fontSize: 11, fill: getComputedStyle(document.body).getPropertyValue('--text-muted').trim(), textAlign: 'center' }, left: 'center', top: 26 }
  ]}];
`, catTotalCost))
				sb.WriteString(`  catCostChart.setOption(catCostOpt);

  window.addEventListener('resize', function() { catTokensChart.resize(); catCostChart.resize(); });
});
</script>
`)
			} else {
				// SVG category donuts
				const radius = 70.0
				const circumference = 439.82
				const strokeWidth = 28.0
				const gapPct = 0.8

				sb.WriteString(`<div class="donut-row">`)

				// Category tokens donut
				sb.WriteString(`<div class="donut-chart"><h4>Tokens by Category</h4><div class="donut-container">`)
				sb.WriteString(`<svg class="donut-svg" width="200" height="200" viewBox="0 0 200 200"><defs>`)
				for i, c := range catColors {
					sb.WriteString(fmt.Sprintf(`<linearGradient id="grad-cat-tok-%d" x1="0%%" y1="0%%" x2="100%%" y2="100%%"><stop offset="0%%" style="stop-color:%s;stop-opacity:1"/><stop offset="100%%" style="stop-color:%s;stop-opacity:0.7"/></linearGradient>`, i, c, c))
				}
				sb.WriteString(`</defs>`)
				offset := 0.0
				for i, tok := range catTokens {
					if tok == 0 {
						continue
					}
					pct := float64(tok) / float64(catTotalTokens) * 100
					displayPct := pct - gapPct
					if displayPct < 0.5 {
						displayPct = pct
					}
					strokeDash := displayPct / 100 * circumference
					strokeOffset := -offset / 100 * circumference
					sb.WriteString(fmt.Sprintf(`<circle class="donut-segment" cx="100" cy="100" r="%.0f" fill="none" stroke="url(#grad-cat-tok-%d)" stroke-width="%.0f" stroke-linecap="round" stroke-dasharray="%.2f %.2f" stroke-dashoffset="%.2f"/>`,
						radius, i, strokeWidth, strokeDash, circumference, strokeOffset))
					offset += pct
				}
				sb.WriteString(`</svg>`)
				sb.WriteString(fmt.Sprintf(`<div class="donut-center"><div class="donut-center-value">%s</div><div class="donut-center-label">tokens</div></div>`, formatNumber(int(catTotalTokens))))
				sb.WriteString(`</div><div class="donut-legend">`)
				for i, tok := range catTokens {
					pct := float64(tok) / float64(catTotalTokens) * 100
					sb.WriteString(fmt.Sprintf(`<div class="donut-legend-item"><span class="donut-legend-color" style="background:%s"></span>%s (%.1f%%)</div>`, catColors[i], catLabels[i], pct))
				}
				sb.WriteString(`</div></div>`)

				// Category cost donut
				sb.WriteString(`<div class="donut-chart"><h4>Cost by Category</h4><div class="donut-container">`)
				sb.WriteString(`<svg class="donut-svg" width="200" height="200" viewBox="0 0 200 200"><defs>`)
				for i, c := range catColors {
					sb.WriteString(fmt.Sprintf(`<linearGradient id="grad-cat-cost-%d" x1="0%%" y1="0%%" x2="100%%" y2="100%%"><stop offset="0%%" style="stop-color:%s;stop-opacity:1"/><stop offset="100%%" style="stop-color:%s;stop-opacity:0.7"/></linearGradient>`, i, c, c))
				}
				sb.WriteString(`</defs>`)
				offset = 0.0
				for i, cost := range catCosts {
					if cost < 0.01 {
						continue
					}
					pct := cost / catTotalCost * 100
					displayPct := pct - gapPct
					if displayPct < 0.5 {
						displayPct = pct
					}
					strokeDash := displayPct / 100 * circumference
					strokeOffset := -offset / 100 * circumference
					sb.WriteString(fmt.Sprintf(`<circle class="donut-segment" cx="100" cy="100" r="%.0f" fill="none" stroke="url(#grad-cat-cost-%d)" stroke-width="%.0f" stroke-linecap="round" stroke-dasharray="%.2f %.2f" stroke-dashoffset="%.2f"/>`,
						radius, i, strokeWidth, strokeDash, circumference, strokeOffset))
					offset += pct
				}
				sb.WriteString(`</svg>`)
				sb.WriteString(fmt.Sprintf(`<div class="donut-center"><div class="donut-center-value">$%.0f</div><div class="donut-center-label">total</div></div>`, catTotalCost))
				sb.WriteString(`</div><div class="donut-legend">`)
				for i, cost := range catCosts {
					pct := cost / catTotalCost * 100
					sb.WriteString(fmt.Sprintf(`<div class="donut-legend-item"><span class="donut-legend-color" style="background:%s"></span>%s (%.1f%%)</div>`, catColors[i], catLabels[i], pct))
				}
				sb.WriteString(`</div></div>`)

				sb.WriteString(`</div>`) // donut-row
			}

			// Summary metrics rows
			sb.WriteString(`<div class="grid" style="margin-top: 1.5rem;">`)
			sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Input Tokens</div><div class="card-value">%s</div></div>`,
				formatNumber(int(r.TokenSpend.TotalInputTokens))))
			sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Output Tokens</div><div class="card-value">%s</div></div>`,
				formatNumber(int(r.TokenSpend.TotalOutputTokens))))
			sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Total Cost</div><div class="card-value">$%.2f</div></div>`,
				r.TokenSpend.TotalCostUSD))
			sb.WriteString(`</div>`)
			sb.WriteString(`<div class="grid" style="margin-top: 0.75rem;">`)
			sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Cache Write Tokens</div><div class="card-value">%s</div></div>`,
				formatNumber(int(r.TokenSpend.TotalCacheCreation))))
			sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Cache Read Tokens</div><div class="card-value">%s</div></div>`,
				formatNumber(int(r.TokenSpend.TotalCacheRead))))
			totalAllTokens := r.TokenSpend.TotalInputTokens + r.TokenSpend.TotalOutputTokens + r.TokenSpend.TotalCacheRead + r.TokenSpend.TotalCacheCreation
			sb.WriteString(fmt.Sprintf(`<div class="card"><div class="card-title">Total Tokens</div><div class="card-value">%s</div></div>`,
				formatNumber(int(totalAllTokens))))
			sb.WriteString(`</div>`)

			// Stacked bar charts - Tokens and Cost by model
			segmentColors := []string{"#3b82f6", "#10b981", "#f59e0b", "#ef4444"} // Input, Output, CacheRead, CacheWrite

			// Calculate max values for scaling
			var maxTokens, maxCost float64
			for _, m := range models {
				mt := r.TokenSpend.ByModel[m.name]
				totalTok := float64(mt.InputTokens + mt.OutputTokens + mt.CacheReadTokens + mt.CacheCreationTokens)
				if totalTok > maxTokens {
					maxTokens = totalTok
				}
				// Calculate cost per component
				pricing, _ := report.LookupPricing(m.name)
				inputCost := float64(mt.InputTokens) / 1e6 * pricing.InputPerMillion
				outputCost := float64(mt.OutputTokens) / 1e6 * pricing.OutputPerMillion
				cacheReadCost := float64(mt.CacheReadTokens) / 1e6 * pricing.CacheReadPerMillion
				cacheWriteCost := float64(mt.CacheCreationTokens) / 1e6 * pricing.CacheCreationPerMillion
				totalCostModel := inputCost + outputCost + cacheReadCost + cacheWriteCost
				if totalCostModel > maxCost {
					maxCost = totalCostModel
				}
			}

			// Tokens stacked bar chart
			const chartHeight = 200.0 // pixels
			sb.WriteString(`<div class="stacked-chart">`)
			sb.WriteString(`<h4>Tokens by Model (stacked: Input, Output, Cache Read, Cache Write)</h4>`)
			sb.WriteString(`<div class="stacked-chart-container">`)
			for _, m := range models {
				mt := r.TokenSpend.ByModel[m.name]
				segments := []int64{mt.InputTokens, mt.OutputTokens, mt.CacheReadTokens, mt.CacheCreationTokens}
				labels := []string{"Input", "Output", "Cache Read", "Cache Write"}
				totalTok := float64(mt.InputTokens + mt.OutputTokens + mt.CacheReadTokens + mt.CacheCreationTokens)
				barHeightPx := totalTok / maxTokens * chartHeight

				sb.WriteString(`<div class="stacked-bar" style="position: relative;">`)
				sb.WriteString(fmt.Sprintf(`<div class="stacked-bar-inner" style="height: %.0fpx;">`, barHeightPx))
				for i, seg := range segments {
					if seg == 0 {
						continue
					}
					segPct := float64(seg) / totalTok * 100
					tooltip := fmt.Sprintf("%s: %s", labels[i], formatWithCommas(seg/1000)+"K")
					sb.WriteString(fmt.Sprintf(`<div class="stacked-bar-segment" style="flex: %.1f; background: %s;" data-tooltip="%s"></div>`,
						segPct, segmentColors[i], tooltip))
				}
				sb.WriteString(`</div>`)
				// Model label - show short name
				shortName := m.name
				if strings.HasPrefix(shortName, "claude-") {
					shortName = strings.TrimPrefix(shortName, "claude-")
				}
				if len(shortName) > 12 {
					shortName = shortName[:12]
				}
				sb.WriteString(fmt.Sprintf(`<div class="stacked-bar-label">%s</div>`, shortName))
				sb.WriteString(`</div>`)
			}
			sb.WriteString(`</div>`)
			sb.WriteString(`<div class="stacked-legend">`)
			for i, label := range []string{"Input", "Output", "Cache Read", "Cache Write"} {
				sb.WriteString(fmt.Sprintf(`<div class="stacked-legend-item"><span class="stacked-legend-color" style="background: %s;"></span>%s</div>`, segmentColors[i], label))
			}
			sb.WriteString(`</div>`)
			sb.WriteString(`</div>`)

			// Cost stacked bar chart
			sb.WriteString(`<div class="stacked-chart">`)
			sb.WriteString(`<h4>Cost by Model (stacked: Input, Output, Cache Read, Cache Write)</h4>`)
			sb.WriteString(`<div class="stacked-chart-container">`)
			for _, m := range models {
				mt := r.TokenSpend.ByModel[m.name]
				pricing, _ := report.LookupPricing(m.name)
				costs := []float64{
					float64(mt.InputTokens) / 1e6 * pricing.InputPerMillion,
					float64(mt.OutputTokens) / 1e6 * pricing.OutputPerMillion,
					float64(mt.CacheReadTokens) / 1e6 * pricing.CacheReadPerMillion,
					float64(mt.CacheCreationTokens) / 1e6 * pricing.CacheCreationPerMillion,
				}
				labels := []string{"Input", "Output", "Cache Read", "Cache Write"}
				totalCostModel := costs[0] + costs[1] + costs[2] + costs[3]
				barHeightPx := totalCostModel / maxCost * chartHeight

				sb.WriteString(`<div class="stacked-bar" style="position: relative;">`)
				sb.WriteString(fmt.Sprintf(`<div class="stacked-bar-inner" style="height: %.0fpx;">`, barHeightPx))
				for i, cost := range costs {
					if cost < 0.01 {
						continue
					}
					segPct := cost / totalCostModel * 100
					tooltip := fmt.Sprintf("%s: $%.2f", labels[i], cost)
					sb.WriteString(fmt.Sprintf(`<div class="stacked-bar-segment" style="flex: %.1f; background: %s;" data-tooltip="%s"></div>`,
						segPct, segmentColors[i], tooltip))
				}
				sb.WriteString(`</div>`)
				shortName := m.name
				if strings.HasPrefix(shortName, "claude-") {
					shortName = strings.TrimPrefix(shortName, "claude-")
				}
				if len(shortName) > 12 {
					shortName = shortName[:12]
				}
				sb.WriteString(fmt.Sprintf(`<div class="stacked-bar-label">%s</div>`, shortName))
				sb.WriteString(`</div>`)
			}
			sb.WriteString(`</div>`)
			sb.WriteString(`<div class="stacked-legend">`)
			for i, label := range []string{"Input", "Output", "Cache Read", "Cache Write"} {
				sb.WriteString(fmt.Sprintf(`<div class="stacked-legend-item"><span class="stacked-legend-color" style="background: %s;"></span>%s</div>`, segmentColors[i], label))
			}
			sb.WriteString(`</div>`)
			sb.WriteString(`</div>`)

			// Model table
			sb.WriteString(`<div class="card" style="margin-top: 1.5rem;">`)
			sb.WriteString(`<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem;">`)
			sb.WriteString(`<h3 style="margin: 0;">By Model</h3>`)
			sb.WriteString(`<button onclick="downloadModelCSV()" style="padding: 0.375rem 0.75rem; font-size: 0.75rem; background: var(--accent); color: white; border: none; border-radius: 4px; cursor: pointer;">Download CSV</button>`)
			sb.WriteString(`</div>`)
			sb.WriteString(`<div style="overflow-x: auto;">`)
			sb.WriteString(`<table class="model-table">`)
			sb.WriteString(`<thead><tr>`)
			sb.WriteString(`<th>Model</th>`)
			sb.WriteString(`<th class="num">Input (K)</th>`)
			sb.WriteString(`<th class="num">Output (K)</th>`)
			sb.WriteString(`<th class="num">Cache Read (K)</th>`)
			sb.WriteString(`<th class="num">Cache Write (K)</th>`)
			sb.WriteString(`<th class="num">Tokens %</th>`)
			sb.WriteString(`<th class="num">Cost</th>`)
			sb.WriteString(`<th class="num">Cost %</th>`)
			sb.WriteString(`</tr></thead>`)
			sb.WriteString(`<tbody>`)
			for _, m := range models {
				mt := r.TokenSpend.ByModel[m.name]
				tokenPct := float64(m.tokens) / float64(totalTokens) * 100
				costPct := m.cost / totalCost * 100
				sb.WriteString(fmt.Sprintf(`<tr>
					<td class="model">%s</td>
					<td class="num">%s</td>
					<td class="num">%s</td>
					<td class="num">%s</td>
					<td class="num">%s</td>
					<td class="num">%.1f%%</td>
					<td class="num">$%.2f</td>
					<td class="num">%.1f%%</td>
				</tr>`,
					m.name,
					formatTokensK(mt.InputTokens),
					formatTokensK(mt.OutputTokens),
					formatTokensK(mt.CacheReadTokens),
					formatTokensK(mt.CacheCreationTokens),
					tokenPct,
					m.cost,
					costPct))
			}
			sb.WriteString(`</tbody>`)
			sb.WriteString(`</table>`)
			sb.WriteString(`</div>`)
			sb.WriteString(`<p style="font-size: 0.75rem; color: var(--text-muted); margin-top: 1rem; line-height: 1.6;">`)
			sb.WriteString(`<strong>Cost Formula:</strong> Cost = (Input × Input$/M) + (Output × Output$/M) + (CacheRead × CacheRead$/M) + (CacheWrite × CacheWrite$/M), where all token counts are divided by 1,000,000. `)
			sb.WriteString(`Tokens % is based on Input + Output only. Cache tokens dominate cost for long sessions due to prompt caching.`)
			sb.WriteString(`</p>`)
			sb.WriteString(`</div>`)

			// Pricing reference table - show only models used in this report
			sb.WriteString(`<div class="card" style="margin-top: 1.5rem;">`)
			sb.WriteString(fmt.Sprintf(`<h3>Pricing Reference <span style="font-weight: normal; font-size: 0.75rem; color: var(--text-muted);">(v%s)</span></h3>`, report.PricingVersion()))
			sb.WriteString(`<p style="font-size: 0.75rem; color: var(--text-muted); margin-bottom: 0.75rem;">`)
			sb.WriteString(fmt.Sprintf(`Source: <a href="%s" style="color: var(--accent);">%s</a>`, report.PricingSource(), report.PricingSource()))
			sb.WriteString(`</p>`)
			sb.WriteString(`<div style="overflow-x: auto;">`)
			sb.WriteString(`<table class="model-table">`)
			sb.WriteString(`<thead><tr>`)
			sb.WriteString(`<th>Model</th>`)
			sb.WriteString(`<th class="num">Input $/M</th>`)
			sb.WriteString(`<th class="num">Output $/M</th>`)
			sb.WriteString(`<th class="num">Cache Read $/M</th>`)
			sb.WriteString(`<th class="num">Cache Write $/M</th>`)
			sb.WriteString(`</tr></thead>`)
			sb.WriteString(`<tbody>`)
			for _, m := range models {
				pricing, ok := report.LookupPricing(m.name)
				if !ok {
					continue
				}
				sb.WriteString(fmt.Sprintf(`<tr>
					<td class="model">%s</td>
					<td class="num">$%.2f</td>
					<td class="num">$%.2f</td>
					<td class="num">$%.2f</td>
					<td class="num">$%.2f</td>
				</tr>`,
					m.name,
					pricing.InputPerMillion,
					pricing.OutputPerMillion,
					pricing.CacheReadPerMillion,
					pricing.CacheCreationPerMillion))
			}
			sb.WriteString(`</tbody>`)
			sb.WriteString(`</table>`)
			sb.WriteString(`</div>`)
			sb.WriteString(`</div>`)
		}
	}

	// Footer
	sb.WriteString(fmt.Sprintf(`<footer>Generated %s by devfolio</footer>`, r.Generated.Format("Jan 2, 2006 15:04 MST")))

	// CSV download script (only if we have token data)
	if r.TokenSpend != nil && len(r.TokenSpend.ByModel) > 0 {
		sb.WriteString(`<script>
function downloadModelCSV() {
  var csv = [];
  csv.push('Model,Input Tokens,Output Tokens,Cache Read Tokens,Cache Write Tokens,Cost USD,Input $/M,Output $/M,Cache Read $/M,Cache Write $/M');
`)
		// Build sorted model list same as table
		type csvModel struct {
			name                                                     string
			input, output, cacheRead, cacheWrite                     int64
			cost                                                     float64
			inputPrice, outputPrice, cacheReadPrice, cacheWritePrice float64
		}
		var csvModels []csvModel
		for _, m := range r.TokenSpend.ByModel {
			if m.Model == "" || (m.InputTokens == 0 && m.OutputTokens == 0) {
				continue
			}
			cm := csvModel{
				name:       m.Model,
				input:      m.InputTokens,
				output:     m.OutputTokens,
				cacheRead:  m.CacheReadTokens,
				cacheWrite: m.CacheCreationTokens,
				cost:       m.CostUSD,
			}
			if pricing, ok := report.LookupPricing(m.Model); ok {
				cm.inputPrice = pricing.InputPerMillion
				cm.outputPrice = pricing.OutputPerMillion
				cm.cacheReadPrice = pricing.CacheReadPerMillion
				cm.cacheWritePrice = pricing.CacheCreationPerMillion
			}
			csvModels = append(csvModels, cm)
		}
		sort.Slice(csvModels, func(i, j int) bool {
			return modelSortKey(csvModels[i].name) < modelSortKey(csvModels[j].name)
		})

		for _, m := range csvModels {
			sb.WriteString(fmt.Sprintf(`  csv.push('%s,%d,%d,%d,%d,%.2f,%.2f,%.2f,%.2f,%.2f');
`,
				m.name, m.input, m.output, m.cacheRead, m.cacheWrite, m.cost,
				m.inputPrice, m.outputPrice, m.cacheReadPrice, m.cacheWritePrice))
		}

		sb.WriteString(`
  var blob = new Blob([csv.join('\n')], {type: 'text/csv'});
  var url = URL.createObjectURL(blob);
  var a = document.createElement('a');
  a.href = url;
  a.download = 'token-spend-by-model.csv';
  a.click();
  URL.revokeObjectURL(url);
}
</script>
`)
	}

	sb.WriteString(`</body></html>`)

	return sb.String()
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}

// formatTokensK formats token counts in K (thousands) as plain numbers with commas.
// e.g., 2100000 -> "2,100", 149800 -> "150", 3372386000 -> "3,372,386"
func formatTokensK(n int64) string {
	k := n / 1000
	if k >= 1000 {
		return formatWithCommas(k)
	}
	// For small values, show decimals
	kf := float64(n) / 1000
	if kf >= 100 {
		return fmt.Sprintf("%.0f", kf)
	}
	if kf >= 10 {
		return fmt.Sprintf("%.1f", kf)
	}
	return fmt.Sprintf("%.2f", kf)
}

// formatWithCommas formats an integer with comma separators.
func formatWithCommas(n int64) string {
	if n < 0 {
		return "-" + formatWithCommas(-n)
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return formatWithCommas(n/1000) + fmt.Sprintf(",%03d", n%1000)
}

// modelSortKey returns a sort key for model names.
// Order: Fable (most expensive) > Opus > Sonnet > Haiku > other
// Within a tier, sort by version descending (4-8 before 4-5).
func modelSortKey(name string) string {
	var tier string
	switch {
	case strings.Contains(name, "fable"):
		tier = "0"
	case strings.Contains(name, "opus"):
		tier = "1"
	case strings.Contains(name, "sonnet"):
		tier = "2"
	case strings.Contains(name, "haiku"):
		tier = "3"
	default:
		tier = "9"
	}
	// Invert version numbers so higher versions sort first
	// e.g., "4-8" -> "4-1" (9-8=1), "4-5" -> "4-4" (9-5=4)
	// This makes 4-8 sort before 4-5
	inverted := make([]byte, len(name))
	for i, c := range name {
		if c >= '0' && c <= '9' {
			inverted[i] = byte('9' - (c - '0'))
		} else {
			inverted[i] = byte(c)
		}
	}
	return tier + "-" + string(inverted)
}

// quarterlyStatsFile is the raw structure of gogithub quarterly JSON files.
type quarterlyStatsFile struct {
	Username           string `json:"username"`
	TotalCommits       int    `json:"totalCommits"`
	TotalIssues        int    `json:"totalIssues"`
	TotalPRs           int    `json:"totalPrs"`
	TotalReviews       int    `json:"totalReviews"`
	TotalReleases      int    `json:"totalReleases"`
	TotalAdditions     int    `json:"totalAdditions"`
	TotalDeletions     int    `json:"totalDeletions"`
	NetAdditions       int    `json:"netAdditions"`
	ReposContributedTo int    `json:"reposContributedTo"`
	TotalReposCreated  int    `json:"totalReposCreated"`
}

func buildCustomReport(ctx context.Context, username string, since, until time.Time, label string, repoPaths []string) *quarterly.Report {
	report := &quarterly.Report{
		Username:  username,
		Label:     label,
		Since:     since,
		Until:     until,
		Generated: time.Now().UTC(),
	}

	if len(repoPaths) > 0 {
		report.CommitStats = gogit.AggregateCommitStats(ctx, repoPaths, gogit.CommitStatsOptions{
			Since:    since,
			Until:    until,
			NoMerges: true,
		}, 0)

		// Collect changelog highlights
		var changelogPaths []string
		for _, repo := range repoPaths {
			changelogPaths = append(changelogPaths, filepath.Join(repo, "CHANGELOG.json"))
		}
		if highlights, err := changelog.ExtractMultiRepoHighlights(changelogPaths, since, until); err == nil {
			report.Highlights = highlights
		}
	}

	// Build SDLC flow
	report.SDLCFlow = buildSDLCFlow(report)

	return report
}

func buildSDLCFlow(r *quarterly.Report) *quarterly.SDLCFlow {
	flow := &quarterly.SDLCFlow{}

	if r.CommitStats != nil {
		flow.TotalLOC = r.CommitStats.TotalStats.Insertions + r.CommitStats.TotalStats.Deletions
		for cat, stats := range r.CommitStats.ByCategory {
			pct := 0.0
			if r.CommitStats.TotalStats.Commits > 0 {
				pct = float64(stats.Commits) / float64(r.CommitStats.TotalStats.Commits) * 100
			}
			flow.ByCategory = append(flow.ByCategory, quarterly.CategoryFlow{
				Category:   cat,
				Commits:    stats.Commits,
				Insertions: stats.Insertions,
				Deletions:  stats.Deletions,
				Percentage: pct,
			})
		}
	}

	if r.GitHubStats != nil {
		flow.TotalReleases = r.GitHubStats.Releases
		flow.TotalRepos = r.GitHubStats.RepoCountContributed + r.GitHubStats.RepoCountCreated
	}

	if r.TokenSpend != nil {
		flow.InputTokens = r.TokenSpend.TotalInputTokens + r.TokenSpend.TotalOutputTokens
		flow.OutputDollars = r.TokenSpend.TotalCostUSD
	}

	return flow
}

// loadTokenSpendFromEvents scans the events directory for token usage.
func loadTokenSpendFromEvents(eventsDir string, since, until time.Time) (*quarterly.TokenSpendSummary, error) {
	ts := &quarterly.TokenSpendSummary{
		ByModel:  make(map[string]quarterly.ModelTokens),
		BySource: make(map[string]quarterly.SourceTokens),
	}

	// Walk date directories
	for d := since; d.Before(until); d = d.AddDate(0, 0, 1) {
		dayDir := filepath.Join(eventsDir, d.Format("2006/01/02"))
		claudeFile := filepath.Join(dayDir, "claude-code.jsonl")

		if _, err := os.Stat(claudeFile); err != nil {
			continue
		}

		if err := processEventsFile(claudeFile, ts); err != nil {
			fmt.Printf("Warning: error processing %s: %v\n", claudeFile, err)
		}
	}

	return ts, nil
}

type eventRecord struct {
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

func processEventsFile(path string, ts *quarterly.TokenSpendSummary) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var ev eventRecord
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue
		}

		if ev.Type != "ai.message.completed" {
			continue
		}

		model, _ := ev.Attributes["model"].(string)
		inputTokens := toInt64(ev.Attributes["input_tokens"])
		outputTokens := toInt64(ev.Attributes["output_tokens"])
		cacheRead := toInt64(ev.Attributes["cache_read_tokens"])
		cacheCreation := toInt64(ev.Attributes["cache_creation_tokens"])

		ts.TotalInputTokens += inputTokens
		ts.TotalOutputTokens += outputTokens
		ts.TotalCacheRead += cacheRead
		ts.TotalCacheCreation += cacheCreation

		// Estimate cost using pricing
		cost := estimateCost(model, inputTokens, outputTokens, cacheRead, cacheCreation)
		ts.TotalCostUSD += cost
		ts.EstimatedCostUSD += cost

		// Per-model tracking
		if model != "" {
			mt := ts.ByModel[model]
			mt.Model = model
			mt.InputTokens += inputTokens
			mt.OutputTokens += outputTokens
			mt.CacheReadTokens += cacheRead
			mt.CacheCreationTokens += cacheCreation
			mt.CostUSD += cost
			ts.ByModel[model] = mt
		}

		// Per-source tracking
		source := "anthropic/claude-code"
		st := ts.BySource[source]
		st.Source = source
		st.InputTokens += inputTokens
		st.OutputTokens += outputTokens
		st.CostUSD += cost
		ts.BySource[source] = st
	}

	return scanner.Err()
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	default:
		return 0
	}
}

// estimateCost returns USD cost based on model pricing
func estimateCost(model string, input, output, cacheRead, cacheCreation int64) float64 {
	const perMillion = 1_000_000.0

	// Pricing per million tokens (simplified)
	var inputPrice, outputPrice, cacheReadPrice, cacheCreationPrice float64

	switch {
	case strings.HasPrefix(model, "claude-opus-4-8"), strings.HasPrefix(model, "claude-opus-4-7"):
		inputPrice, outputPrice = 5.0, 25.0
		cacheReadPrice, cacheCreationPrice = 0.5, 6.25
	case strings.HasPrefix(model, "claude-opus-4-5"), strings.HasPrefix(model, "claude-opus-4-6"):
		inputPrice, outputPrice = 5.0, 25.0
		cacheReadPrice, cacheCreationPrice = 0.5, 6.25
	case strings.HasPrefix(model, "claude-sonnet-5"):
		inputPrice, outputPrice = 2.0, 10.0
		cacheReadPrice, cacheCreationPrice = 0.2, 2.5
	case strings.HasPrefix(model, "claude-haiku-4-5"):
		inputPrice, outputPrice = 1.0, 5.0
		cacheReadPrice, cacheCreationPrice = 0.1, 1.25
	case strings.HasPrefix(model, "claude-fable-5"):
		inputPrice, outputPrice = 10.0, 50.0
		cacheReadPrice, cacheCreationPrice = 1.0, 12.5
	default:
		// Unknown model, use Sonnet pricing as default
		inputPrice, outputPrice = 2.0, 10.0
		cacheReadPrice, cacheCreationPrice = 0.2, 2.5
	}

	return float64(input)/perMillion*inputPrice +
		float64(output)/perMillion*outputPrice +
		float64(cacheRead)/perMillion*cacheReadPrice +
		float64(cacheCreation)/perMillion*cacheCreationPrice
}

func loadQuarterlyStatsFile(path string) (*profile.AggregateStats, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw quarterlyStatsFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return &profile.AggregateStats{
		Commits:              raw.TotalCommits,
		Issues:               raw.TotalIssues,
		PRs:                  raw.TotalPRs,
		Reviews:              raw.TotalReviews,
		Releases:             raw.TotalReleases,
		Additions:            raw.TotalAdditions,
		Deletions:            raw.TotalDeletions,
		NetAdditions:         raw.NetAdditions,
		RepoCountContributed: raw.ReposContributedTo,
		RepoCountCreated:     raw.TotalReposCreated,
	}, nil
}
