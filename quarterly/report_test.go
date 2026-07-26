package quarterly

import (
	"testing"
	"time"
)

func TestQuarterDates(t *testing.T) {
	tests := []struct {
		year, quarter int
		wantSince     time.Time
		wantUntil     time.Time
	}{
		{2026, 1, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)},
		{2026, 2, time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
		{2026, 3, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
		{2026, 4, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		since, until := QuarterDates(tt.year, tt.quarter)
		if !since.Equal(tt.wantSince) {
			t.Errorf("Q%d %d: since = %v, want %v", tt.quarter, tt.year, since, tt.wantSince)
		}
		if !until.Equal(tt.wantUntil) {
			t.Errorf("Q%d %d: until = %v, want %v", tt.quarter, tt.year, until, tt.wantUntil)
		}
	}
}

func TestReportCategoryPercentages_Nil(t *testing.T) {
	r := &Report{}
	pcts := r.CategoryPercentages()
	if pcts != nil {
		t.Errorf("expected nil for report with no commit stats, got %v", pcts)
	}
}

func TestReportTopProjects_Nil(t *testing.T) {
	r := &Report{}
	top := r.TopProjects(5)
	if top != nil {
		t.Errorf("expected nil for report with no highlights, got %v", top)
	}
}
