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
