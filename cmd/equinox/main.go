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
		fmt.Println("usage: equinox <serve|fixture-demo|list-clusters|route-order|live-inspect|live-epl>")
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
	case "list-clusters":
		if err := runListClusters(); err != nil {
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
	case "live-epl":
		if err := runLivePremierLeague(); err != nil {
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
	source := fs.String("source", "fixture", "snapshot source: fixture or live-epl")
	matchweeks := fs.Int("matchweeks", 4, "number of current/upcoming EPL matchweek-style windows to fetch when source=live-epl")
	_ = fs.Parse(os.Args[2:])

	snapshot, err := loadSnapshotForSource(*source, *matchweeks)
	if err != nil {
		return err
	}
	snapshot, err = demo.MaterializeSnapshot(context.Background(), "equinox.db", "artifacts", snapshot)
	if err != nil {
		return err
	}

	app, err := web.New(snapshot, *source)
	if err != nil {
		return err
	}

	fmt.Printf("Equinox demo UI available at http://%s\n", *addr)
	fmt.Printf("source: %s\n", *source)
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
	printFixtureSummary(snapshot.Events, snapshot.Props, snapshot.Decisions)
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
		return fmt.Errorf("missing required --cluster flag; run `make list-clusters ROUTEABLE_ONLY=1` first")
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

func runListClusters() error {
	fs := flag.NewFlagSet("list-clusters", flag.ExitOnError)
	routeableOnly := fs.Bool("routeable-only", false, "show only routeable proposition clusters")
	_ = fs.Parse(os.Args[2:])

	snapshot, err := demo.LoadFixtureSnapshot()
	if err != nil {
		return err
	}

	eventTitles := map[string]string{}
	for _, event := range snapshot.Events {
		eventTitles[event.ClusterID] = event.Title
	}

	fmt.Println("proposition clusters:")
	for _, prop := range snapshot.Props {
		if *routeableOnly && prop.Routeability != model.Routeable {
			continue
		}
		fmt.Printf("- %s | event=%s (%s) | routeability=%s | proposition=%s | venues=%s\n",
			prop.ClusterID,
			prop.EventClusterID,
			eventTitles[prop.EventClusterID],
			prop.Routeability,
			prop.Proposition,
			joinVenues(prop.MarketInstances),
		)
		if len(prop.RefusalReasons) > 0 {
			fmt.Printf("  refusal_reasons=%s\n", strings.Join(prop.RefusalReasons, "; "))
		}
		if len(prop.AmbiguityNotes) > 0 {
			fmt.Printf("  ambiguity_notes=%s\n", strings.Join(prop.AmbiguityNotes, "; "))
		}
	}
	return nil
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

func runLivePremierLeague() error {
	fs := flag.NewFlagSet("live-epl", flag.ExitOnError)
	matchweeks := fs.Int("matchweeks", 4, "number of current/upcoming EPL matchweek-style windows to fetch")
	_ = fs.Parse(os.Args[2:])

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	snapshot, err := demo.LoadLivePremierLeagueSnapshot(ctx, *matchweeks)
	if err != nil {
		return err
	}
	snapshot, err = demo.MaterializeSnapshot(context.Background(), "equinox.db", "artifacts", snapshot)
	if err != nil {
		return err
	}

	fmt.Printf("live premier league snapshot complete\nartifact: %s\n\n", snapshot.ArtifactPath)
	printLivePremierLeagueSummary(snapshot, *matchweeks)
	return nil
}

func printFixtureSummary(events []model.EventCluster, props []model.PropositionCluster, decisions []model.RoutingDecision) {
	eventTitles := map[string]string{}
	for _, event := range events {
		eventTitles[event.ClusterID] = event.Title
	}

	fmt.Println("routeable proposition clusters:")
	foundRouteable := false
	for _, p := range props {
		if p.Routeability != model.Routeable {
			continue
		}
		foundRouteable = true
		fmt.Printf("- %s | event=%s | proposition=%s | venues=%s\n", p.ClusterID, eventTitles[p.EventClusterID], p.Proposition, joinVenues(p.MarketInstances))
	}
	if !foundRouteable {
		fmt.Println("- none")
	}

	fmt.Println("\nexample route-order usage:")
	fmt.Println("  make list-clusters ROUTEABLE_ONLY=1")
	for _, p := range props {
		if p.Routeability == model.Routeable {
			if buyLimit, ok := bestBuyLimit(p); ok {
				fmt.Printf("  make route-order CLUSTER=%s SIDE=buy_yes LIMIT=%.2f SIZE=1000\n", p.ClusterID, buyLimit)
			}
			if sellLimit, ok := bestSellLimit(p); ok {
				fmt.Printf("  make route-order CLUSTER=%s SIDE=sell_yes LIMIT=%.2f SIZE=1000\n", p.ClusterID, sellLimit)
			}
		}
	}

	fmt.Println("\ncurrent demo routing outcomes:")
	for _, d := range decisions {
		fmt.Printf("- %s | %s | %s\n", d.Order.PropositionClusterID, d.Action, strings.Join(d.Reasons, "; "))
	}
}

func printLivePremierLeagueSummary(snapshot demo.Snapshot, matchweeks int) {
	eventTitles := map[string]string{}
	for _, event := range snapshot.Events {
		eventTitles[event.ClusterID] = event.Title
	}

	fmt.Printf("event clusters: %d\n", len(snapshot.Events))
	fmt.Printf("proposition clusters: %d\n", len(snapshot.Props))
	fmt.Printf("routeable proposition clusters: %d\n\n", countRouteable(snapshot.Props))

	fmt.Println("routeable proposition clusters:")
	foundRouteable := false
	for _, p := range snapshot.Props {
		if p.Routeability != model.Routeable {
			continue
		}
		foundRouteable = true
		fmt.Printf("- %s | event=%s | proposition=%s | venues=%s\n", p.ClusterID, eventTitles[p.EventClusterID], p.Proposition, joinVenues(p.MarketInstances))
	}
	if !foundRouteable {
		fmt.Println("- none")
	}

	fmt.Println("\nmarketable routing outcomes:")
	if len(snapshot.Decisions) == 0 {
		fmt.Println("- none")
	}
	for _, d := range snapshot.Decisions {
		outcome := string(d.SelectedVenue)
		if d.Action != "route" {
			outcome = "refuse"
		}
		fmt.Printf("- %s | %s @ %.2f | %s | %s\n",
			d.Order.PropositionClusterID,
			d.Order.Side,
			d.Order.LimitProbability,
			outcome,
			strings.Join(d.Reasons, "; "),
		)
	}

	fmt.Println("\nfor a browser demo of live EPL:")
	fmt.Printf("  make dev-live-epl LIVE_MATCHWEEKS=%d\n", matchweeks)
}

func loadSnapshotForSource(source string, matchweeks int) (demo.Snapshot, error) {
	switch source {
	case "fixture":
		return demo.LoadFixtureSnapshot()
	case "live-epl":
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return demo.LoadLivePremierLeagueSnapshot(ctx, matchweeks)
	default:
		return demo.Snapshot{}, fmt.Errorf("unknown source %q", source)
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

func countRouteable(props []model.PropositionCluster) int {
	count := 0
	for _, p := range props {
		if p.Routeability == model.Routeable {
			count++
		}
	}
	return count
}

func bestBuyLimit(prop model.PropositionCluster) (float64, bool) {
	best := 0.0
	found := false
	for _, instance := range prop.MarketInstances {
		if instance.Quote.YesAsk <= 0 {
			continue
		}
		if !found || instance.Quote.YesAsk < best {
			best = instance.Quote.YesAsk
			found = true
		}
	}
	return best, found
}

func bestSellLimit(prop model.PropositionCluster) (float64, bool) {
	best := 0.0
	found := false
	for _, instance := range prop.MarketInstances {
		if instance.Quote.YesBid <= 0 {
			continue
		}
		if !found || instance.Quote.YesBid > best {
			best = instance.Quote.YesBid
			found = true
		}
	}
	return best, found
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
