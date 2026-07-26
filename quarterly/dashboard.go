package quarterly

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/plexusone/uiforge/dashboardir"
)

// ExportDashboard converts a quarterly Report into a uiforge Dashboard IR.
// The dashboard includes metric tiles, category breakdown charts, AI stats,
// and token spend visualization.
func ExportDashboard(r *Report) (*dashboardir.Dashboard, error) {
	if r == nil {
		return nil, fmt.Errorf("quarterly: report is nil")
	}

	dataSources := buildDataSources(r)
	widgets := buildDashboardWidgets(r)

	return &dashboardir.Dashboard{
		ID:          fmt.Sprintf("quarterly-%s-%s", r.Username, r.Label),
		Title:       fmt.Sprintf("Quarterly Report - %s", r.Label),
		Description: fmt.Sprintf("Developer report for %s (%s)", r.Username, r.Label),
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
		Widgets:     widgets,
	}, nil
}

func buildDataSources(r *Report) []dashboardir.DataSource {
	var ds []dashboardir.DataSource

	// Summary metrics
	summaryData := map[string]any{
		"commits":       0,
		"additions":     0,
		"deletions":     0,
		"netLines":      0,
		"releases":      0,
		"repos":         0,
		"prs":           0,
		"reviews":       0,
		"aiAssisted":    0,
		"aiAssistedPct": 0.0,
	}

	if r.GitHubStats != nil {
		summaryData["commits"] = r.GitHubStats.Commits
		summaryData["additions"] = r.GitHubStats.Additions
		summaryData["deletions"] = r.GitHubStats.Deletions
		summaryData["netLines"] = r.GitHubStats.NetAdditions
		summaryData["releases"] = r.GitHubStats.Releases
		summaryData["repos"] = r.GitHubStats.RepoCountContributed + r.GitHubStats.RepoCountCreated
		summaryData["prs"] = r.GitHubStats.PRs
		summaryData["reviews"] = r.GitHubStats.Reviews
	}

	if r.CommitStats != nil {
		summaryData["aiAssisted"] = r.CommitStats.AIStats.AIAssistedCount
		summaryData["aiAssistedPct"] = r.CommitStats.AIStats.AIAssistedPct
	}

	summaryJSON, _ := json.Marshal(summaryData)
	ds = append(ds, dashboardir.DataSource{
		ID:   "summary",
		Name: "Summary Metrics",
		Type: dashboardir.DataSourceTypeInline,
		Data: summaryJSON,
	})

	// Category breakdown
	if r.CommitStats != nil && len(r.CommitStats.ByCategory) > 0 {
		var categories []map[string]any
		for cat, stats := range r.CommitStats.ByCategory {
			pct := 0.0
			if r.CommitStats.TotalStats.Commits > 0 {
				pct = float64(stats.Commits) / float64(r.CommitStats.TotalStats.Commits) * 100
			}
			categories = append(categories, map[string]any{
				"category":   cat,
				"commits":    stats.Commits,
				"insertions": stats.Insertions,
				"deletions":  stats.Deletions,
				"percentage": pct,
			})
		}
		// Sort by commits descending
		sort.Slice(categories, func(i, j int) bool {
			return categories[i]["commits"].(int) > categories[j]["commits"].(int)
		})
		catJSON, _ := json.Marshal(categories)
		ds = append(ds, dashboardir.DataSource{
			ID:   "categories",
			Name: "Commit Categories",
			Type: dashboardir.DataSourceTypeInline,
			Data: catJSON,
		})
	}

	// AI model breakdown
	if r.CommitStats != nil && len(r.CommitStats.AIStats.ByModel) > 0 {
		var models []map[string]any
		for key, stats := range r.CommitStats.AIStats.ByModel {
			models = append(models, map[string]any{
				"model":      key,
				"commits":    stats.Commits,
				"insertions": stats.Insertions,
				"deletions":  stats.Deletions,
			})
		}
		sort.Slice(models, func(i, j int) bool {
			return models[i]["commits"].(int) > models[j]["commits"].(int)
		})
		modelJSON, _ := json.Marshal(models)
		ds = append(ds, dashboardir.DataSource{
			ID:   "ai-models",
			Name: "AI Models",
			Type: dashboardir.DataSourceTypeInline,
			Data: modelJSON,
		})
	}

	// Token spend
	if r.TokenSpend != nil {
		tokenData := map[string]any{
			"inputTokens":   r.TokenSpend.TotalInputTokens,
			"outputTokens":  r.TokenSpend.TotalOutputTokens,
			"cacheRead":     r.TokenSpend.TotalCacheRead,
			"cacheCreation": r.TokenSpend.TotalCacheCreation,
			"totalCost":     r.TokenSpend.TotalCostUSD,
			"estimatedCost": r.TokenSpend.EstimatedCostUSD,
		}
		tokenJSON, _ := json.Marshal(tokenData)
		ds = append(ds, dashboardir.DataSource{
			ID:   "tokens",
			Name: "Token Spend",
			Type: dashboardir.DataSourceTypeInline,
			Data: tokenJSON,
		})

		// Per-model token breakdown
		if len(r.TokenSpend.ByModel) > 0 {
			var modelTokens []map[string]any
			for model, mt := range r.TokenSpend.ByModel {
				modelTokens = append(modelTokens, map[string]any{
					"model":        model,
					"inputTokens":  mt.InputTokens,
					"outputTokens": mt.OutputTokens,
					"cost":         mt.CostUSD,
				})
			}
			sort.Slice(modelTokens, func(i, j int) bool {
				return modelTokens[i]["cost"].(float64) > modelTokens[j]["cost"].(float64)
			})
			mtJSON, _ := json.Marshal(modelTokens)
			ds = append(ds, dashboardir.DataSource{
				ID:   "token-models",
				Name: "Token Spend by Model",
				Type: dashboardir.DataSourceTypeInline,
				Data: mtJSON,
			})
		}
	}

	return ds
}

func buildDashboardWidgets(r *Report) []dashboardir.Widget {
	var widgets []dashboardir.Widget
	row := 0

	// Row 1: Summary metrics (6 tiles, 2 columns each)
	widgets = append(widgets,
		metricWidget("commits", "Commits", pos(0, row, 2, 2), "summary", "commits"),
		metricWidget("net-lines", "Net Lines", pos(2, row, 2, 2), "summary", "netLines"),
		metricWidget("releases", "Releases", pos(4, row, 2, 2), "summary", "releases"),
		metricWidget("repos", "Repos", pos(6, row, 2, 2), "summary", "repos"),
		metricWidget("ai-assisted", "AI-Assisted", pos(8, row, 2, 2), "summary", "aiAssisted"),
		metricWidget("ai-pct", "AI %", pos(10, row, 2, 2), "summary", "aiAssistedPct"),
	)
	row += 2

	// Row 2: Category breakdown (bar chart)
	if r.CommitStats != nil && len(r.CommitStats.ByCategory) > 0 {
		widgets = append(widgets, barChartWidget("category-chart", "Commit Categories",
			pos(0, row, 6, 5), "categories", "category", "commits"))
		widgets = append(widgets, pieChartWidget("category-pie", "Category Distribution",
			pos(6, row, 6, 5), "categories", "category", "commits"))
		row += 5
	}

	// Row 3: AI model breakdown
	if r.CommitStats != nil && len(r.CommitStats.AIStats.ByModel) > 0 {
		widgets = append(widgets, barChartWidget("ai-model-chart", "Commits by AI Model",
			pos(0, row, 6, 5), "ai-models", "model", "commits"))
		row += 5
	}

	// Row 4: Token spend
	if r.TokenSpend != nil {
		widgets = append(widgets,
			metricWidget("input-tokens", "Input Tokens", pos(0, row, 3, 2), "tokens", "inputTokens"),
			metricWidget("output-tokens", "Output Tokens", pos(3, row, 3, 2), "tokens", "outputTokens"),
			metricWidget("total-cost", "Total Cost", pos(6, row, 3, 2), "tokens", "totalCost"),
		)
		row += 2

		if len(r.TokenSpend.ByModel) > 0 {
			widgets = append(widgets, barChartWidget("token-model-chart", "Cost by Model",
				pos(0, row, 12, 5), "token-models", "model", "cost"))
			row += 5
		}
	}

	return widgets
}

func pos(x, y, w, h int) dashboardir.Position {
	return dashboardir.Position{X: x, Y: y, W: w, H: h}
}

func metricWidget(id, title string, p dashboardir.Position, dataSourceID, valueField string) dashboardir.Widget {
	cfg, _ := json.Marshal(dashboardir.MetricConfig{
		ValueField: valueField,
		Format:     "number",
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

func barChartWidget(id, title string, p dashboardir.Position, dataSourceID, xField, yField string) dashboardir.Widget {
	cfg := struct {
		Marks []struct {
			ID     string            `json:"id"`
			Geom   string            `json:"geometry"`
			Encode map[string]string `json:"encode"`
		} `json:"marks"`
		Axes []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Position string `json:"position"`
		} `json:"axes"`
	}{
		Marks: []struct {
			ID     string            `json:"id"`
			Geom   string            `json:"geometry"`
			Encode map[string]string `json:"encode"`
		}{
			{ID: "bar-1", Geom: "bar", Encode: map[string]string{"x": xField, "y": yField}},
		},
		Axes: []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Position string `json:"position"`
		}{
			{ID: "x", Type: "category", Position: "bottom"},
			{ID: "y", Type: "value", Position: "left"},
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

func pieChartWidget(id, title string, p dashboardir.Position, dataSourceID, nameField, valueField string) dashboardir.Widget {
	cfg := struct {
		Marks []struct {
			ID     string            `json:"id"`
			Geom   string            `json:"geometry"`
			Encode map[string]string `json:"encode"`
		} `json:"marks"`
	}{
		Marks: []struct {
			ID     string            `json:"id"`
			Geom   string            `json:"geometry"`
			Encode map[string]string `json:"encode"`
		}{
			{ID: "pie-1", Geom: "pie", Encode: map[string]string{"name": nameField, "value": valueField}},
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
