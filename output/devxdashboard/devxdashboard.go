// Package devxdashboard projects an omnidevx-core DeveloperPeriodReport into
// a dashforge Dashboard: the disclosure-safe view a caller has already
// chosen to show (this package performs no redaction itself — callers pass
// in whatever report they intend to display).
//
// Chart widget config is hand-built JSON matching the shape dashforge's own
// viewer (viewer/index.html, compileChartIR) actually parses — marks[] with
// geometry/encode, not the singular mark/encodings shape used by devfolio's
// older output/dashboard package, which predates that shape and does not
// render in the current viewer.
package devxdashboard

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/plexusone/dashforge/dashboardir"
	report "github.com/plexusone/omnidevx-core/report"
)

// DailyPoint is one day's activity. Callers build this slice themselves via
// report.BuildDaily (see cmd/devfolio/devx_dashboard.go for the reference
// implementation) since DeveloperPeriodReport itself only carries period
// totals, not a daily series.
type DailyPoint struct {
	Date    string  `json:"date"`
	Commits float64 `json:"commits"`
	Prompts float64 `json:"prompts"`
	CostUSD float64 `json:"costUsd"`
}

// Export converts a DeveloperPeriodReport into a dashforge Dashboard: eight
// headline metric tiles, a daily commits/prompts chart, a daily cost chart,
// and a source-coverage table. All data is embedded inline, so the result
// is a single portable JSON file.
func Export(r *report.DeveloperPeriodReport, daily []DailyPoint) (*dashboardir.Dashboard, error) {
	if r == nil {
		return nil, fmt.Errorf("devxdashboard: report is nil")
	}

	dataSources, err := buildDataSources(r, daily)
	if err != nil {
		return nil, fmt.Errorf("devxdashboard: building data sources: %w", err)
	}

	return &dashboardir.Dashboard{
		ID:          "omnidevx-period-report",
		Title:       "OmniDevX Period Report",
		Description: fmt.Sprintf("Developer activity for %s", r.Subject.PersonID),
		Version:     "1.0.0",
		Layout: dashboardir.Layout{
			Type:      dashboardir.LayoutTypeGrid,
			Columns:   12,
			RowHeight: 64,
			Gap:       16,
			Padding:   16,
		},
		Theme:       &dashboardir.Theme{Mode: "light"},
		DataSources: dataSources,
		Widgets:     buildWidgets(),
	}, nil
}

func metricValue(r *report.DeveloperPeriodReport, key string) float64 {
	if m, ok := r.Metrics.Combined[key]; ok {
		return m.Value
	}
	return 0
}

func buildDataSources(r *report.DeveloperPeriodReport, daily []DailyPoint) ([]dashboardir.DataSource, error) {
	commits := metricValue(r, "commits")
	aiAssisted := metricValue(r, "ai_assisted_commits")
	toolCalls := metricValue(r, "tool_calls")
	toolFailed := metricValue(r, "tool_calls_failed")

	var aiAssistedPct, failureRate float64
	if commits > 0 {
		aiAssistedPct = aiAssisted / commits * 100
	}
	if toolCalls > 0 {
		failureRate = toolFailed / toolCalls * 100
	}

	// Percent-format metric widgets expect a pre-scaled 0-100 value, not a
	// 0-1 fraction (matches the established convention in dashforge's own
	// compliance-dashboard example; the viewer's format="percent" heuristic
	// mishandles fractional values below 1 either way, so this is also the
	// only convention that renders our headline values correctly).
	summary := map[string]any{
		"sessions":        metricValue(r, "sessions"),
		"prompts":         metricValue(r, "prompts"),
		"commits":         commits,
		"aiAssistedPct":   aiAssistedPct,
		"toolCalls":       toolCalls,
		"toolFailureRate": failureRate,
		"costUsd":         metricValue(r, "cost_usd"),
		"coveragePct":     r.Quality.CoverageScore * 100,
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return nil, err
	}

	dailyJSON, err := json.Marshal(daily)
	if err != nil {
		return nil, err
	}

	sourcesJSON, err := json.Marshal(sourceRows(r.Sources))
	if err != nil {
		return nil, err
	}

	return []dashboardir.DataSource{
		{ID: "summary", Name: "Summary", Type: dashboardir.DataSourceTypeInline, Data: summaryJSON},
		{ID: "daily", Name: "Daily activity", Type: dashboardir.DataSourceTypeInline, Data: dailyJSON},
		{ID: "sources", Name: "Sources", Type: dashboardir.DataSourceTypeInline, Data: sourcesJSON},
	}, nil
}

// sourceRow flattens SourceCoverage for table rendering — dashforge's table
// widget binds columns to top-level fields, not nested paths.
type sourceRow struct {
	Source        string  `json:"source"`
	Modes         string  `json:"modes"`
	EventCount    int     `json:"eventCount"`
	MinConfidence float64 `json:"minConfidence"`
}

func sourceRows(sources []report.SourceCoverage) []sourceRow {
	rows := make([]sourceRow, 0, len(sources))
	for _, s := range sources {
		modes := make([]string, 0, len(s.CollectionModes))
		for _, m := range s.CollectionModes {
			modes = append(modes, string(m))
		}
		rows = append(rows, sourceRow{
			Source:        s.Source.Provider + "/" + s.Source.Product,
			Modes:         strings.Join(modes, ", "),
			EventCount:    s.EventCount,
			MinConfidence: s.MinConfidence,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Source < rows[j].Source })
	return rows
}
