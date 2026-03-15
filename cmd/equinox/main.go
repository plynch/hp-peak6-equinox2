package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"equinox/internal/adapters/kalshi"
	"equinox/internal/adapters/polymarket"
	"equinox/internal/demo"
	"equinox/internal/model"
	"equinox/internal/web"
)

func main() {
	if len(os.Args) < 2 {
		printMainUsage()
		return
	}
	switch os.Args[1] {
	case "help", "-h", "--help":
		printMainUsage()
		return
	}
	switch os.Args[1] {
	case "serve":
		if err := runServe(); err != nil {
			panic(err)
		}
	case "scan":
		if err := runScan(); err != nil {
			panic(err)
		}
	case "showcase":
		if err := runShowcase(); err != nil {
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
	case "live-fed":
		if err := runLiveFed(); err != nil {
			panic(err)
		}
	default:
		fmt.Printf("unknown command %q\n\n", os.Args[1])
		printMainUsage()
		os.Exit(2)
	}
}

func runServe() error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8080", "server bind address")
	source := fs.String("source", "fixture", "snapshot source: fixture, live-epl, live-fed, or all-live")
	matchweeks := fs.Int("matchweeks", 4, "number of current/upcoming EPL matchweek-style windows to fetch when source=live-epl")
	meetings := fs.Int("meetings", 4, "number of current/upcoming Fed meetings to fetch when source=live-fed")
	_ = fs.Parse(os.Args[2:])

	snapshot, err := loadSnapshotForSource(*source, *matchweeks, *meetings)
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

func runScan() error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	source := fs.String("source", "all-live", "snapshot source: fixture, live-epl, live-fed, or all-live")
	matchweeks := fs.Int("matchweeks", 4, "number of current/upcoming EPL matchweek-style windows to fetch when source=live-epl")
	meetings := fs.Int("meetings", 4, "number of current/upcoming Fed meetings to fetch when source=live-fed")
	_ = fs.Parse(os.Args[2:])

	snapshot, err := loadSnapshotForSource(*source, *matchweeks, *meetings)
	if err != nil {
		return err
	}
	snapshot, err = demo.MaterializeSnapshot(context.Background(), "equinox.db", "artifacts", snapshot)
	if err != nil {
		return err
	}
	return printScanSummary(*source, snapshot, *matchweeks, *meetings)
}

func runShowcase() error {
	fs := flag.NewFlagSet("showcase", flag.ExitOnError)
	matchweeks := fs.Int("matchweeks", 4, "number of current/upcoming EPL matchweek-style windows to fetch")
	meetings := fs.Int("meetings", 4, "number of current/upcoming Fed meetings to fetch")
	_ = fs.Parse(os.Args[2:])

	sources := []string{"fixture", "live-fed", "live-epl"}
	var firstErr error
	successes := 0
	for i, source := range sources {
		snapshot, err := loadSnapshotForSource(source, *matchweeks, *meetings)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			fmt.Printf("\n=== %s ===\nerror: %v\n", sourceHeading(source), err)
			continue
		}
		snapshot, err = demo.MaterializeSnapshot(context.Background(), "equinox.db", "artifacts", snapshot)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			fmt.Printf("\n=== %s ===\nerror: %v\n", sourceHeading(source), err)
			continue
		}
		if i > 0 {
			fmt.Println()
		}
		successes++
		if err := printScanSummary(source, snapshot, *matchweeks, *meetings); err != nil && firstErr == nil {
			firstErr = err
		}
		if i < len(sources)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	if successes == 0 {
		return firstErr
	}
	return nil
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
	printFixtureSummary(snapshot)
	return nil
}

func runRouteOrder() error {
	fs := flag.NewFlagSet("route-order", flag.ExitOnError)
	source := fs.String("source", "fixture", "snapshot source: fixture, live-epl, live-fed, or all-live")
	matchweeks := fs.Int("matchweeks", 4, "number of current/upcoming EPL matchweek-style windows to fetch when source=live-epl")
	meetings := fs.Int("meetings", 4, "number of current/upcoming Fed meetings to fetch when source=live-fed")
	clusterID := fs.String("cluster", "", "proposition cluster id to route against (for example prop-001)")
	eventQuery := fs.String("event-query", "", "case-insensitive event title selector used when --cluster is not provided")
	propQuery := fs.String("prop-query", "", "case-insensitive proposition selector used when --cluster is not provided")
	side := fs.String("side", "buy_yes", "hypothetical order side: buy_yes or sell_yes")
	limit := fs.Float64("limit", 0.60, "limit probability")
	size := fs.Float64("size", 1000, "size notional")
	_ = fs.Parse(os.Args[2:])

	snapshot, err := loadSnapshotForSource(*source, *matchweeks, *meetings)
	if err != nil {
		return err
	}
	selectedClusterID, err := resolveClusterID(snapshot, *source, *clusterID, *eventQuery, *propQuery)
	if err != nil {
		return err
	}

	target, decision, err := demo.SimulateOrder(snapshot, selectedClusterID, *side, *limit, *size)
	if err != nil {
		return err
	}
	out := map[string]any{
		"source":      *source,
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
	source := fs.String("source", "fixture", "snapshot source: fixture, live-epl, live-fed, or all-live")
	matchweeks := fs.Int("matchweeks", 4, "number of current/upcoming EPL matchweek-style windows to fetch when source=live-epl")
	meetings := fs.Int("meetings", 4, "number of current/upcoming Fed meetings to fetch when source=live-fed")
	routeableOnly := fs.Bool("routeable-only", false, "show only routeable proposition clusters")
	_ = fs.Parse(os.Args[2:])

	snapshot, err := loadSnapshotForSource(*source, *matchweeks, *meetings)
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

func runLiveFed() error {
	fs := flag.NewFlagSet("live-fed", flag.ExitOnError)
	meetings := fs.Int("meetings", 4, "number of current/upcoming Fed meetings to fetch")
	_ = fs.Parse(os.Args[2:])

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	snapshot, err := demo.LoadLiveFedSnapshot(ctx, *meetings)
	if err != nil {
		return err
	}
	snapshot, err = demo.MaterializeSnapshot(context.Background(), "equinox.db", "artifacts", snapshot)
	if err != nil {
		return err
	}

	fmt.Printf("live fed snapshot complete\nartifact: %s\n\n", snapshot.ArtifactPath)
	return printScanSummary("live-fed", snapshot, 0, *meetings)
}

func printFixtureSummary(snapshot demo.Snapshot) {
	eventTitles := map[string]string{}
	for _, event := range snapshot.Events {
		eventTitles[event.ClusterID] = event.Title
	}

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

	fmt.Println("\nexample route-order usage:")
	fmt.Println("  make list-clusters ROUTEABLE_ONLY=1")
	for _, p := range snapshot.Props {
		if p.Routeability == model.Routeable {
			if buyLimit, ok := bestBuyLimit(p); ok {
				fmt.Printf("  %s\n", makeRouteCommand(snapshot, "fixture", 0, 0, p, "buy_yes", buyLimit))
			}
			if sellLimit, ok := bestSellLimit(p); ok {
				fmt.Printf("  %s\n", makeRouteCommand(snapshot, "fixture", 0, 0, p, "sell_yes", sellLimit))
			}
		}
	}

	fmt.Println("\ncurrent demo routing outcomes:")
	for _, d := range snapshot.Decisions {
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

func loadSnapshotForSource(source string, matchweeks int, meetings int) (demo.Snapshot, error) {
	switch source {
	case "fixture":
		return demo.LoadFixtureSnapshot()
	case "live-epl":
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return demo.LoadLivePremierLeagueSnapshot(ctx, matchweeks)
	case "live-fed":
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return demo.LoadLiveFedSnapshot(ctx, meetings)
	case "all-live":
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		return demo.LoadAllLiveSnapshot(ctx, matchweeks, meetings)
	default:
		return demo.Snapshot{}, fmt.Errorf("unknown source %q", source)
	}
}

func printScanSummary(source string, snapshot demo.Snapshot, matchweeks int, meetings int) error {
	fmt.Printf("=== %s ===\n", sourceHeading(source))
	fmt.Printf("artifact: %s\n", snapshot.ArtifactPath)
	fmt.Printf("event clusters: %d | proposition clusters: %d | routeable: %d | assessments: %d\n", len(snapshot.Events), len(snapshot.Props), countRouteable(snapshot.Props), len(snapshot.Assessments))
	switch source {
	case "live-epl":
		fmt.Printf("window: current + next %d matchweek-style windows\n", matchweeks)
	case "live-fed":
		fmt.Printf("window: current + next %d Fed meetings\n", meetings)
	case "all-live":
		fmt.Printf("window: current + next %d Fed meetings and current + next %d EPL matchweek-style windows\n", meetings, matchweeks)
	}

	eventTitles := map[string]string{}
	for _, event := range snapshot.Events {
		eventTitles[event.ClusterID] = event.Title
	}

	printRouteableEventsSummary(snapshot)

	fmt.Println("\nrouteable proposition clusters:")
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CLUSTER\tEVENT\tPROPOSITION\tVENUES")
	routeableCount := 0
	for _, prop := range snapshot.Props {
		if prop.Routeability != model.Routeable {
			continue
		}
		routeableCount++
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", prop.ClusterID, eventTitles[prop.EventClusterID], prop.Proposition, joinVenues(prop.MarketInstances))
	}
	_ = tw.Flush()
	if routeableCount == 0 {
		fmt.Println("(none)")
	}

	fmt.Println("\nmarketable routing outcomes:")
	tw = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CLUSTER\tSIDE\tLIMIT\tACTION\tVENUE\tREASONS")
	decisionCount := 0
	for _, d := range snapshot.Decisions {
		if d.Action != "route" {
			continue
		}
		decisionCount++
		fmt.Fprintf(tw, "%s\t%s\t%.2f\t%s\t%s\t%s\n", d.Order.PropositionClusterID, d.Order.Side, d.Order.LimitProbability, d.Action, d.SelectedVenue, strings.Join(d.Reasons, "; "))
	}
	_ = tw.Flush()
	if decisionCount == 0 {
		fmt.Println("(none)")
	}

	fmt.Println("\nnext commands:")
	switch source {
	case "fixture":
		fmt.Println("  make list-clusters ROUTEABLE_ONLY=1")
		printSuggestedRouteCommands(snapshot, source, matchweeks, meetings)
	case "live-epl":
		fmt.Printf("  make list-clusters ROUTEABLE_ONLY=1 SOURCE=live-epl LIVE_MATCHWEEKS=%d\n", matchweeks)
		printSuggestedRouteCommands(snapshot, source, matchweeks, meetings)
		fmt.Printf("  make dev-live-epl LIVE_MATCHWEEKS=%d\n", matchweeks)
	case "live-fed":
		fmt.Printf("  make list-clusters ROUTEABLE_ONLY=1 SOURCE=live-fed FED_MEETINGS=%d\n", meetings)
		printSuggestedRouteCommands(snapshot, source, matchweeks, meetings)
	case "all-live":
		fmt.Printf("  make scan SOURCE=live-fed FED_MEETINGS=%d\n", meetings)
		fmt.Printf("  make scan SOURCE=live-epl LIVE_MATCHWEEKS=%d\n", matchweeks)
		printSuggestedRouteCommands(snapshot, source, matchweeks, meetings)
	}
	return nil
}

func sourceHeading(source string) string {
	switch source {
	case "fixture":
		return "FIXTURE SNAPSHOT"
	case "live-epl":
		return "LIVE PREMIER LEAGUE"
	case "live-fed":
		return "LIVE FED DECISIONS"
	case "all-live":
		return "ALL LIVE ROUTEABLE MARKETS"
	default:
		return strings.ToUpper(source)
	}
}

func printMainUsage() {
	fmt.Println("Equinox CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  equinox <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  serve         Start the local demo web UI")
	fmt.Println("  scan          Scan fixture, live-fed, live-epl, or all-live and print routeable events")
	fmt.Println("  showcase      Run the full terminal showcase across fixture + live-fed + live-epl")
	fmt.Println("  fixture-demo  Materialize the deterministic fixture snapshot")
	fmt.Println("  list-clusters List proposition clusters for a source")
	fmt.Println("  route-order   Simulate a hypothetical order against a routeable proposition cluster")
	fmt.Println("  live-inspect  Check current public API ingestion viability")
	fmt.Println("  live-fed      Run the live Fed scan directly")
	fmt.Println("  live-epl      Run the live EPL scan directly")
	fmt.Println()
	fmt.Println("Best first commands:")
	fmt.Println("  equinox scan")
	fmt.Println("  equinox route-order --source live-epl --event-query 'liverpool vs tottenham' --prop-query 'liverpool win' --limit 0.76 --size 77")
	fmt.Println("  equinox serve")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  equinox scan --source all-live --meetings 2 --matchweeks 1")
	fmt.Println("  equinox scan --source live-fed --meetings 2")
	fmt.Println("  equinox list-clusters --source live-epl --matchweeks 1 --routeable-only")
	fmt.Println("  equinox route-order --source live-epl --event-query 'liverpool vs tottenham' --prop-query 'liverpool win' --limit 0.76 --size 77")
	fmt.Println()
	fmt.Println("Use '<command> -h' for subcommand flags.")
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

type routeableEventRow struct {
	Title            string
	RouteableCount   int
	Propositions     []string
	Venues           string
	EarliestDeadline time.Time
	HasDeadline      bool
}

func printRouteableEventsSummary(snapshot demo.Snapshot) {
	rows := collectRouteableEventRows(snapshot)
	fmt.Println("\nrouteable events:")
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "EVENT\tROUTEABLE_PROPS\tPROPOSITIONS\tVENUES")
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n", row.Title, row.RouteableCount, strings.Join(row.Propositions, ", "), row.Venues)
	}
	_ = tw.Flush()
	if len(rows) == 0 {
		fmt.Println("(none)")
	}
}

func collectRouteableEventRows(snapshot demo.Snapshot) []routeableEventRow {
	eventIndex := map[string]model.EventCluster{}
	for _, event := range snapshot.Events {
		eventIndex[event.ClusterID] = event
	}

	type accumulator struct {
		title          string
		propositions   []string
		venues         map[string]bool
		earliest       time.Time
		hasEarliest    bool
		routeableCount int
	}
	acc := map[string]*accumulator{}

	for _, prop := range snapshot.Props {
		if prop.Routeability != model.Routeable {
			continue
		}
		event := eventIndex[prop.EventClusterID]
		row := acc[prop.EventClusterID]
		if row == nil {
			row = &accumulator{title: event.Title, venues: map[string]bool{}}
			acc[prop.EventClusterID] = row
		}
		row.routeableCount++
		row.propositions = append(row.propositions, prop.Proposition)
		for _, inst := range prop.MarketInstances {
			row.venues[string(inst.Venue)] = true
			if inst.DeadlineUTC != nil && (!row.hasEarliest || inst.DeadlineUTC.Before(row.earliest)) {
				row.earliest = *inst.DeadlineUTC
				row.hasEarliest = true
			}
		}
	}

	rows := make([]routeableEventRow, 0, len(acc))
	for _, row := range acc {
		props := dedupeStrings(row.propositions)
		sort.Strings(props)
		venues := make([]string, 0, len(row.venues))
		for venue := range row.venues {
			venues = append(venues, venue)
		}
		sort.Strings(venues)
		rows = append(rows, routeableEventRow{
			Title:            row.title,
			RouteableCount:   row.routeableCount,
			Propositions:     props,
			Venues:           strings.Join(venues, ","),
			EarliestDeadline: row.earliest,
			HasDeadline:      row.hasEarliest,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].HasDeadline != rows[j].HasDeadline {
			return rows[i].HasDeadline
		}
		if rows[i].HasDeadline && !rows[i].EarliestDeadline.Equal(rows[j].EarliestDeadline) {
			return rows[i].EarliestDeadline.Before(rows[j].EarliestDeadline)
		}
		return rows[i].Title < rows[j].Title
	})
	return rows
}

func dedupeStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, item := range in {
		if seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func resolveClusterID(snapshot demo.Snapshot, source, clusterID, eventQuery, propQuery string) (string, error) {
	if clusterID != "" {
		return clusterID, nil
	}
	if eventQuery == "" && propQuery == "" {
		return "", fmt.Errorf("missing required selector; provide --cluster or a combination of --event-query/--prop-query. Run `equinox list-clusters --source %s --routeable-only` first", source)
	}

	eventTitles := map[string]string{}
	for _, event := range snapshot.Events {
		eventTitles[event.ClusterID] = event.Title
	}

	matches := make([]model.PropositionCluster, 0)
	for _, prop := range snapshot.Props {
		if prop.Routeability != model.Routeable {
			continue
		}
		if eventQuery != "" && !containsFold(eventTitles[prop.EventClusterID], eventQuery) {
			continue
		}
		if propQuery != "" && !containsFold(prop.Proposition, propQuery) {
			continue
		}
		matches = append(matches, prop)
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no routeable proposition cluster matched event-query=%q prop-query=%q. Run `equinox list-clusters --source %s --routeable-only` first", eventQuery, propQuery, source)
	case 1:
		return matches[0].ClusterID, nil
	default:
		lines := make([]string, 0, len(matches))
		for _, match := range matches {
			lines = append(lines, fmt.Sprintf("%s (%s | %s)", match.ClusterID, eventTitles[match.EventClusterID], match.Proposition))
		}
		return "", fmt.Errorf("selector was ambiguous; matched %d routeable proposition clusters: %s", len(matches), strings.Join(lines, "; "))
	}
}

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(strings.TrimSpace(needle)))
}

func printSuggestedRouteCommands(snapshot demo.Snapshot, source string, matchweeks int, meetings int) {
	for _, prop := range snapshot.Props {
		if prop.Routeability != model.Routeable {
			continue
		}
		if buyLimit, ok := bestBuyLimit(prop); ok {
			fmt.Printf("  %s\n", makeRouteCommand(snapshot, source, matchweeks, meetings, prop, "buy_yes", buyLimit))
		}
		break
	}
}

func makeRouteCommand(snapshot demo.Snapshot, source string, matchweeks int, meetings int, prop model.PropositionCluster, side string, limit float64) string {
	eventTitle := ""
	for _, event := range snapshot.Events {
		if event.ClusterID == prop.EventClusterID {
			eventTitle = event.Title
			break
		}
	}
	args := []string{"make route-order"}
	if source != "fixture" {
		args = append(args, fmt.Sprintf("SOURCE=%s", source))
	}
	if source == "live-epl" {
		args = append(args, fmt.Sprintf("LIVE_MATCHWEEKS=%d", matchweeks))
	}
	if source == "live-fed" {
		args = append(args, fmt.Sprintf("FED_MEETINGS=%d", meetings))
	}
	args = append(args,
		fmt.Sprintf("EVENT_QUERY='%s'", eventTitle),
		fmt.Sprintf("PROP_QUERY='%s'", prop.Proposition),
		fmt.Sprintf("SIDE=%s", side),
		fmt.Sprintf("LIMIT=%.2f", limit),
		"SIZE=1000",
	)
	return strings.Join(args, " ")
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
