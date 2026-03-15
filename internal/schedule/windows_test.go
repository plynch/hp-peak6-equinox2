package schedule

import (
	"testing"
	"time"
)

func TestKeepUpcomingWindowsGroupsByGap(t *testing.T) {
	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	items := []time.Time{
		base,
		base.Add(24 * time.Hour),
		base.Add(6 * 24 * time.Hour),
		base.Add(7 * 24 * time.Hour),
		base.Add(13 * 24 * time.Hour),
	}

	selected := KeepUpcomingWindows(items, func(v time.Time) *time.Time { return &v }, 2, 72*time.Hour)
	if len(selected) != 4 {
		t.Fatalf("expected 4 items in first two windows, got %d", len(selected))
	}
	if !selected[0].Equal(items[0]) || !selected[3].Equal(items[3]) {
		t.Fatalf("unexpected selection: %#v", selected)
	}
}

func TestKeepUpcomingWindowsHandlesUnlimited(t *testing.T) {
	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	items := []time.Time{
		base,
		base.Add(24 * time.Hour),
		base.Add(8 * 24 * time.Hour),
	}

	selected := KeepUpcomingWindows(items, func(v time.Time) *time.Time { return &v }, 0, 72*time.Hour)
	if len(selected) != len(items) {
		t.Fatalf("expected all items when windowLimit=0, got %d", len(selected))
	}
}
