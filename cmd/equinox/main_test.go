package main

import (
	"testing"

	"equinox/internal/demo"
	"equinox/internal/model"
)

func TestResolveClusterIDByHumanReadableSelectors(t *testing.T) {
	snapshot := demo.Snapshot{
		Events: []model.EventCluster{
			{ClusterID: "event-001", Title: "fed decision in march?"},
		},
		Props: []model.PropositionCluster{
			{
				ClusterID:      "prop-001",
				EventClusterID: "event-001",
				Proposition:    "fed no change",
				Routeability:   model.Routeable,
			},
			{
				ClusterID:      "prop-002",
				EventClusterID: "event-001",
				Proposition:    "fed cut exactly 25bps",
				Routeability:   model.Routeable,
			},
		},
	}

	clusterID, err := resolveClusterID(snapshot, "live-fed", "", "march", "no change")
	if err != nil {
		t.Fatalf("expected selector resolution to succeed, got error: %v", err)
	}
	if clusterID != "prop-001" {
		t.Fatalf("expected prop-001, got %s", clusterID)
	}
}

func TestResolveClusterIDRejectsAmbiguousSelectors(t *testing.T) {
	snapshot := demo.Snapshot{
		Events: []model.EventCluster{
			{ClusterID: "event-001", Title: "liverpool vs tottenham"},
		},
		Props: []model.PropositionCluster{
			{ClusterID: "prop-001", EventClusterID: "event-001", Proposition: "liverpool win", Routeability: model.Routeable},
			{ClusterID: "prop-002", EventClusterID: "event-001", Proposition: "tottenham win", Routeability: model.Routeable},
		},
	}

	if _, err := resolveClusterID(snapshot, "live-epl", "", "liverpool vs tottenham", "win"); err == nil {
		t.Fatalf("expected ambiguous selector to fail")
	}
}
