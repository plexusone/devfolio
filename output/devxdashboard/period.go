package devxdashboard

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/plexusone/omnidevx-core/report"
	"github.com/plexusone/uiforge/dashboardir"
)

// PeriodType identifies the granularity of a period report.
type PeriodType string

const (
	PeriodWeekly    PeriodType = "weekly"
	PeriodMonthly   PeriodType = "monthly"
	PeriodQuarterly PeriodType = "quarterly"
)

// PeriodReport holds all data needed to generate a period dashboard.
type PeriodReport struct {
	Type   PeriodType
	Label  string // e.g., "2026-W30", "2026-07", "2026-Q3"
	Report *report.DeveloperPeriodReport
	Daily  []DailyPoint

	// WeeklyByModel holds weekly model breakdown (for monthly/quarterly).
	WeeklyByModel []ModelPeriodPoint

	// MonthlyByModel holds monthly model breakdown (for quarterly only).
	MonthlyByModel []ModelPeriodPoint
}

// ModelPeriodPoint represents one sub-period's per-model metrics (e.g. one
// week's input/output tokens and cost, keyed by model). Models holds every
// metric a caller may want to chart (input_tokens, output_tokens, cost_usd,
// ...) so a single point can back both the tokens and cost stacked-bar
// charts without rebuilding sub-periods twice.
type ModelPeriodPoint struct {
	Label  string                        `json:"label"`  // e.g., "W28", "Jul"
	Models map[string]map[string]float64 `json:"models"` // model name → metric → value
}

// ExportPeriod builds a dashboard for a period report with model breakdowns.
func ExportPeriod(pr *PeriodReport) (*dashboardir.Dashboard, error) {
	if pr == nil || pr.Report == nil {
		return nil, fmt.Errorf("devxdashboard: period report is nil")
	}

	dataSources, err := buildPeriodDataSources(pr)
	if err != nil {
		return nil, fmt.Errorf("devxdashboard: building data sources: %w", err)
	}

	return &dashboardir.Dashboard{
		ID:          fmt.Sprintf("omnidevx-%s-report", pr.Type),
		Title:       fmt.Sprintf("OmniDevX %s Report: %s", periodTitle(pr.Type), pr.Label),
		Description: fmt.Sprintf("Developer activity for %s (%s)", pr.Report.Subject.PersonID, pr.Label),
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
		Widgets:     buildPeriodWidgets(pr.Type),
	}, nil
}

func periodTitle(t PeriodType) string {
	switch t {
	case PeriodWeekly:
		return "Weekly"
	case PeriodMonthly:
		return "Monthly"
	case PeriodQuarterly:
		return "Quarterly"
	default:
		return "Period"
	}
}

func buildPeriodDataSources(pr *PeriodReport) ([]dashboardir.DataSource, error) {
	r := pr.Report

	// Base summary metrics (same as regular export)
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

	sources := []dashboardir.DataSource{}

	summaryJSON, _ := json.Marshal(summary)
	sources = append(sources, dashboardir.DataSource{
		ID: "summary", Name: "Summary", Type: dashboardir.DataSourceTypeInline, Data: summaryJSON,
	})

	dailyJSON, _ := json.Marshal(pr.Daily)
	sources = append(sources, dashboardir.DataSource{
		ID: "daily", Name: "Daily activity", Type: dashboardir.DataSourceTypeInline, Data: dailyJSON,
	})

	sourcesJSON, _ := json.Marshal(sourceRows(r.Sources))
	sources = append(sources, dashboardir.DataSource{
		ID: "sources", Name: "Sources", Type: dashboardir.DataSourceTypeInline, Data: sourcesJSON,
	})

	// Model breakdown data sources
	if len(r.Metrics.ByModel) > 0 {
		// Donut data: tokens by model
		tokensByModel := buildModelDonutData(r.Metrics.ByModel, "input_tokens", "output_tokens")
		tokensByModelJSON, _ := json.Marshal(tokensByModel)
		sources = append(sources, dashboardir.DataSource{
			ID: "model-tokens-total", Name: "Tokens by model", Type: dashboardir.DataSourceTypeInline, Data: tokensByModelJSON,
		})

		// Donut data: cost by model
		costByModel := buildModelDonutData(r.Metrics.ByModel, "cost_usd")
		costByModelJSON, _ := json.Marshal(costByModel)
		sources = append(sources, dashboardir.DataSource{
			ID: "model-cost-total", Name: "Cost by model", Type: dashboardir.DataSourceTypeInline, Data: costByModelJSON,
		})
	}

	// Weekly stacked bar data (for monthly/quarterly)
	if len(pr.WeeklyByModel) > 0 {
		weeklyTokens := buildStackedBarData(pr.WeeklyByModel, "input_tokens", "output_tokens")
		weeklyTokensJSON, _ := json.Marshal(weeklyTokens)
		sources = append(sources, dashboardir.DataSource{
			ID: "weekly-tokens-by-model", Name: "Weekly tokens by model", Type: dashboardir.DataSourceTypeInline, Data: weeklyTokensJSON,
		})

		weeklyCost := buildStackedBarData(pr.WeeklyByModel, "cost_usd")
		weeklyCostJSON, _ := json.Marshal(weeklyCost)
		sources = append(sources, dashboardir.DataSource{
			ID: "weekly-cost-by-model", Name: "Weekly cost by model", Type: dashboardir.DataSourceTypeInline, Data: weeklyCostJSON,
		})
	}

	// Monthly stacked bar data (for quarterly only)
	if len(pr.MonthlyByModel) > 0 {
		monthlyTokens := buildStackedBarData(pr.MonthlyByModel, "input_tokens", "output_tokens")
		monthlyTokensJSON, _ := json.Marshal(monthlyTokens)
		sources = append(sources, dashboardir.DataSource{
			ID: "monthly-tokens-by-model", Name: "Monthly tokens by model", Type: dashboardir.DataSourceTypeInline, Data: monthlyTokensJSON,
		})

		monthlyCost := buildStackedBarData(pr.MonthlyByModel, "cost_usd")
		monthlyCostJSON, _ := json.Marshal(monthlyCost)
		sources = append(sources, dashboardir.DataSource{
			ID: "monthly-cost-by-model", Name: "Monthly cost by model", Type: dashboardir.DataSourceTypeInline, Data: monthlyCostJSON,
		})
	}

	return sources, nil
}

// donutPoint is one segment in a donut/pie chart.
type donutPoint struct {
	Model string  `json:"model"`
	Value float64 `json:"value"`
}

func buildModelDonutData(byModel map[string]map[string]report.Metric, metrics ...string) []donutPoint {
	points := make([]donutPoint, 0, len(byModel))
	for model, modelMetrics := range byModel {
		var total float64
		for _, metric := range metrics {
			if m, ok := modelMetrics[metric]; ok {
				total += m.Value
			}
		}
		if total > 0 {
			points = append(points, donutPoint{Model: model, Value: total})
		}
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Value > points[j].Value })
	return points
}

func buildStackedBarData(periodPoints []ModelPeriodPoint, metrics ...string) []map[string]any {
	result := make([]map[string]any, 0, len(periodPoints))
	for _, pp := range periodPoints {
		row := map[string]any{"period": pp.Label}
		for model, modelMetrics := range pp.Models {
			var total float64
			for _, metric := range metrics {
				total += modelMetrics[metric]
			}
			row[model] = total
		}
		result = append(result, row)
	}
	return result
}

// WeekLabel returns the ISO week label for a date (e.g., "W28").
func WeekLabel(t time.Time) string {
	_, week := t.ISOWeek()
	return fmt.Sprintf("W%02d", week)
}

// MonthLabel returns the month label for a date (e.g., "Jul").
func MonthLabel(t time.Time) string {
	return t.Format("Jan")
}
