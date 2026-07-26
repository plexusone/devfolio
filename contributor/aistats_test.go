package contributor

import (
	"testing"

	"github.com/grokify/gogit"
)

func TestAICollabStatsFromGogit(t *testing.T) {
	stats := gogit.AIStats{
		TotalCommits:    100,
		AIAssistedCount: 45,
		AIAssistedPct:   45.0,
		ByTool: map[string]gogit.ToolStats{
			"Claude Code": {Tool: "Claude Code", Commits: 40, Insertions: 5000, Deletions: 1000},
			"Cursor":      {Tool: "Cursor", Commits: 5, Insertions: 500, Deletions: 100},
		},
		ByModel: map[string]gogit.ModelStats{
			"Claude Code/Sonnet 4": {Tool: "Claude Code", Model: "Sonnet 4", Commits: 25},
			"Claude Code/Opus 4":   {Tool: "Claude Code", Model: "Opus 4", Commits: 15},
			"Cursor/gpt-4":         {Tool: "Cursor", Model: "gpt-4", Commits: 5},
		},
	}

	result := AICollabStatsFromGogit(stats, 100)

	if result.TotalAICommits != 45 {
		t.Errorf("TotalAICommits: expected 45, got %d", result.TotalAICommits)
	}

	if result.AICommitPercent != 45.0 {
		t.Errorf("AICommitPercent: expected 45.0, got %f", result.AICommitPercent)
	}

	if result.MostUsedTool != "Claude Code" {
		t.Errorf("MostUsedTool: expected Claude Code, got %s", result.MostUsedTool)
	}

	if result.MostUsedModel != "Claude Code/Sonnet 4" {
		t.Errorf("MostUsedModel: expected Claude Code/Sonnet 4, got %s", result.MostUsedModel)
	}

	// Check tool stats
	claude, ok := result.ByTool["Claude Code"]
	if !ok {
		t.Fatal("missing Claude Code in ByTool")
	}
	if claude.Commits != 40 {
		t.Errorf("Claude Code commits: expected 40, got %d", claude.Commits)
	}
	if !claude.Recognized {
		t.Error("Claude Code should be recognized")
	}

	// Check model stats within tool
	sonnet, ok := claude.ByModel["Sonnet 4"]
	if !ok {
		t.Fatal("missing Sonnet 4 in Claude Code ByModel")
	}
	if sonnet.Commits != 25 {
		t.Errorf("Sonnet 4 commits: expected 25, got %d", sonnet.Commits)
	}
}

func TestAICollabStatsFromGogitEmpty(t *testing.T) {
	stats := gogit.AIStats{
		ByTool:  make(map[string]gogit.ToolStats),
		ByModel: make(map[string]gogit.ModelStats),
	}

	result := AICollabStatsFromGogit(stats, 100)

	if result.TotalAICommits != 0 {
		t.Errorf("TotalAICommits: expected 0, got %d", result.TotalAICommits)
	}

	if result.MostUsedTool != "" {
		t.Errorf("MostUsedTool: expected empty, got %s", result.MostUsedTool)
	}
}

func TestAICollabStatsFromGogitZeroTotal(t *testing.T) {
	stats := gogit.AIStats{
		AIAssistedCount: 5,
		ByTool:          make(map[string]gogit.ToolStats),
		ByModel:         make(map[string]gogit.ModelStats),
	}

	result := AICollabStatsFromGogit(stats, 0)

	if result.AICommitPercent != 0 {
		t.Errorf("AICommitPercent: expected 0 (no divide by zero), got %f", result.AICommitPercent)
	}
}
