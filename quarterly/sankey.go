package quarterly

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/plexusone/uiforge/dashboardir"
)

// SankeyLink is one edge in the SDLC Sankey diagram.
type SankeyLink struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Value  float64 `json:"value"`
}

// BuildSankeyData converts SDLCFlow into Sankey links.
// The flow is: Inputs (tokens) → Outputs (LOC/releases) → Categories
func BuildSankeyData(flow *SDLCFlow) []SankeyLink {
	if flow == nil {
		return nil
	}

	var links []SankeyLink

	if flow.TotalLOC > 0 {
		links = append(links, SankeyLink{
			Source: "Tokens",
			Target: "LOC",
			Value:  float64(flow.TotalLOC),
		})
	}

	if flow.TotalReleases > 0 {
		links = append(links, SankeyLink{
			Source: "Tokens",
			Target: "Releases",
			Value:  float64(flow.TotalReleases * 1000),
		})
	}

	if flow.TotalRepos > 0 {
		links = append(links, SankeyLink{
			Source: "Tokens",
			Target: "Repos",
			Value:  float64(flow.TotalRepos * 500),
		})
	}

	sorted := make([]CategoryFlow, len(flow.ByCategory))
	copy(sorted, flow.ByCategory)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Commits > sorted[j].Commits
	})

	for _, cat := range sorted {
		loc := cat.Insertions + cat.Deletions
		if loc > 0 {
			links = append(links, SankeyLink{
				Source: "LOC",
				Target: cat.Category,
				Value:  float64(loc),
			})
		}
	}

	return links
}

// ExportSankeyDashboard creates a dashboard with the SDLC Sankey diagram.
func ExportSankeyDashboard(r *Report) (*dashboardir.Dashboard, error) {
	if r == nil {
		return nil, fmt.Errorf("quarterly: report is nil")
	}

	links := BuildSankeyData(r.SDLCFlow)
	if len(links) == 0 {
		return nil, fmt.Errorf("quarterly: no SDLC flow data")
	}

	inlineData, _ := json.Marshal(links)

	ds := dashboardir.DataSource{
		ID:   "sdlc-flow",
		Name: "SDLC Flow",
		Type: dashboardir.DataSourceTypeInline,
		Data: inlineData,
	}

	sankeyConfig := struct {
		ChartType string `json:"chartType"`
		Marks     []struct {
			ID       string            `json:"id"`
			Geometry string            `json:"geometry"`
			Encode   map[string]string `json:"encode"`
		} `json:"marks"`
	}{
		ChartType: "sankey",
		Marks: []struct {
			ID       string            `json:"id"`
			Geometry string            `json:"geometry"`
			Encode   map[string]string `json:"encode"`
		}{
			{
				ID:       "sankey-1",
				Geometry: "sankey",
				Encode: map[string]string{
					"source": "source",
					"target": "target",
					"value":  "value",
				},
			},
		},
	}
	configJSON, _ := json.Marshal(sankeyConfig)

	sankeyWidget := dashboardir.Widget{
		ID:    "sdlc-sankey",
		Title: fmt.Sprintf("SDLC Flow - %s", r.Label),
		Type:  dashboardir.WidgetTypeChart,
		Position: dashboardir.Position{
			X: 0,
			Y: 0,
			W: 12,
			H: 8,
		},
		DataSourceID: "sdlc-flow",
		Config:       configJSON,
	}

	return &dashboardir.Dashboard{
		ID:          fmt.Sprintf("quarterly-%d-q%d-sankey", r.Year, r.Quarter),
		Title:       fmt.Sprintf("Quarterly Report - %s - SDLC Flow", r.Label),
		Description: fmt.Sprintf("SDLC Sankey diagram for %s (%s)", r.Username, r.Label),
		Version:     "1.0.0",
		Layout: dashboardir.Layout{
			Type:      dashboardir.LayoutTypeGrid,
			Columns:   12,
			RowHeight: 64,
			Gap:       16,
			Padding:   16,
		},
		Theme:       &dashboardir.Theme{Mode: "light"},
		DataSources: []dashboardir.DataSource{ds},
		Widgets:     []dashboardir.Widget{sankeyWidget},
	}, nil
}
