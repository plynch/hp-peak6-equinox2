package router

import (
	"testing"

	"equinox/internal/model"
)

func TestSimulateRouteable(t *testing.T) {
	p := model.PropositionCluster{
		ClusterID:    "prop-1",
		Routeability: model.Routeable,
		MarketInstances: []model.VenueMarketInstance{
			{InstanceID: "a", Venue: model.VenuePolymarket, Quote: model.QuoteView{YesAsk: 0.62, DepthNotional: 2000, Observed: true}},
			{InstanceID: "b", Venue: model.VenueKalshi, Quote: model.QuoteView{YesAsk: 0.60, DepthNotional: 1800, Observed: true}},
		},
	}
	d := Simulate(model.HypotheticalOrder{OrderID: "1", PropositionClusterID: "prop-1", Side: "buy_yes", LimitProbability: 0.60}, []model.PropositionCluster{p})
	if d.Action != "route" {
		t.Fatalf("expected route, got %s", d.Action)
	}
	if d.SelectedInstanceID != "b" {
		t.Fatalf("expected b to satisfy limit and win, got %s", d.SelectedInstanceID)
	}
}

func TestSimulateRejectsWhenAllViolateLimit(t *testing.T) {
	p := model.PropositionCluster{
		ClusterID:    "prop-1",
		Routeability: model.Routeable,
		MarketInstances: []model.VenueMarketInstance{
			{InstanceID: "a", Venue: model.VenuePolymarket, Quote: model.QuoteView{YesAsk: 0.62, DepthNotional: 2000, Observed: true}},
			{InstanceID: "b", Venue: model.VenueKalshi, Quote: model.QuoteView{YesAsk: 0.64, DepthNotional: 1800, Observed: true}},
		},
	}
	d := Simulate(model.HypotheticalOrder{OrderID: "2", PropositionClusterID: "prop-1", Side: "buy_yes", LimitProbability: 0.60}, []model.PropositionCluster{p})
	if d.Action != "refuse" {
		t.Fatalf("expected refusal, got %s", d.Action)
	}
	if len(d.RankedCandidates) == 0 {
		t.Fatalf("expected ranked candidates with rejection reasons")
	}
}
