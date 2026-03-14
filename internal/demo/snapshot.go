package demo

import (
	"context"
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

type Snapshot struct {
	Instances    []model.VenueMarketInstance
	Events       []model.EventCluster
	Props        []model.PropositionCluster
	Assessments  []model.EquivalenceAssessment
	Decisions    []model.RoutingDecision
	Evaluation   map[string]string
	GeneratedAt  time.Time
	ArtifactPath string
}

func LoadFixtureSnapshot() (Snapshot, error) {
	pmFixture, err := fixturePath("polymarket_markets.json")
	if err != nil {
		return Snapshot{}, err
	}
	kalshiFixture, err := fixturePath("kalshi_markets.json")
	if err != nil {
		return Snapshot{}, err
	}

	pmRows, err := (polymarket.Adapter{}).LoadFixture(pmFixture)
	if err != nil {
		return Snapshot{}, err
	}
	kRows, err := (kalshi.Adapter{}).LoadFixture(kalshiFixture)
	if err != nil {
		return Snapshot{}, err
	}

	instances := append(normalize.FromPolymarket(pmRows), normalize.FromKalshi(kRows)...)
	events := cluster.BuildEventClusters(instances)
	props, assessments := cluster.BuildPropositionClusters(events)
	decisions := defaultDecisions(props)

	return Snapshot{
		Instances:   instances,
		Events:      events,
		Props:       props,
		Assessments: assessments,
		Decisions:   decisions,
		Evaluation:  DeriveEvaluationLabels(events, props, assessments),
		GeneratedAt: time.Now().UTC(),
	}, nil
}

func fixturePath(name string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := wd; ; {
		candidate := filepath.Join(dir, "testdata", "fixtures", name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	return "", fmt.Errorf("fixture %q not found from %s", name, wd)
}

func MaterializeSnapshot(ctx context.Context, dbPath string, artifactRoot string, snapshot Snapshot) (Snapshot, error) {
	st, err := store.Open(dbPath)
	if err != nil {
		return Snapshot{}, err
	}
	defer st.Close()

	if err := st.PersistRun(ctx, snapshot.Events, snapshot.Props, snapshot.Assessments, snapshot.Decisions); err != nil {
		return Snapshot{}, err
	}

	runDir := filepath.Join(artifactRoot, time.Now().UTC().Format("20060102-150405"))
	if err := artifacts.Write(runDir, artifacts.Bundle{
		Instances:   snapshot.Instances,
		Events:      snapshot.Events,
		Props:       snapshot.Props,
		Assessments: snapshot.Assessments,
		Decisions:   snapshot.Decisions,
		Evaluation:  snapshot.Evaluation,
	}); err != nil {
		return Snapshot{}, err
	}

	snapshot.ArtifactPath = filepath.Join(runDir, "bundle.json")
	return snapshot, nil
}

func SimulateOrder(snapshot Snapshot, clusterID, side string, limit, size float64) (*model.PropositionCluster, model.RoutingDecision, error) {
	var target *model.PropositionCluster
	for i := range snapshot.Props {
		if snapshot.Props[i].ClusterID == clusterID {
			target = &snapshot.Props[i]
			break
		}
	}
	if target == nil {
		return nil, model.RoutingDecision{}, fmt.Errorf("unknown proposition cluster %q", clusterID)
	}

	order := model.HypotheticalOrder{
		OrderID:              "manual-" + clusterID,
		PropositionClusterID: clusterID,
		Side:                 side,
		LimitProbability:     limit,
		SizeNotional:         size,
	}
	return target, router.Simulate(order, snapshot.Props), nil
}

func DeriveEvaluationLabels(events []model.EventCluster, props []model.PropositionCluster, assessments []model.EquivalenceAssessment) map[string]string {
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

func defaultDecisions(props []model.PropositionCluster) []model.RoutingDecision {
	orders := make([]model.HypotheticalOrder, 0, len(props))
	for _, p := range props {
		orders = append(orders, model.HypotheticalOrder{
			OrderID:              p.ClusterID,
			PropositionClusterID: p.ClusterID,
			Side:                 "buy_yes",
			LimitProbability:     0.60,
			SizeNotional:         1000,
		})
	}

	decisions := make([]model.RoutingDecision, 0, len(orders))
	for _, o := range orders {
		decisions = append(decisions, router.Simulate(o, props))
	}
	return decisions
}
