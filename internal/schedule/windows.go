package schedule

import (
	"sort"
	"time"
)

// KeepUpcomingWindows retains the earliest upcoming groups of items, where a new
// group starts when the gap between consecutive item times exceeds maxGap.
// This is used as a matchweek-style approximation for sports feeds that do not
// expose an official matchweek number in their public APIs.
func KeepUpcomingWindows[T any](items []T, timeFn func(T) *time.Time, windowLimit int, maxGap time.Duration) []T {
	type timedItem struct {
		item T
		when time.Time
	}

	timed := make([]timedItem, 0, len(items))
	for _, item := range items {
		when := timeFn(item)
		if when == nil {
			continue
		}
		timed = append(timed, timedItem{item: item, when: *when})
	}

	sort.Slice(timed, func(i, j int) bool {
		return timed[i].when.Before(timed[j].when)
	})

	if windowLimit <= 0 || len(timed) == 0 {
		out := make([]T, 0, len(timed))
		for _, item := range timed {
			out = append(out, item.item)
		}
		return out
	}

	out := make([]T, 0, len(timed))
	windowCount := 0
	var last time.Time
	for i, item := range timed {
		if i == 0 || item.when.Sub(last) > maxGap {
			windowCount++
			if windowCount > windowLimit {
				break
			}
		}
		out = append(out, item.item)
		last = item.when
	}
	return out
}
