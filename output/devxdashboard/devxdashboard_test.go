package devxdashboard

import (
	"encoding/json"
	"testing"

	"github.com/plexusone/dashforge/dashboardir"
	omnidevx "github.com/plexusone/omnidevx-core"
	report "github.com/plexusone/omnidevx-core/report"
)

func metric(v float64) report.Metric {
	return report.Metric{Value: v, Measurement: report.Measurement{Kind: report.KindObserved, Confidence: 0.9}}
}

func sampleReport() *report.DeveloperPeriodReport {
	return &report.DeveloperPeriodReport{
		SchemaVersion: report.SchemaVersion,
		Subject:       report.Subject{PersonID: "person:john"},
		Sources: []report.SourceCoverage{
			{
				Source:          omnidevx.Source{Provider: "anthropic", Product: "claude-code"},
				EventCount:      1000,
				CollectionModes: []omnidevx.CollectionMode{omnidevx.ModeHistory, omnidevx.ModeOTel},
				MinConfidence:   0.9,
			},
			{
				Source:          omnidevx.Source{Provider: "git", Product: "git"},
				EventCount:      50,
				CollectionModes: []omnidevx.CollectionMode{omnidevx.ModeHistory},
				MinConfidence:   0.95,
			},
		},
		Metrics: report.MetricSet{
			Combined: map[string]report.Metric{
				"sessions":            metric(10),
				"prompts":             metric(200),
				"commits":             metric(50),
				"ai_assisted_commits": metric(35),
				"tool_calls":          metric(400),
				"tool_calls_failed":   metric(20),
				"cost_usd":            metric(12.5),
			},
		},
		Quality: report.DataQuality{CoverageScore: 0.8},
	}
}

func sampleDaily() []DailyPoint {
	return []DailyPoint{
		{Date: "2026-07-01", Commits: 5, Prompts: 20, CostUSD: 1.5},
		{Date: "2026-07-02", Commits: 3, Prompts: 15, CostUSD: 0.8},
	}
}

func TestExportNilReport(t *testing.T) {
	if _, err := Export(nil, nil); err == nil {
		t.Fatal("expected error for nil report")
	}
}

func TestExportProducesValidDashboard(t *testing.T) {
	dash, err := Export(sampleReport(), sampleDaily())
	if err != nil {
		t.Fatal(err)
	}
	if dash.ID == "" || dash.Title == "" {
		t.Errorf("dashboard missing id/title: %+v", dash)
	}
	if len(dash.DataSources) != 3 {
		t.Fatalf("data sources: got %d, want 3", len(dash.DataSources))
	}
	if len(dash.Widgets) != 11 { // 8 metrics + 2 charts + 1 table
		t.Fatalf("widgets: got %d, want 11", len(dash.Widgets))
	}

	// Round-trips through JSON without error (the artifact devfolio writes
	// to disk and hands to dashforge).
	data, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back dashboardir.Dashboard
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func TestExportSummaryComputedFields(t *testing.T) {
	dash, err := Export(sampleReport(), sampleDaily())
	if err != nil {
		t.Fatal(err)
	}
	var summarySrc *dashboardir.DataSource
	for i := range dash.DataSources {
		if dash.DataSources[i].ID == "summary" {
			summarySrc = &dash.DataSources[i]
		}
	}
	if summarySrc == nil {
		t.Fatal("summary data source not found")
	}
	var summary map[string]float64
	if err := json.Unmarshal(summarySrc.Data, &summary); err != nil {
		t.Fatal(err)
	}

	// 35/50 * 100 = 70
	if got := summary["aiAssistedPct"]; got != 70 {
		t.Errorf("aiAssistedPct: got %v, want 70", got)
	}
	// 20/400 * 100 = 5
	if got := summary["toolFailureRate"]; got != 5 {
		t.Errorf("toolFailureRate: got %v, want 5", got)
	}
	// 0.8 * 100 = 80 (pre-scaled for the viewer's percent-format convention)
	if got := summary["coveragePct"]; got != 80 {
		t.Errorf("coveragePct: got %v, want 80", got)
	}
}

func TestExportZeroDenominatorsDoNotDivideByZero(t *testing.T) {
	r := &report.DeveloperPeriodReport{
		Subject: report.Subject{PersonID: "person:empty"},
		Metrics: report.MetricSet{Combined: map[string]report.Metric{}},
	}
	dash, err := Export(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	var summarySrc dashboardir.DataSource
	for _, ds := range dash.DataSources {
		if ds.ID == "summary" {
			summarySrc = ds
		}
	}
	var summary map[string]float64
	if err := json.Unmarshal(summarySrc.Data, &summary); err != nil {
		t.Fatal(err)
	}
	if summary["aiAssistedPct"] != 0 || summary["toolFailureRate"] != 0 {
		t.Errorf("expected zero rates with no denominator, got %+v", summary)
	}
}

func TestExportSourcesTableFlattensNestedSource(t *testing.T) {
	dash, err := Export(sampleReport(), sampleDaily())
	if err != nil {
		t.Fatal(err)
	}
	var sourcesSrc dashboardir.DataSource
	for _, ds := range dash.DataSources {
		if ds.ID == "sources" {
			sourcesSrc = ds
		}
	}
	var rows []sourceRow
	if err := json.Unmarshal(sourcesSrc.Data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows: got %d, want 2", len(rows))
	}
	// Sorted alphabetically: "anthropic/claude-code" before "git/git".
	if rows[0].Source != "anthropic/claude-code" {
		t.Errorf("row 0 source: got %q", rows[0].Source)
	}
	if rows[0].Modes != "history, otel" {
		t.Errorf("row 0 modes: got %q", rows[0].Modes)
	}
}

func TestExportChartWidgetsUseMarksShape(t *testing.T) {
	dash, err := Export(sampleReport(), sampleDaily())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, w := range dash.Widgets {
		if w.Type != dashboardir.WidgetTypeChart {
			continue
		}
		found = true
		var cfg struct {
			Marks []struct {
				Geometry string            `json:"geometry"`
				Encode   map[string]string `json:"encode"`
			} `json:"marks"`
		}
		if err := json.Unmarshal(w.Config, &cfg); err != nil {
			t.Fatalf("widget %s: unmarshal config: %v", w.ID, err)
		}
		if len(cfg.Marks) == 0 {
			t.Fatalf("widget %s: expected marks[], got none (wrong shape?)", w.ID)
		}
		for _, m := range cfg.Marks {
			if m.Geometry == "" || m.Encode["x"] == "" || m.Encode["y"] == "" {
				t.Errorf("widget %s: incomplete mark %+v", w.ID, m)
			}
		}
	}
	if !found {
		t.Fatal("expected at least one chart widget")
	}
}
