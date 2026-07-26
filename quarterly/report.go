package quarterly

import (
	"time"

	"github.com/grokify/gogit"
	"github.com/grokify/gogithub/profile"
	"github.com/grokify/structured-changelog/changelog"
)

// Report is a quarterly developer report joining four data sources.
type Report struct {
	// Metadata
	Username  string    `json:"username"`
	Year      int       `json:"year"`
	Quarter   int       `json:"quarter"` // 1-4
	Label     string    `json:"label"`   // e.g., "Q2 2026"
	Since     time.Time `json:"since"`
	Until     time.Time `json:"until"`
	Generated time.Time `json:"generated"`

	// Section A: Major projects/highlights (curated narrative)
	Highlights *changelog.MultiRepoHighlights `json:"highlights,omitempty"`

	// Section B: Conventional commit category breakdown
	CommitStats *gogit.MultiRepoCommitStats `json:"commitStats,omitempty"`

	// Section C: Raw output stats (GitHub API totals)
	GitHubStats *profile.AggregateStats `json:"githubStats,omitempty"`

	// Section D: Token spend and model distribution (blocked on omnidevx-core)
	TokenSpend *TokenSpendSummary `json:"tokenSpend,omitempty"`

	// Section E: SDLC flow for Sankey (computed from above)
	SDLCFlow *SDLCFlow `json:"sdlcFlow,omitempty"`
}

// TokenSpendSummary holds AI token consumption for the quarter.
// ByModel is blocked on INIT-DEVXREPORTS-001 phases 1+3.
type TokenSpendSummary struct {
	TotalInputTokens   int64   `json:"totalInputTokens"`
	TotalOutputTokens  int64   `json:"totalOutputTokens"`
	TotalCacheRead     int64   `json:"totalCacheRead"`
	TotalCacheCreation int64   `json:"totalCacheCreation"`
	TotalCostUSD       float64 `json:"totalCostUsd,omitempty"`
	// ByModel is nil until INIT-DEVXREPORTS-001 completes.
	ByModel map[string]ModelTokens `json:"byModel,omitempty"`
	// BySource aggregates by provider/product (e.g., "anthropic/claude-code").
	BySource map[string]SourceTokens `json:"bySource,omitempty"`
}

// ModelTokens holds per-model token breakdown.
type ModelTokens struct {
	Model        string  `json:"model"`
	InputTokens  int64   `json:"inputTokens"`
	OutputTokens int64   `json:"outputTokens"`
	CostUSD      float64 `json:"costUsd,omitempty"`
}

// SourceTokens holds per-source token breakdown.
type SourceTokens struct {
	Source       string  `json:"source"` // "provider/product"
	InputTokens  int64   `json:"inputTokens"`
	OutputTokens int64   `json:"outputTokens"`
	CostUSD      float64 `json:"costUsd,omitempty"`
}

// SDLCFlow represents the pipeline for Sankey visualization.
// Inputs (tokens/dollars) → Outputs (LOC/releases/repos) → Categories
type SDLCFlow struct {
	// Inputs
	InputTokens int64   `json:"inputTokens"`
	OutputDollars float64 `json:"outputDollars"`

	// Outputs (intermediate)
	TotalLOC     int `json:"totalLoc"`      // insertions + deletions
	TotalReleases int `json:"totalReleases"`
	TotalRepos   int `json:"totalRepos"`    // contributed + created

	// Category breakdown from commits
	ByCategory []CategoryFlow `json:"byCategory"`
}

// CategoryFlow is one category in the SDLC Sankey.
type CategoryFlow struct {
	Category   string  `json:"category"`
	Commits    int     `json:"commits"`
	Insertions int     `json:"insertions"`
	Deletions  int     `json:"deletions"`
	Percentage float64 `json:"percentage"`
}

// QuarterDates returns the start and end dates for a given quarter.
func QuarterDates(year, quarter int) (since, until time.Time) {
	startMonth := time.Month((quarter-1)*3 + 1)
	since = time.Date(year, startMonth, 1, 0, 0, 0, 0, time.UTC)
	until = since.AddDate(0, 3, 0)
	return
}
