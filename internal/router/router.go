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
			id          string
			venue       model.Venue
			score       float64
			execPrice   float64
			execAllowed bool
			reason      string
		}
		var cands []cand
		for _, m := range p.MarketInstances {
			execPrice, ok := executablePrice(order.Side, m.Quote)
			if !ok {
				cands = append(cands, cand{id: m.InstanceID, venue: m.Venue, execAllowed: false, reason: "missing executable quote side"})
				continue
			}
			if !withinLimit(order.Side, execPrice, order.LimitProbability) {
				cands = append(cands, cand{id: m.InstanceID, venue: m.Venue, execAllowed: false, execPrice: execPrice, reason: fmt.Sprintf("violates limit %.4f", order.LimitProbability)})
				continue
			}
			score := (1.0 - abs(execPrice-order.LimitProbability)) + (m.Quote.DepthNotional / 10000.0)
			if !m.Quote.Observed {
				score -= 0.2
			}
			cands = append(cands, cand{id: m.InstanceID, venue: m.Venue, score: score, execPrice: execPrice, execAllowed: true})
		}

		sort.Slice(cands, func(i, j int) bool {
			if cands[i].execAllowed != cands[j].execAllowed {
				return cands[i].execAllowed
			}
			return cands[i].score > cands[j].score
		})
		ranked := make([]string, len(cands))
		allowedCount := 0
		for i, c := range cands {
			if c.execAllowed {
				allowedCount++
				ranked[i] = fmt.Sprintf("%s(exec=%.4f score=%.4f)", c.id, c.execPrice, c.score)
			} else {
				ranked[i] = fmt.Sprintf("%s(rejected:%s)", c.id, c.reason)
			}
		}
		if allowedCount == 0 {
			return model.RoutingDecision{
				DecisionID:       "route-" + order.OrderID,
				Order:            order,
				Action:           "refuse",
				Reasons:          []string{"no venue satisfies limit-price feasibility"},
				RankedCandidates: ranked,
			}
		}
		winner := cands[0]
		return model.RoutingDecision{
			DecisionID:         "route-" + order.OrderID,
			Order:              order,
			SelectedInstanceID: winner.id,
			SelectedVenue:      winner.venue,
			Action:             "route",
			Reasons:            []string{"best feasible normalized score by executable price proximity and depth", "router consumed normalized proposition cluster inputs only"},
			RankedCandidates:   ranked,
		}
	}
	return model.RoutingDecision{DecisionID: "route-" + order.OrderID, Order: order, Action: "refuse", Reasons: []string{"proposition cluster not found"}}
}

func executablePrice(side string, q model.QuoteView) (float64, bool) {
	switch side {
	case "buy_yes":
		if q.YesAsk <= 0 {
			return 0, false
		}
		return q.YesAsk, true
	case "sell_yes":
		if q.YesBid <= 0 {
			return 0, false
		}
		return q.YesBid, true
	default:
		return 0, false
	}
}

func withinLimit(side string, executionPrice, limit float64) bool {
	switch side {
	case "buy_yes":
		return executionPrice <= limit
	case "sell_yes":
		return executionPrice >= limit
	default:
		return false
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
