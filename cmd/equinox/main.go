package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"equinox/internal/adapters/kalshi"
	"equinox/internal/adapters/polymarket"
	"equinox/internal/artifacts"
	"equinox/internal/cluster"
	"equinox/internal/model"
	"equinox/internal/normalize"
	"equinox/internal/router"
	"equinox/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: equinox <fixture-demo|route-order|live-inspect>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "fixture-demo":
		if err := runFixtureDemo(); err != nil {
			panic(err)
		}
	case "route-order":
		if err := runRouteOrder(); err != nil {
			panic(err)
		}
	case "live-inspect":
		if err := runLiveInspect(); err != nil {
			panic(err)
		}
	default:
		fmt.Println("unknown command")
		os.Exit(1)
	}
}

func loadFixtureState() ([]model.VenueMarketInstance, []model.EventCluster, []model.PropositionCluster, []model.EquivalenceAssessment, error) {
	pmRows, err := (polymarket.Adapter{}).LoadFixture("testdata/fixtures/polymarket_markets.json")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	kRows, err := (kalshi.Adapter{}).LoadFixture("testdata/fixtures/kalshi_markets.json")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	instances := append(normalize.FromPolymarket(pmRows), normalize.FromKalshi(kRows)...)
	events := cluster.BuildEventClusters(instances)
	props, assessments := cluster.BuildPropositionClusters(events)
	return instances, events, props, assessments, nil
}

func runFixtureDemo() error {
	instances, events, props, assessments, err := loadFixtureState()
	if err != nil {
		return err
	}

	orders := []model.HypotheticalOrder{}
	for _, p := range props {
		orders = append(orders, model.HypotheticalOrder{OrderID: p.ClusterID, PropositionClusterID: p.ClusterID, Side: "buy_yes", LimitProbability: 0.60, SizeNotional: 1000})
	}
	decisions := make([]model.RoutingDecision, 0, len(orders))
	for _, o := range orders {
		decisions = append(decisions, router.Simulate(o, props))
	}

	st, err := store.Open("equinox.db")
	if err != nil {
		return err
	}
	defer st.Close()
	if err := st.PersistRun(context.Background(), events, props, assessments, decisions); err != nil {
		return err
	}

	runDir := filepath.Join("artifacts", time.Now().UTC().Format("20060102-150405"))
	eval := deriveEvaluationLabels(events, props, assessments)
	if err := artifacts.Write(runDir, artifacts.Bundle{Instances: instances, Events: events, Props: props, Assessments: assessments, Decisions: decisions, Evaluation: eval}); err != nil {
		return err
	}
	fmt.Printf("fixture demo complete\nartifact: %s/bundle.json\n\n", runDir)
	printFixtureSummary(props, decisions)
	return nil
}

func runRouteOrder() error {
	fs := flag.NewFlagSet("route-order", flag.ExitOnError)
	clusterID := fs.String("cluster", "", "proposition cluster id to route against (for example prop-001)")
	side := fs.String("side", "buy_yes", "hypothetical order side: buy_yes or sell_yes")
	limit := fs.Float64("limit", 0.60, "limit probability")
	size := fs.Float64("size", 1000, "size notional")
	_ = fs.Parse(os.Args[2:])

	if *clusterID == "" {
		return fmt.Errorf("missing required --cluster flag")
	}

	_, _, props, _, err := loadFixtureState()
	if err != nil {
		return err
	}

	var target *model.PropositionCluster
	for i := range props {
		if props[i].ClusterID == *clusterID {
			target = &props[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("unknown proposition cluster %q", *clusterID)
	}

	order := model.HypotheticalOrder{
		OrderID:              "manual-" + *clusterID,
		PropositionClusterID: *clusterID,
		Side:                 *side,
		LimitProbability:     *limit,
		SizeNotional:         *size,
	}
	decision := router.Simulate(order, props)
	out := map[string]any{
		"order":      order,
		"cluster":    target,
		"decision":   decision,
		"how_to_read": "routeable clusters can accept buy_yes or sell_yes hypothetical orders; the router rejects clusters that are unsupported, ambiguous, event-only, or outside the order limit",
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func deriveEvaluationLabels(events []model.EventCluster, props []model.PropositionCluster, assessments []model.EquivalenceAssessment) map[string]string {
	labels := map[string]string{}
	for _, p := range props {
		switch p.Routeability {
		case model.Routeable:
			if labels["strong_route_safe_proposition_cluster"] == "" {
				labels["strong_route_safe_proposition_cluster"] = p.ClusterID
			}
		case model.EventOnly:
			if labels["near_match_or_event_only_case"] == "" {
				labels["near_match_or_event_only_case"] = p.ClusterID
			}
		case model.Unsupported:
			if labels["unsupported_shape_case"] == "" {
				labels["unsupported_shape_case"] = p.ClusterID
			}
		case model.Ambiguous:
			if labels["ambiguity_case"] == "" {
				labels["ambiguity_case"] = p.ClusterID
			}
		}
	}
	for _, a := range assessments {
		if a.Classification == "explicit_non_match" {
			labels["clear_non_match_case"] = a.AssessmentID
			break
		}
	}
	if labels["clear_non_match_case"] == "" {
		for _, e := range events {
			if len(e.MarketInstances) == 1 {
				labels["clear_non_match_case"] = e.ClusterID
				break
			}
		}
	}
	return labels
}

func runLiveInspect() error {
	fs := flag.NewFlagSet("live-inspect", flag.ExitOnError)
	limit := fs.Int("limit", 3, "items per venue")
	_ = fs.Parse(os.Args[2:])
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pm, pmErr := (polymarket.Adapter{}).LiveInspect(ctx, *limit)
	ka, kaErr := (kalshi.Adapter{}).LiveInspect(ctx, *limit)
	out := map[string]any{"polymarket_count": len(pm), "kalshi_count": len(ka), "polymarket_error": errText(pmErr), "kalshi_error": errText(kaErr), "timestamp": time.Now().UTC()}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printFixtureSummary(props []model.PropositionCluster, decisions []model.RoutingDecision) {
	fmt.Println("routeable proposition clusters:")
	foundRouteable := false
	for _, p := range props {
		if p.Routeability != model.Routeable {
			continue
		}
		foundRouteable = true
		fmt.Printf("- %s | %s | venues=%s\n", p.ClusterID, p.Proposition, joinVenues(p.MarketInstances))
	}
	if !foundRouteable {
		fmt.Println("- none")
	}

	fmt.Println("\nexample route-order usage:")
	for _, p := range props {
		if p.Routeability == model.Routeable {
			fmt.Printf("  make route-order CLUSTER=%s SIDE=buy_yes LIMIT=0.60 SIZE=1000\n", p.ClusterID)
			fmt.Printf("  make route-order CLUSTER=%s SIDE=sell_yes LIMIT=0.55 SIZE=1000\n", p.ClusterID)
			break
		}
	}

	fmt.Println("\ncurrent demo routing outcomes:")
	for _, d := range decisions {
		fmt.Printf("- %s | %s | %s\n", d.Order.PropositionClusterID, d.Action, strings.Join(d.Reasons, "; "))
	}
}

func joinVenues(instances []model.VenueMarketInstance) string {
	seen := map[model.Venue]bool{}
	out := make([]string, 0, len(instances))
	for _, inst := range instances {
		if !seen[inst.Venue] {
			seen[inst.Venue] = true
			out = append(out, string(inst.Venue))
		}
	}
	return strings.Join(out, ",")
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
