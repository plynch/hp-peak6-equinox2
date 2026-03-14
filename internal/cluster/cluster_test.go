package cluster

import (
	"testing"

	"equinox/internal/adapters/kalshi"
	"equinox/internal/adapters/polymarket"
	"equinox/internal/model"
	"equinox/internal/normalize"
)

func TestFixtureYieldsRouteableAndUnsupportedAndAmbiguous(t *testing.T) {
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
	props, _ := BuildPropositionClusters(events)
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
}
