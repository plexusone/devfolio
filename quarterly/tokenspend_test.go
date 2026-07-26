package quarterly

import (
	"testing"

	"github.com/plexusone/omnidevx-core/report"
)

func TestTokenSpendFromReport(t *testing.T) {
	r := &report.DeveloperPeriodReport{
		Metrics: report.MetricSet{
			Combined: map[string]report.Metric{
				"input_tokens":          {Value: 1000000},
				"output_tokens":         {Value: 500000},
				"cache_read_tokens":     {Value: 200000},
				"cache_creation_tokens": {Value: 100000},
				"cost_usd":              {Value: 25.50},
				"cost_usd_estimated":    {Value: 5.00},
			},
			ByModel: map[string]map[string]report.Metric{
				"claude-sonnet-5": {
					"input_tokens":  {Value: 600000},
					"output_tokens": {Value: 300000},
					"cost_usd":      {Value: 15.00},
				},
				"claude-opus-4-5": {
					"input_tokens":  {Value: 400000},
					"output_tokens": {Value: 200000},
					"cost_usd":      {Value: 10.50},
				},
			},
			BySource: map[string]map[string]report.Metric{
				"anthropic/claude-code": {
					"input_tokens":  {Value: 1000000},
					"output_tokens": {Value: 500000},
					"cost_usd":      {Value: 25.50},
				},
			},
		},
	}

	ts := TokenSpendFromReport(r)

	if ts.TotalInputTokens != 1000000 {
		t.Errorf("TotalInputTokens: expected 1000000, got %d", ts.TotalInputTokens)
	}
	if ts.TotalOutputTokens != 500000 {
		t.Errorf("TotalOutputTokens: expected 500000, got %d", ts.TotalOutputTokens)
	}
	if ts.TotalCostUSD != 25.50 {
		t.Errorf("TotalCostUSD: expected 25.50, got %f", ts.TotalCostUSD)
	}
	if ts.EstimatedCostUSD != 5.00 {
		t.Errorf("EstimatedCostUSD: expected 5.00, got %f", ts.EstimatedCostUSD)
	}

	// Check ByModel
	if len(ts.ByModel) != 2 {
		t.Errorf("ByModel: expected 2 models, got %d", len(ts.ByModel))
	}
	sonnet, ok := ts.ByModel["claude-sonnet-5"]
	if !ok {
		t.Fatal("missing claude-sonnet-5 in ByModel")
	}
	if sonnet.InputTokens != 600000 {
		t.Errorf("sonnet InputTokens: expected 600000, got %d", sonnet.InputTokens)
	}
	if sonnet.CostUSD != 15.00 {
		t.Errorf("sonnet CostUSD: expected 15.00, got %f", sonnet.CostUSD)
	}

	// Check BySource
	if len(ts.BySource) != 1 {
		t.Errorf("BySource: expected 1 source, got %d", len(ts.BySource))
	}
}

func TestTokenSpendFromReportNil(t *testing.T) {
	ts := TokenSpendFromReport(nil)
	if ts != nil {
		t.Error("expected nil for nil report")
	}
}

func TestModelBreakdown(t *testing.T) {
	ts := &TokenSpendSummary{
		ByModel: map[string]ModelTokens{
			"claude-sonnet-5": {Model: "claude-sonnet-5", InputTokens: 600000, OutputTokens: 300000, CostUSD: 15.00},
			"claude-opus-4-5": {Model: "claude-opus-4-5", InputTokens: 400000, OutputTokens: 200000, CostUSD: 20.00},
			"claude-haiku-4-5": {Model: "claude-haiku-4-5", InputTokens: 100000, OutputTokens: 50000, CostUSD: 1.00},
		},
	}

	breakdown := ts.ModelBreakdown()

	if len(breakdown) != 3 {
		t.Fatalf("expected 3 models, got %d", len(breakdown))
	}

	// Should be sorted by cost descending
	if breakdown[0].Model != "claude-opus-4-5" {
		t.Errorf("first model should be opus (highest cost), got %s", breakdown[0].Model)
	}
	if breakdown[1].Model != "claude-sonnet-5" {
		t.Errorf("second model should be sonnet, got %s", breakdown[1].Model)
	}
	if breakdown[2].Model != "claude-haiku-4-5" {
		t.Errorf("third model should be haiku (lowest cost), got %s", breakdown[2].Model)
	}
}

func TestTopModel(t *testing.T) {
	ts := &TokenSpendSummary{
		ByModel: map[string]ModelTokens{
			"claude-sonnet-5": {Model: "claude-sonnet-5", CostUSD: 15.00},
			"claude-opus-4-5": {Model: "claude-opus-4-5", CostUSD: 20.00},
		},
	}

	top := ts.TopModel()
	if top != "claude-opus-4-5" {
		t.Errorf("TopModel: expected claude-opus-4-5, got %s", top)
	}
}

func TestTopModelEmpty(t *testing.T) {
	ts := &TokenSpendSummary{}
	top := ts.TopModel()
	if top != "" {
		t.Errorf("TopModel: expected empty string for empty ByModel, got %s", top)
	}
}

func TestTopModelNil(t *testing.T) {
	var ts *TokenSpendSummary
	top := ts.TopModel()
	if top != "" {
		t.Errorf("TopModel: expected empty string for nil, got %s", top)
	}
}
