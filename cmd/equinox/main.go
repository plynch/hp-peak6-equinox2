package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
		fmt.Println("usage: equinox <fixture-demo|live-inspect>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "fixture-demo":
		if err := runFixtureDemo(); err != nil {
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

func runFixtureDemo() error {
	pmRows, err := (polymarket.Adapter{}).LoadFixture("testdata/fixtures/polymarket_markets.json")
	if err != nil {
		return err
	}
	kRows, err := (kalshi.Adapter{}).LoadFixture("testdata/fixtures/kalshi_markets.json")
	if err != nil {
		return err
	}
	instances := append(normalize.FromPolymarket(pmRows), normalize.FromKalshi(kRows)...)
	events := cluster.BuildEventClusters(instances)
	props, assessments := cluster.BuildPropositionClusters(events)

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
	eval := deriveEvaluationLabels(events, props)
	if err := artifacts.Write(runDir, artifacts.Bundle{Instances: instances, Events: events, Props: props, Assessments: assessments, Decisions: decisions, Evaluation: eval}); err != nil {
		return err
	}
	fmt.Printf("fixture demo complete\nartifact: %s/bundle.json\n", runDir)
	return nil
}

func deriveEvaluationLabels(events []model.EventCluster, props []model.PropositionCluster) map[string]string {
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
	for _, e := range events {
		if len(e.MarketInstances) == 1 {
			labels["clear_non_match_case"] = e.ClusterID
			break
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

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
