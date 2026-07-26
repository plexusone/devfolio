package quarterly

import (
	"github.com/plexusone/omnidevx-core/report"
)

// TokenSpendFromReport extracts token spend summary from a DeveloperPeriodReport.
func TokenSpendFromReport(r *report.DeveloperPeriodReport) *TokenSpendSummary {
	if r == nil {
		return nil
	}

	ts := &TokenSpendSummary{
		ByModel:  make(map[string]ModelTokens),
		BySource: make(map[string]SourceTokens),
	}

	// Extract combined totals
	if m, ok := r.Metrics.Combined["input_tokens"]; ok {
		ts.TotalInputTokens = int64(m.Value)
	}
	if m, ok := r.Metrics.Combined["output_tokens"]; ok {
		ts.TotalOutputTokens = int64(m.Value)
	}
	if m, ok := r.Metrics.Combined["cache_read_tokens"]; ok {
		ts.TotalCacheRead = int64(m.Value)
	}
	if m, ok := r.Metrics.Combined["cache_creation_tokens"]; ok {
		ts.TotalCacheCreation = int64(m.Value)
	}
	if m, ok := r.Metrics.Combined["cost_usd"]; ok {
		ts.TotalCostUSD = m.Value
	}
	if m, ok := r.Metrics.Combined["cost_usd_estimated"]; ok {
		ts.EstimatedCostUSD = m.Value
	}

	// Extract per-model breakdown
	for model, metrics := range r.Metrics.ByModel {
		mt := ModelTokens{Model: model}
		if m, ok := metrics["input_tokens"]; ok {
			mt.InputTokens = int64(m.Value)
		}
		if m, ok := metrics["output_tokens"]; ok {
			mt.OutputTokens = int64(m.Value)
		}
		if m, ok := metrics["cost_usd"]; ok {
			mt.CostUSD = m.Value
		}
		ts.ByModel[model] = mt
	}

	// Extract per-source breakdown
	for source, metrics := range r.Metrics.BySource {
		st := SourceTokens{Source: source}
		if m, ok := metrics["input_tokens"]; ok {
			st.InputTokens = int64(m.Value)
		}
		if m, ok := metrics["output_tokens"]; ok {
			st.OutputTokens = int64(m.Value)
		}
		if m, ok := metrics["cost_usd"]; ok {
			st.CostUSD = m.Value
		}
		ts.BySource[source] = st
	}

	return ts
}

// ModelBreakdown returns models sorted by cost (descending).
func (ts *TokenSpendSummary) ModelBreakdown() []ModelTokens {
	if ts == nil || len(ts.ByModel) == 0 {
		return nil
	}

	models := make([]ModelTokens, 0, len(ts.ByModel))
	for _, mt := range ts.ByModel {
		models = append(models, mt)
	}

	// Sort by cost descending, then by total tokens
	for i := 0; i < len(models)-1; i++ {
		for j := i + 1; j < len(models); j++ {
			if models[j].CostUSD > models[i].CostUSD ||
				(models[j].CostUSD == models[i].CostUSD &&
					models[j].InputTokens+models[j].OutputTokens > models[i].InputTokens+models[i].OutputTokens) {
				models[i], models[j] = models[j], models[i]
			}
		}
	}

	return models
}

// TopModel returns the model with highest cost, or empty string if none.
func (ts *TokenSpendSummary) TopModel() string {
	breakdown := ts.ModelBreakdown()
	if len(breakdown) == 0 {
		return ""
	}
	return breakdown[0].Model
}
