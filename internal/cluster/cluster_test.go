package cluster

import (
	"testing"
	"time"

	"equinox/internal/model"
)

func TestBuildEventClustersSeparatesFedMeetingsByMonth(t *testing.T) {
	marKalshi := time.Date(2026, 3, 18, 18, 0, 0, 0, time.UTC)
	marPoly := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	aprKalshi := time.Date(2026, 4, 29, 18, 0, 0, 0, time.UTC)
	aprPoly := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)

	instances := []model.VenueMarketInstance{
		{
			InstanceID:          "kalshi:mar",
			Venue:               model.VenueKalshi,
			EventTitle:          "Fed decision in Mar 2026?",
			EventFamily:         "fed_fomc",
			Category:            "economy",
			DeadlineUTC:         &marKalshi,
			NormalizedYesTarget: "fed no change",
		},
		{
			InstanceID:          "polymarket:mar",
			Venue:               model.VenuePolymarket,
			EventTitle:          "Fed decision in March?",
			EventFamily:         "fed_fomc",
			Category:            "economy",
			DeadlineUTC:         &marPoly,
			NormalizedYesTarget: "fed no change",
		},
		{
			InstanceID:          "kalshi:apr",
			Venue:               model.VenueKalshi,
			EventTitle:          "Fed decision in Apr 2026?",
			EventFamily:         "fed_fomc",
			Category:            "economy",
			DeadlineUTC:         &aprKalshi,
			NormalizedYesTarget: "fed no change",
		},
		{
			InstanceID:          "polymarket:apr",
			Venue:               model.VenuePolymarket,
			EventTitle:          "Fed decision in April?",
			EventFamily:         "fed_fomc",
			Category:            "economy",
			DeadlineUTC:         &aprPoly,
			NormalizedYesTarget: "fed no change",
		},
	}

	events := BuildEventClusters(instances)
	if len(events) != 2 {
		t.Fatalf("expected 2 event clusters, got %d", len(events))
	}
	for _, event := range events {
		if len(event.MarketInstances) != 2 {
			t.Fatalf("expected each event cluster to contain 2 venue instances, got %d", len(event.MarketInstances))
		}
	}
}
