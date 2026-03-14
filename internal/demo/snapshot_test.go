package demo

import (
	"testing"

	"equinox/internal/model"
)

func TestDeriveEvaluationLabelsPrefersExplicitNonMatchAssessment(t *testing.T) {
	labels := DeriveEvaluationLabels(
		[]model.EventCluster{{ClusterID: "event-001"}},
		[]model.PropositionCluster{{ClusterID: "prop-001", Routeability: model.Routeable}},
		[]model.EquivalenceAssessment{{AssessmentID: "assess-007", Classification: "explicit_non_match"}},
	)
	if labels["clear_non_match_case"] != "assess-007" {
		t.Fatalf("expected explicit non-match assessment id, got %q", labels["clear_non_match_case"])
	}
}

func TestLoadFixtureSnapshotIncludesMultipleRouteableClusters(t *testing.T) {
	snapshot, err := LoadFixtureSnapshot()
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}

	routeable := 0
	for _, prop := range snapshot.Props {
		if prop.Routeability == model.Routeable {
			routeable++
		}
	}
	if routeable < 2 {
		t.Fatalf("expected at least 2 routeable proposition clusters, got %d", routeable)
	}
}
