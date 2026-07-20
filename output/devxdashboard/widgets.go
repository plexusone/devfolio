package devxdashboard

import (
	"encoding/json"

	"github.com/plexusone/dashforge/dashboardir"
)

func buildWidgets() []dashboardir.Widget {
	var widgets []dashboardir.Widget

	// Row 1: four headline metric tiles (12 columns / 3 width each).
	widgets = append(widgets,
		metricWidget("sessions", "Sessions", pos(0, 0, 3, 2), "summary", "sessions", "number", nil),
		metricWidget("prompts", "Prompts", pos(3, 0, 3, 2), "summary", "prompts", "number", nil),
		metricWidget("commits", "Commits", pos(6, 0, 3, 2), "summary", "commits", "number", nil),
		metricWidget("ai-assisted-pct", "AI-assisted commits", pos(9, 0, 3, 2), "summary", "aiAssistedPct", "percent", &dashboardir.FormatOptions{Decimals: 1}),
	)

	// Row 2: four more.
	widgets = append(widgets,
		metricWidget("tool-calls", "Tool calls", pos(0, 2, 3, 2), "summary", "toolCalls", "number", nil),
		metricWidget("tool-failure-rate", "Tool failure rate", pos(3, 2, 3, 2), "summary", "toolFailureRate", "percent", &dashboardir.FormatOptions{Decimals: 1}),
		metricWidget("cost", "Cost", pos(6, 2, 3, 2), "summary", "costUsd", "currency", &dashboardir.FormatOptions{Decimals: 0, Prefix: "$"}),
		metricWidget("coverage", "Coverage", pos(9, 2, 3, 2), "summary", "coveragePct", "percent", &dashboardir.FormatOptions{Decimals: 0}),
	)

	// Row 3: daily activity charts.
	widgets = append(widgets,
		lineChartWidget("daily-activity", "Commits & prompts per day", pos(0, 4, 7, 5), "daily",
			[]chartMark{
				{ID: "commits", Geometry: "line", XField: "date", YField: "commits", Name: "Commits", Color: "#2a78d6"},
				{ID: "prompts", Geometry: "line", XField: "date", YField: "prompts", Name: "Prompts", Color: "#008300"},
			}, true),
		lineChartWidget("daily-cost", "Cost per day (USD)", pos(7, 4, 5, 5), "daily",
			[]chartMark{
				{ID: "cost", Geometry: "line", XField: "date", YField: "costUsd", Name: "Cost", Color: "#2a78d6"},
			}, false),
	)

	// Row 4: source coverage table.
	widgets = append(widgets, sourcesTableWidget(pos(0, 9, 12, 5)))

	return widgets
}

func pos(x, y, w, h int) dashboardir.Position {
	return dashboardir.Position{X: x, Y: y, W: w, H: h}
}

func metricWidget(id, title string, p dashboardir.Position, dataSourceID, valueField, format string, opts *dashboardir.FormatOptions) dashboardir.Widget {
	cfg, _ := json.Marshal(dashboardir.MetricConfig{
		ValueField:    valueField,
		Format:        format,
		FormatOptions: opts,
	})
	return dashboardir.Widget{
		ID:           id,
		Title:        title,
		Type:         dashboardir.WidgetTypeMetric,
		Position:     p,
		DataSourceID: dataSourceID,
		Config:       cfg,
	}
}

// chartMark describes one line-chart series. Marshaled by hand (not via a
// dashboardir type — chart config has no concrete Go type in dashboardir,
// it's opaque JSON matching echartify's ChartIR) to match exactly what
// dashforge's viewer parses: marks[] with geometry/encode.
type chartMark struct {
	ID       string
	Geometry string
	XField   string
	YField   string
	Name     string
	Color    string
}

func lineChartWidget(id, title string, p dashboardir.Position, dataSourceID string, marks []chartMark, showLegend bool) dashboardir.Widget {
	type mark struct {
		ID     string            `json:"id"`
		Name   string            `json:"name,omitempty"`
		Geom   string            `json:"geometry"`
		Encode map[string]string `json:"encode"`
		Style  map[string]string `json:"style,omitempty"`
	}
	type axis struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Position string `json:"position"`
	}
	cfg := struct {
		Marks   []mark `json:"marks"`
		Axes    []axis `json:"axes"`
		Tooltip struct {
			Show    bool   `json:"show"`
			Trigger string `json:"trigger"`
		} `json:"tooltip"`
		Legend struct {
			Show     bool   `json:"show"`
			Position string `json:"position"`
		} `json:"legend"`
		Grid map[string]string `json:"grid"`
	}{
		Axes: []axis{
			{ID: "x", Type: "category", Position: "bottom"},
			{ID: "y", Type: "value", Position: "left"},
		},
		Grid: map[string]string{"left": "3%", "right": "4%", "bottom": "10%", "containLabel": "true"},
	}
	cfg.Tooltip.Show = true
	cfg.Tooltip.Trigger = "axis"
	cfg.Legend.Show = showLegend
	cfg.Legend.Position = "top"

	for _, m := range marks {
		cfg.Marks = append(cfg.Marks, mark{
			ID:     m.ID,
			Name:   m.Name,
			Geom:   m.Geometry,
			Encode: map[string]string{"x": m.XField, "y": m.YField},
			Style:  map[string]string{"color": m.Color},
		})
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

func sourcesTableWidget(p dashboardir.Position) dashboardir.Widget {
	cfg, _ := json.Marshal(dashboardir.TableConfig{
		Columns: []dashboardir.TableColumn{
			{Field: "source", Header: "Source", Width: "25%"},
			{Field: "modes", Header: "Collection modes", Width: "30%"},
			{Field: "eventCount", Header: "Events", Width: "20%", Align: "right", Format: "number"},
			{Field: "minConfidence", Header: "Min confidence", Width: "25%", Align: "right"},
		},
		Sortable: true,
		Striped:  true,
	})
	return dashboardir.Widget{
		ID:           "sources-table",
		Title:        "Sources",
		Type:         dashboardir.WidgetTypeTable,
		Position:     p,
		DataSourceID: "sources",
		Config:       cfg,
	}
}
