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
	if d.SelectedInstanceID == "" {
		t.Fatal("expected selected instance")
	}
}
