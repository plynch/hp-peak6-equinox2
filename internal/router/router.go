package router

import (
	"fmt"
	"sort"

	"equinox/internal/model"
)

func Simulate(order model.HypotheticalOrder, props []model.PropositionCluster) model.RoutingDecision {
	for _, p := range props {
		if p.ClusterID != order.PropositionClusterID {
			continue
		}
		if p.Routeability != model.Routeable {
			return model.RoutingDecision{DecisionID: "route-" + order.OrderID, Order: order, Action: "refuse", Reasons: append([]string{"proposition cluster not routeable"}, p.RefusalReasons...)}
		}
		type cand struct {
			id    string
			venue model.Venue
			score float64
		}
		var cands []cand
		for _, m := range p.MarketInstances {
			price := m.Quote.YesAsk
			if order.Side == "sell_yes" {
				price = m.Quote.YesBid
			}
			score := (1.0 - abs(price-order.LimitProbability)) + (m.Quote.DepthNotional / 10000.0)
			if !m.Quote.Observed {
				score -= 0.2
			}
			cands = append(cands, cand{id: m.InstanceID, venue: m.Venue, score: score})
		}
		sort.Slice(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
		ranked := make([]string, len(cands))
		for i, c := range cands {
			ranked[i] = fmt.Sprintf("%s(score=%.4f)", c.id, c.score)
		}
		if len(cands) == 0 {
			return model.RoutingDecision{DecisionID: "route-" + order.OrderID, Order: order, Action: "refuse", Reasons: []string{"no candidates"}}
		}
		return model.RoutingDecision{
			DecisionID:         "route-" + order.OrderID,
			Order:              order,
			SelectedInstanceID: cands[0].id,
			SelectedVenue:      cands[0].venue,
			Action:             "route",
			Reasons:            []string{"best normalized score by price proximity and depth", "router consumed normalized proposition cluster inputs only"},
			RankedCandidates:   ranked,
		}
	}
	return model.RoutingDecision{DecisionID: "route-" + order.OrderID, Order: order, Action: "refuse", Reasons: []string{"proposition cluster not found"}}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
