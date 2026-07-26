package contributor

import (
	"github.com/grokify/gogit"
)

// AICollabStatsFromGogit converts gogit.AIStats to contributor.AICollabStats.
// This bridges the low-level commit aggregator with the profile-level view.
func AICollabStatsFromGogit(stats gogit.AIStats, totalCommits int) AICollabStats {
	result := AICollabStats{
		TotalAICommits: stats.AIAssistedCount,
		ByTool:         make(map[string]AIToolStat),
	}

	if totalCommits > 0 {
		result.AICommitPercent = float64(stats.AIAssistedCount) / float64(totalCommits) * 100
	}

	// Build per-tool stats
	for toolName, ts := range stats.ByTool {
		toolStat := AIToolStat{
			Name:       toolName,
			Commits:    ts.Commits,
			Recognized: isRecognizedTool(toolName),
			ByModel:    make(map[string]AIModelStat),
		}
		result.ByTool[toolName] = toolStat
	}

	// Populate per-model within each tool
	for _, ms := range stats.ByModel {
		if toolStat, ok := result.ByTool[ms.Tool]; ok {
			toolStat.ByModel[ms.Model] = AIModelStat{
				Model:   ms.Model,
				Commits: ms.Commits,
			}
			result.ByTool[ms.Tool] = toolStat
		}
	}

	// Find most used tool and model
	var maxToolCommits, maxModelCommits int
	for toolName, ts := range result.ByTool {
		if ts.Commits > maxToolCommits {
			maxToolCommits = ts.Commits
			result.MostUsedTool = toolName
		}
		for modelName, ms := range ts.ByModel {
			if ms.Commits > maxModelCommits {
				maxModelCommits = ms.Commits
				result.MostUsedModel = toolName + "/" + modelName
			}
		}
	}

	return result
}

// isRecognizedTool returns true if GitHub recognizes the tool's co-author email.
func isRecognizedTool(name string) bool {
	for _, tool := range KnownAITools {
		if tool.Name == name {
			return tool.Recognized
		}
	}
	return false
}
