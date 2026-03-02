package dashboard

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/plexusone/devfolio/contributor"
)

// ExportContributorDashboard converts a contributor Profile to a dashforge Dashboard.
func ExportContributorDashboard(profile *contributor.Profile) (*Dashboard, error) {
	dashboard := &Dashboard{
		ID:          fmt.Sprintf("contributor-%s", profile.Username),
		Title:       fmt.Sprintf("%s - Contributor Profile", profile.Name),
		Description: fmt.Sprintf("Development activity profile for %s", profile.Username),
		Version:     "1.0.0",
		Layout: Layout{
			Type:      LayoutTypeGrid,
			Columns:   12,
			RowHeight: 60,
			Gap:       16,
			Padding:   16,
		},
		Theme: &Theme{
			Mode: "light",
		},
	}

	// Build data sources
	dataSources, err := buildContributorDataSources(profile)
	if err != nil {
		return nil, fmt.Errorf("building data sources: %w", err)
	}
	dashboard.DataSources = dataSources

	// Build widgets
	dashboard.Widgets = buildContributorWidgets(profile)

	return dashboard, nil
}

func buildContributorDataSources(profile *contributor.Profile) ([]DataSource, error) {
	var dataSources []DataSource

	// Summary data source with key metrics
	summaryData := map[string]any{
		"username":        profile.Username,
		"name":            profile.Name,
		"totalCommits":    profile.Stats.TotalCommits,
		"totalPRs":        profile.Stats.TotalPRs,
		"totalRepos":      profile.Stats.TotalRepositories,
		"ownedRepos":      profile.Stats.OwnedRepos,
		"contributedRepos": profile.Stats.ContributedRepos,
		"aiCommits":       profile.AIStats.TotalAICommits,
		"aiPercent":       profile.AIStats.AICommitPercent,
		"mostUsedTool":    profile.AIStats.MostUsedTool,
	}
	summaryJSON, err := json.Marshal(summaryData)
	if err != nil {
		return nil, err
	}
	dataSources = append(dataSources, DataSource{
		ID:   "summary",
		Name: "Summary Metrics",
		Type: DataSourceTypeInline,
		Data: summaryJSON,
	})

	// Activity heatmap data
	activityData := make([]map[string]any, 0, len(profile.Activity))
	for _, a := range profile.Activity {
		activityData = append(activityData, map[string]any{
			"date":  a.Date,
			"count": a.Count,
		})
	}
	activityJSON, err := json.Marshal(activityData)
	if err != nil {
		return nil, err
	}
	dataSources = append(dataSources, DataSource{
		ID:   "activity",
		Name: "Activity Heatmap",
		Type: DataSourceTypeInline,
		Data: activityJSON,
	})

	// Language breakdown data
	languageData := make([]map[string]any, 0, len(profile.Languages))
	for lang, count := range profile.Languages {
		languageData = append(languageData, map[string]any{
			"language": lang,
			"count":    count,
		})
	}
	// Sort by count descending
	sort.Slice(languageData, func(i, j int) bool {
		return languageData[i]["count"].(int) > languageData[j]["count"].(int)
	})
	languageJSON, err := json.Marshal(languageData)
	if err != nil {
		return nil, err
	}
	dataSources = append(dataSources, DataSource{
		ID:   "languages",
		Name: "Languages",
		Type: DataSourceTypeInline,
		Data: languageJSON,
	})

	// AI tools data
	aiToolsData := make([]map[string]any, 0, len(profile.AIStats.ByTool))
	for name, stat := range profile.AIStats.ByTool {
		aiToolsData = append(aiToolsData, map[string]any{
			"tool":       name,
			"commits":    stat.Commits,
			"firstUsed":  stat.FirstUsed,
			"lastUsed":   stat.LastUsed,
			"recognized": stat.Recognized,
		})
	}
	// Sort by commits descending
	sort.Slice(aiToolsData, func(i, j int) bool {
		return aiToolsData[i]["commits"].(int) > aiToolsData[j]["commits"].(int)
	})
	aiToolsJSON, err := json.Marshal(aiToolsData)
	if err != nil {
		return nil, err
	}
	dataSources = append(dataSources, DataSource{
		ID:   "aiTools",
		Name: "AI Tools",
		Type: DataSourceTypeInline,
		Data: aiToolsJSON,
	})

	// Repositories data
	reposData := make([]map[string]any, 0, len(profile.Repositories))
	for _, repo := range profile.Repositories {
		reposData = append(reposData, map[string]any{
			"name":        fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
			"description": repo.Description,
			"commits":     repo.Commits,
			"prs":         repo.PRs,
			"language":    repo.Language,
			"stars":       repo.Stars,
			"url":         repo.URL,
			"isOwner":     repo.IsOwner,
		})
	}
	reposJSON, err := json.Marshal(reposData)
	if err != nil {
		return nil, err
	}
	dataSources = append(dataSources, DataSource{
		ID:   "repos",
		Name: "Repositories",
		Type: DataSourceTypeInline,
		Data: reposJSON,
	})

	return dataSources, nil
}

func buildContributorWidgets(profile *contributor.Profile) []Widget {
	var widgets []Widget

	// Row 1: Summary metrics (4 metrics across 12 columns)
	widgets = append(widgets, buildMetricWidget("total-commits", "Total Commits", "summary", "totalCommits", Position{X: 0, Y: 0, W: 3, H: 2}))
	widgets = append(widgets, buildMetricWidget("total-prs", "Pull Requests", "summary", "totalPRs", Position{X: 3, Y: 0, W: 3, H: 2}))
	widgets = append(widgets, buildMetricWidget("total-repos", "Repositories", "summary", "totalRepos", Position{X: 6, Y: 0, W: 3, H: 2}))
	widgets = append(widgets, buildAIPercentMetric(profile, Position{X: 9, Y: 0, W: 3, H: 2}))

	// Row 2: Activity heatmap (full width)
	widgets = append(widgets, buildActivityHeatmap(profile, Position{X: 0, Y: 2, W: 12, H: 4}))

	// Row 3: Language pie chart and AI tools bar chart
	widgets = append(widgets, buildLanguagePieChart(Position{X: 0, Y: 6, W: 6, H: 5}))
	widgets = append(widgets, buildAIToolsBarChart(Position{X: 6, Y: 6, W: 6, H: 5}))

	// Row 4: Repository table (full width)
	widgets = append(widgets, buildRepoTable(Position{X: 0, Y: 11, W: 12, H: 6}))

	return widgets
}

func buildMetricWidget(id, title, dataSourceID, valueField string, pos Position) Widget {
	config := MetricConfig{
		ValueField: valueField,
		Format:     "number",
	}
	configJSON, _ := json.Marshal(config)

	return Widget{
		ID:           id,
		Title:        title,
		Type:         WidgetTypeMetric,
		Position:     pos,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}
}

func buildAIPercentMetric(profile *contributor.Profile, pos Position) Widget {
	config := MetricConfig{
		ValueField: "aiPercent",
		Format:     "number",
		FormatOptions: &FormatOptions{
			Decimals: 1,
			Suffix:   "%",
		},
	}
	configJSON, _ := json.Marshal(config)

	title := "AI Commits"
	if profile.AIStats.MostUsedTool != "" {
		title = fmt.Sprintf("AI Commits (%s)", profile.AIStats.MostUsedTool)
	}

	return Widget{
		ID:           "ai-percent",
		Title:        title,
		Type:         WidgetTypeMetric,
		Position:     pos,
		DataSourceID: "summary",
		Config:       configJSON,
	}
}

func buildActivityHeatmap(profile *contributor.Profile, pos Position) Widget {
	// Determine year range for calendar
	year := time.Now().Year()
	if profile.DateRange.Start != "" {
		if t, err := time.Parse("2006-01-02", profile.DateRange.Start); err == nil {
			year = t.Year()
		}
	}

	chart := ChartIR{
		Mark: Mark{Type: ChartTypeHeatmap},
		Encodings: Encodings{
			X:    Encoding{Field: "date", Type: "temporal"},
			Heat: Encoding{Field: "count", Type: "quantitative"},
		},
		Style: &Style{
			Calendar: &CalendarStyle{
				Range:    fmt.Sprintf("%d", year),
				CellSize: []any{"auto", 13},
				Colors:   []string{"#ebedf0", "#9be9a8", "#40c463", "#30a14e", "#216e39"},
			},
		},
	}
	configJSON, _ := json.Marshal(chart)

	return Widget{
		ID:           "activity-heatmap",
		Title:        "Activity",
		Type:         WidgetTypeChart,
		Position:     pos,
		DataSourceID: "activity",
		Config:       configJSON,
	}
}

func buildLanguagePieChart(pos Position) Widget {
	chart := ChartIR{
		Mark: Mark{Type: ChartTypePie},
		Encodings: Encodings{
			Value:    Encoding{Field: "count", Type: "quantitative"},
			Category: Encoding{Field: "language", Type: "nominal"},
		},
		Style: &Style{
			Legend: &Legend{Show: true, Position: "right"},
		},
	}
	configJSON, _ := json.Marshal(chart)

	return Widget{
		ID:           "language-breakdown",
		Title:        "Languages",
		Type:         WidgetTypeChart,
		Position:     pos,
		DataSourceID: "languages",
		Config:       configJSON,
	}
}

func buildAIToolsBarChart(pos Position) Widget {
	chart := ChartIR{
		Mark: Mark{
			Type: ChartTypeBar,
			Style: &MarkStyle{
				BorderRadius: 4,
			},
		},
		Encodings: Encodings{
			X: Encoding{Field: "tool", Type: "nominal"},
			Y: Encoding{Field: "commits", Type: "quantitative", Title: "Commits"},
		},
		Style: &Style{
			Colors: []string{"#8b5cf6", "#06b6d4", "#10b981", "#f59e0b", "#ef4444"},
			XAxis:  &Axis{Show: true, LabelRotate: -45},
			YAxis:  &Axis{Show: true, Title: "Commits"},
		},
	}
	configJSON, _ := json.Marshal(chart)

	return Widget{
		ID:           "ai-tools",
		Title:        "AI Tool Usage",
		Type:         WidgetTypeChart,
		Position:     pos,
		DataSourceID: "aiTools",
		Config:       configJSON,
	}
}

func buildRepoTable(pos Position) Widget {
	config := TableConfig{
		Columns: []TableColumn{
			{Field: "name", Header: "Repository", Width: "30%"},
			{Field: "commits", Header: "Commits", Width: "12%", Align: "right", Format: "number"},
			{Field: "prs", Header: "PRs", Width: "10%", Align: "right", Format: "number"},
			{Field: "language", Header: "Language", Width: "15%"},
			{Field: "stars", Header: "Stars", Width: "10%", Align: "right", Format: "number"},
			{Field: "description", Header: "Description", Width: "23%"},
		},
		Sortable: true,
		Compact:  true,
	}
	configJSON, _ := json.Marshal(config)

	return Widget{
		ID:           "repo-table",
		Title:        "Repositories",
		Type:         WidgetTypeTable,
		Position:     pos,
		DataSourceID: "repos",
		Config:       configJSON,
	}
}
