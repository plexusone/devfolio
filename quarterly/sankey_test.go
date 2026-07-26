package quarterly

import (
	"encoding/json"
	"testing"
)

func TestBuildSankeyData(t *testing.T) {
	flow := &SDLCFlow{
		InputTokens:   1000000,
		OutputDollars: 50.0,
		TotalLOC:      5000,
		TotalReleases: 3,
		TotalRepos:    10,
		ByCategory: []CategoryFlow{
			{Category: "feat", Commits: 20, Insertions: 2000, Deletions: 500, Percentage: 40},
			{Category: "fix", Commits: 15, Insertions: 800, Deletions: 300, Percentage: 30},
			{Category: "chore", Commits: 10, Insertions: 400, Deletions: 200, Percentage: 20},
			{Category: "docs", Commits: 5, Insertions: 300, Deletions: 100, Percentage: 10},
		},
	}

	links := BuildSankeyData(flow)

	if len(links) == 0 {
		t.Fatal("expected some links, got none")
	}

	tokensToLOC := false
	locToFeat := false
	for _, link := range links {
		if link.Source == "Tokens" && link.Target == "LOC" {
			tokensToLOC = true
			if link.Value != 5000 {
				t.Errorf("Tokens→LOC: expected 5000, got %v", link.Value)
			}
		}
		if link.Source == "LOC" && link.Target == "feat" {
			locToFeat = true
			if link.Value != 2500 {
				t.Errorf("LOC→feat: expected 2500, got %v", link.Value)
			}
		}
	}

	if !tokensToLOC {
		t.Error("missing Tokens→LOC link")
	}
	if !locToFeat {
		t.Error("missing LOC→feat link")
	}
}

func TestBuildSankeyDataNil(t *testing.T) {
	links := BuildSankeyData(nil)
	if links != nil {
		t.Errorf("expected nil for nil flow, got %v", links)
	}
}

func TestExportSankeyDashboard(t *testing.T) {
	r := &Report{
		Username: "testuser",
		Year:     2026,
		Quarter:  2,
		Label:    "Q2 2026",
		SDLCFlow: &SDLCFlow{
			TotalLOC:      1000,
			TotalReleases: 2,
			TotalRepos:    5,
			ByCategory: []CategoryFlow{
				{Category: "feat", Insertions: 500, Deletions: 100},
				{Category: "fix", Insertions: 300, Deletions: 50},
			},
		},
	}

	dashboard, err := ExportSankeyDashboard(r)
	if err != nil {
		t.Fatalf("ExportSankeyDashboard: %v", err)
	}

	if dashboard.ID != "quarterly-2026-q2-sankey" {
		t.Errorf("ID: expected quarterly-2026-q2-sankey, got %s", dashboard.ID)
	}

	if len(dashboard.DataSources) != 1 {
		t.Errorf("expected 1 datasource, got %d", len(dashboard.DataSources))
	}

	if len(dashboard.Widgets) != 1 {
		t.Errorf("expected 1 widget, got %d", len(dashboard.Widgets))
	}

	var cfg struct {
		ChartType string `json:"chartType"`
	}
	if err := json.Unmarshal(dashboard.Widgets[0].Config, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if cfg.ChartType != "sankey" {
		t.Errorf("chartType: expected sankey, got %s", cfg.ChartType)
	}
}

func TestExportSankeyDashboardNilReport(t *testing.T) {
	_, err := ExportSankeyDashboard(nil)
	if err == nil {
		t.Error("expected error for nil report")
	}
}

func TestExportSankeyDashboardNoFlow(t *testing.T) {
	r := &Report{Username: "test", Year: 2026, Quarter: 2}
	_, err := ExportSankeyDashboard(r)
	if err == nil {
		t.Error("expected error for report with no SDLC flow")
	}
}
