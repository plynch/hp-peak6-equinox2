package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"equinox/internal/adapters/kalshi"
	"equinox/internal/adapters/polymarket"
	"equinox/internal/demo"
	"equinox/internal/model"
	"equinox/internal/web"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: equinox <serve|fixture-demo|route-order|live-inspect>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "serve":
		if err := runServe(); err != nil {
			panic(err)
		}
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

func runServe() error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8080", "server bind address")
	_ = fs.Parse(os.Args[2:])

	snapshot, err := demo.LoadFixtureSnapshot()
	if err != nil {
		return err
	}
	snapshot, err = demo.MaterializeSnapshot(context.Background(), "equinox.db", "artifacts", snapshot)
	if err != nil {
		return err
	}

	app, err := web.New(snapshot)
	if err != nil {
		return err
	}

	fmt.Printf("Equinox demo UI available at http://%s\n", *addr)
	fmt.Printf("artifact: %s\n", snapshot.ArtifactPath)
	fmt.Println("press Ctrl+C to stop")

	server := &http.Server{
		Addr:              *addr,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server.ListenAndServe()
}

func runFixtureDemo() error {
	snapshot, err := demo.LoadFixtureSnapshot()
	if err != nil {
		return err
	}
	snapshot, err = demo.MaterializeSnapshot(context.Background(), "equinox.db", "artifacts", snapshot)
	if err != nil {
		return err
	}
	fmt.Printf("fixture demo complete\nartifact: %s\n\n", snapshot.ArtifactPath)
	printFixtureSummary(snapshot.Props, snapshot.Decisions)
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

	snapshot, err := demo.LoadFixtureSnapshot()
	if err != nil {
		return err
	}
	target, decision, err := demo.SimulateOrder(snapshot, *clusterID, *side, *limit, *size)
	if err != nil {
		return err
	}
	out := map[string]any{
		"order":       decision.Order,
		"cluster":     target,
		"decision":    decision,
		"how_to_read": "routeable clusters can accept buy_yes or sell_yes hypothetical orders; the router rejects clusters that are unsupported, ambiguous, event-only, or outside the order limit",
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
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
