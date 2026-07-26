package quarterly

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/grokify/gogit"
	"github.com/grokify/gogithub/profile"
	"github.com/grokify/structured-changelog/changelog"
	"github.com/plexusone/omnidevx-core/report"
)

// BuilderOptions configures quarterly report generation.
type BuilderOptions struct {
	Username string
	Year     int
	Quarter  int

	// Paths to local git repositories for commit stats and changelogs.
	RepoPaths []string

	// Path to gogithub profile stats directory (contains report.json).
	StatsDir string

	// Path to omnidevx period report directory (contains period JSON files).
	OmniDevXDir string

	// Workers for parallel repo processing; 0 = GOMAXPROCS.
	Workers int
}

// Builder constructs quarterly reports from multiple sources.
type Builder struct {
	opts BuilderOptions
}

// NewBuilder creates a quarterly report builder.
func NewBuilder(opts BuilderOptions) *Builder {
	return &Builder{opts: opts}
}

// Build generates a complete quarterly report.
func (b *Builder) Build(ctx context.Context) (*Report, error) {
	since, until := QuarterDates(b.opts.Year, b.opts.Quarter)

	r := &Report{
		Username:  b.opts.Username,
		Year:      b.opts.Year,
		Quarter:   b.opts.Quarter,
		Label:     fmt.Sprintf("Q%d %d", b.opts.Quarter, b.opts.Year),
		Since:     since,
		Until:     until,
		Generated: time.Now().UTC(),
	}

	if len(b.opts.RepoPaths) > 0 {
		r.CommitStats = b.collectCommitStats(ctx, since, until)

		changelogPaths := b.changelogPaths()
		if len(changelogPaths) > 0 {
			highlights, err := changelog.ExtractMultiRepoHighlights(changelogPaths, since, until)
			if err == nil {
				r.Highlights = highlights
			}
		}
	}

	if b.opts.StatsDir != "" {
		githubStats, err := b.loadGitHubStats()
		if err == nil {
			r.GitHubStats = githubStats
		}
	}

	if b.opts.OmniDevXDir != "" {
		tokenSpend, err := b.loadTokenSpend(since, until)
		if err == nil {
			r.TokenSpend = tokenSpend
		}
	}

	r.SDLCFlow = b.buildSDLCFlow(r)

	return r, nil
}

// collectCommitStats aggregates commit stats across repos for the date range.
func (b *Builder) collectCommitStats(ctx context.Context, since, until time.Time) *gogit.MultiRepoCommitStats {
	return gogit.AggregateCommitStats(ctx, b.opts.RepoPaths, gogit.CommitStatsOptions{
		Since:    since,
		Until:    until,
		NoMerges: true,
	}, b.opts.Workers)
}

// changelogPaths returns CHANGELOG.json paths for all repos.
func (b *Builder) changelogPaths() []string {
	paths := make([]string, 0, len(b.opts.RepoPaths))
	for _, repo := range b.opts.RepoPaths {
		paths = append(paths, filepath.Join(repo, "CHANGELOG.json"))
	}
	return paths
}

// loadGitHubStats loads the profile stats report and extracts the quarter.
func (b *Builder) loadGitHubStats() (*profile.AggregateStats, error) {
	reportPath := filepath.Join(b.opts.StatsDir, "report.json")
	rpt, err := profile.LoadStatsReport(reportPath)
	if err != nil {
		return nil, err
	}

	quarter := rpt.GetQuarter(b.opts.Year, b.opts.Quarter)
	if quarter == nil {
		return nil, fmt.Errorf("quarter Q%d %d not found in stats report", b.opts.Quarter, b.opts.Year)
	}

	return &quarter.Stats, nil
}

// loadTokenSpend loads the omnidevx period report for the quarter.
func (b *Builder) loadTokenSpend(since, until time.Time) (*TokenSpendSummary, error) {
	// Look for a period report file matching the quarter
	// Format: period-YYYY-QN.json or similar
	pattern := filepath.Join(b.opts.OmniDevXDir, fmt.Sprintf("period-%d-Q%d.json", b.opts.Year, b.opts.Quarter))
	rpt, err := loadPeriodReport(pattern)
	if err != nil {
		return nil, err
	}
	return TokenSpendFromReport(rpt), nil
}

// loadPeriodReport reads a DeveloperPeriodReport from a JSON file.
func loadPeriodReport(path string) (*report.DeveloperPeriodReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rpt report.DeveloperPeriodReport
	if err := json.Unmarshal(data, &rpt); err != nil {
		return nil, err
	}
	return &rpt, nil
}

// buildSDLCFlow computes the Sankey flow from collected data.
func (b *Builder) buildSDLCFlow(r *Report) *SDLCFlow {
	flow := &SDLCFlow{}

	if r.CommitStats != nil {
		flow.TotalLOC = r.CommitStats.TotalStats.Insertions + r.CommitStats.TotalStats.Deletions

		for cat, stats := range r.CommitStats.ByCategory {
			pct := 0.0
			if r.CommitStats.TotalStats.Commits > 0 {
				pct = float64(stats.Commits) / float64(r.CommitStats.TotalStats.Commits) * 100
			}
			flow.ByCategory = append(flow.ByCategory, CategoryFlow{
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

// CategoryPercentages returns a map of category to percentage for easy lookup.
func (r *Report) CategoryPercentages() map[string]float64 {
	if r.CommitStats == nil {
		return nil
	}
	return r.CommitStats.CategoryPercentages()
}

// TopProjects returns the top N projects by highlight entry count.
func (r *Report) TopProjects(n int) []struct {
	Project string
	Count   int
} {
	if r.Highlights == nil {
		return nil
	}
	return r.Highlights.TopProjects(n)
}
