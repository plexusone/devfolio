package devxdashboard

import (
	"encoding/json"

	"github.com/plexusone/uiforge/dashboardir"
)

func buildPeriodWidgets(periodType PeriodType) []dashboardir.Widget {
	var widgets []dashboardir.Widget

	// Row 0-1: eight headline metric tiles (same as base dashboard)
	widgets = append(widgets,
		metricWidget("sessions", "Sessions", pos(0, 0, 3, 2), "summary", "sessions", "number", nil),
		metricWidget("prompts", "Prompts", pos(3, 0, 3, 2), "summary", "prompts", "number", nil),
		metricWidget("commits", "Commits", pos(6, 0, 3, 2), "summary", "commits", "number", nil),
		metricWidget("ai-assisted-pct", "AI-assisted commits", pos(9, 0, 3, 2), "summary", "aiAssistedPct", "percent", &dashboardir.FormatOptions{Decimals: 1}),
	)
	widgets = append(widgets,
		metricWidget("tool-calls", "Tool calls", pos(0, 2, 3, 2), "summary", "toolCalls", "number", nil),
		metricWidget("tool-failure-rate", "Tool failure rate", pos(3, 2, 3, 2), "summary", "toolFailureRate", "percent", &dashboardir.FormatOptions{Decimals: 1}),
		metricWidget("cost", "Cost", pos(6, 2, 3, 2), "summary", "costUsd", "currency", &dashboardir.FormatOptions{Decimals: 0, Prefix: "$"}),
		metricWidget("coverage", "Coverage", pos(9, 2, 3, 2), "summary", "coveragePct", "percent", &dashboardir.FormatOptions{Decimals: 0}),
	)

	// Row 2: donut charts for tokens and cost by model
	widgets = append(widgets,
		donutChartWidget("tokens-by-model", "Tokens by Model", pos(0, 4, 6, 5), "model-tokens-total"),
		donutChartWidget("cost-by-model", "Cost by Model", pos(6, 4, 6, 5), "model-cost-total"),
	)

	yOffset := 9

	// For monthly/quarterly: weekly stacked bar charts
	if periodType == PeriodMonthly || periodType == PeriodQuarterly {
		widgets = append(widgets,
			stackedBarChartWidget("weekly-tokens", "Weekly Tokens by Model", pos(0, yOffset, 12, 5), "weekly-tokens-by-model"),
			stackedBarChartWidget("weekly-cost", "Weekly Cost by Model", pos(0, yOffset+5, 12, 5), "weekly-cost-by-model"),
		)
		yOffset += 10
	}

	// For quarterly: monthly stacked bar charts
	if periodType == PeriodQuarterly {
		widgets = append(widgets,
			stackedBarChartWidget("monthly-tokens", "Monthly Tokens by Model", pos(0, yOffset, 12, 5), "monthly-tokens-by-model"),
			stackedBarChartWidget("monthly-cost", "Monthly Cost by Model", pos(0, yOffset+5, 12, 5), "monthly-cost-by-model"),
		)
		yOffset += 10
	}

	// Daily activity charts
	widgets = append(widgets,
		lineChartWidget("daily-activity", "Commits & prompts per day", pos(0, yOffset, 7, 5), "daily",
			[]chartMark{
				{ID: "commits", Geometry: "line", XField: "date", YField: "commits", Name: "Commits", Color: "#2a78d6"},
				{ID: "prompts", Geometry: "line", XField: "date", YField: "prompts", Name: "Prompts", Color: "#008300"},
			}, true),
		lineChartWidget("daily-cost", "Cost per day (USD)", pos(7, yOffset, 5, 5), "daily",
			[]chartMark{
				{ID: "cost", Geometry: "line", XField: "date", YField: "costUsd", Name: "Cost", Color: "#2a78d6"},
			}, false),
	)
	yOffset += 5

	// Source coverage table
	widgets = append(widgets, sourcesTableWidget(pos(0, yOffset, 12, 5)))

	return widgets
}

func donutChartWidget(id, title string, p dashboardir.Position, dataSourceID string) dashboardir.Widget {
	cfg := struct {
		Marks []struct {
			ID     string            `json:"id"`
			Geom   string            `json:"geometry"`
			Encode map[string]string `json:"encode"`
		} `json:"marks"`
		Legend struct {
			Show     bool   `json:"show"`
			Position string `json:"position"`
		} `json:"legend"`
		Tooltip struct {
			Show    bool   `json:"show"`
			Trigger string `json:"trigger"`
		} `json:"tooltip"`
	}{}

	cfg.Marks = []struct {
		ID     string            `json:"id"`
		Geom   string            `json:"geometry"`
		Encode map[string]string `json:"encode"`
	}{
		{
			ID:   "pie",
			Geom: "pie",
			Encode: map[string]string{
				"value": "value",
				"name":  "model",
			},
		},
	}
	cfg.Legend.Show = true
	cfg.Legend.Position = "right"
	cfg.Tooltip.Show = true
	cfg.Tooltip.Trigger = "item"

	cfgJSON, _ := json.Marshal(cfg)
	return dashboardir.Widget{
		ID:           id,
		Title:        title,
		Type:         dashboardir.WidgetTypeChart,
		Position:     p,
		DataSourceID: dataSourceID,
		Config:       cfgJSON,
	}
}

func stackedBarChartWidget(id, title string, p dashboardir.Position, dataSourceID string) dashboardir.Widget {
	cfg := struct {
		Marks []struct {
			ID     string            `json:"id"`
			Name   string            `json:"name,omitempty"`
			Geom   string            `json:"geometry"`
			Stack  string            `json:"stack,omitempty"`
			Encode map[string]string `json:"encode"`
		} `json:"marks"`
		Axes []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Position string `json:"position"`
		} `json:"axes"`
		Legend struct {
			Show     bool   `json:"show"`
			Position string `json:"position"`
		} `json:"legend"`
		Tooltip struct {
			Show    bool   `json:"show"`
			Trigger string `json:"trigger"`
		} `json:"tooltip"`
		Grid map[string]string `json:"grid"`
	}{
		Axes: []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Position string `json:"position"`
		}{
			{ID: "x", Type: "category", Position: "bottom"},
			{ID: "y", Type: "value", Position: "left"},
		},
		Grid: map[string]string{"left": "3%", "right": "4%", "bottom": "10%", "containLabel": "true"},
	}
	cfg.Legend.Show = true
	cfg.Legend.Position = "top"
	cfg.Tooltip.Show = true
	cfg.Tooltip.Trigger = "axis"

	// Dynamic marks will be populated by the renderer based on data
	// For now, we use a placeholder mark that indicates stacked bar
	cfg.Marks = []struct {
		ID     string            `json:"id"`
		Name   string            `json:"name,omitempty"`
		Geom   string            `json:"geometry"`
		Stack  string            `json:"stack,omitempty"`
		Encode map[string]string `json:"encode"`
	}{
		{
			ID:    "stacked",
			Geom:  "bar",
			Stack: "total",
			Encode: map[string]string{
				"x": "period",
			},
		},
	}

	cfgJSON, _ := json.Marshal(cfg)
	return dashboardir.Widget{
		ID:           id,
		Title:        title,
		Type:         dashboardir.WidgetTypeChart,
		Position:     p,
		DataSourceID: dataSourceID,
		Config:       cfgJSON,
	}
}
