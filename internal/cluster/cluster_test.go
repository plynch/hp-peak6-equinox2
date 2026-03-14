package cluster

import (
	"testing"

	"equinox/internal/adapters/kalshi"
	"equinox/internal/adapters/polymarket"
	"equinox/internal/model"
	"equinox/internal/normalize"
)

func TestFixtureClassificationsAndExplicitNonMatch(t *testing.T) {
	pmRows, err := (polymarket.Adapter{}).LoadFixture("../../testdata/fixtures/polymarket_markets.json")
	if err != nil {
		t.Fatal(err)
	}
	kRows, err := (kalshi.Adapter{}).LoadFixture("../../testdata/fixtures/kalshi_markets.json")
	if err != nil {
		t.Fatal(err)
	}
	instances := append(normalize.FromPolymarket(pmRows), normalize.FromKalshi(kRows)...)
	events := BuildEventClusters(instances)
	props, assessments := BuildPropositionClusters(events)
	count := map[model.Routeability]int{}
	for _, p := range props {
		count[p.Routeability]++
	}
	if count[model.Routeable] < 1 {
		t.Fatalf("expected at least one routeable proposition")
	}
	if count[model.Unsupported] < 1 {
		t.Fatalf("expected unsupported proposition")
	}
	if count[model.EventOnly] < 1 {
		t.Fatalf("expected event-only proposition")
	}
	if count[model.Ambiguous] < 1 {
		t.Fatalf("expected ambiguous proposition")
	}

	hasExplicit := false
	for _, a := range assessments {
		if a.Classification == "explicit_non_match" {
			hasExplicit = true
			break
		}
	}
	if !hasExplicit {
		t.Fatalf("expected explicit_non_match assessment for paired non-match case")
	}
}
