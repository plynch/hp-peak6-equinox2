package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"equinox/internal/demo"
	"equinox/internal/model"
)

//go:embed templates/index.html.tmpl
var templateFS embed.FS

type App struct {
	snapshot demo.Snapshot
	source   string
	tmpl     *template.Template
}

type routeResult struct {
	Cluster  *model.PropositionCluster
	Decision model.RoutingDecision
}

type viewData struct {
	Snapshot          demo.Snapshot
	SourceLabel       string
	SourceDescription string
	RouteableCount    int
	RouteableClusters []model.PropositionCluster
	EvaluationRows    []evaluationRow
	RouteResult       *routeResult
	RouteError        string
	DefaultClusterID  string
	DefaultSide       string
	DefaultLimit      string
	DefaultSize       string
}

type evaluationRow struct {
	Label string
	ID    string
}

func New(snapshot demo.Snapshot, source string) (*App, error) {
	funcs := template.FuncMap{
		"join":        strings.Join,
		"joinVenues":  joinVenues,
		"formatFloat": func(v float64) string { return fmt.Sprintf("%.2f", v) },
	}
	tmpl, err := template.New("index.html.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/index.html.tmpl")
	if err != nil {
		return nil, err
	}
	return &App{snapshot: snapshot, source: source, tmpl: tmpl}, nil
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	defaultCluster, defaultLimit := defaultRoute(a.snapshot.Props)
	data := viewData{
		Snapshot:          a.snapshot,
		SourceLabel:       sourceLabel(a.source),
		SourceDescription: sourceDescription(a.source),
		RouteableCount:    countRouteable(a.snapshot.Props),
		RouteableClusters: sortedRouteableProps(a.snapshot.Props),
		EvaluationRows:    sortedEvaluationRows(a.snapshot.Evaluation),
		DefaultClusterID:  defaultCluster,
		DefaultSide:       "buy_yes",
		DefaultLimit:      fmt.Sprintf("%.2f", defaultLimit),
		DefaultSize:       "1000",
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad route form", http.StatusBadRequest)
			return
		}
		clusterID := r.FormValue("cluster")
		side := r.FormValue("side")
		limit := r.FormValue("limit")
		size := r.FormValue("size")

		data.DefaultClusterID = clusterID
		data.DefaultSide = side
		data.DefaultLimit = limit
		data.DefaultSize = size

		limitValue, err := strconv.ParseFloat(limit, 64)
		if err != nil {
			data.RouteError = "limit must be a valid probability"
			a.render(w, data)
			return
		}
		sizeValue, err := strconv.ParseFloat(size, 64)
		if err != nil {
			data.RouteError = "size must be a valid notional amount"
			a.render(w, data)
			return
		}

		target, decision, err := demo.SimulateOrder(a.snapshot, clusterID, side, limitValue, sizeValue)
		if err != nil {
			data.RouteError = err.Error()
			a.render(w, data)
			return
		}
		data.RouteResult = &routeResult{Cluster: target, Decision: decision}
	}

	a.render(w, data)
}

func (a *App) render(w http.ResponseWriter, data viewData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func defaultRoute(props []model.PropositionCluster) (string, float64) {
	routeable := sortedRouteableProps(props)
	if len(routeable) == 0 {
		if len(props) == 0 {
			return "", 0.60
		}
		return props[0].ClusterID, 0.60
	}
	sort.Slice(routeable, func(i, j int) bool {
		di, okI := earliestDeadline(routeable[i])
		dj, okJ := earliestDeadline(routeable[j])
		if okI != okJ {
			return okI
		}
		if okI && !di.Equal(dj) {
			return di.Before(dj)
		}
		winI := strings.Contains(routeable[i].Proposition, "win")
		winJ := strings.Contains(routeable[j].Proposition, "win")
		if winI != winJ {
			return winI
		}
		return routeable[i].Proposition < routeable[j].Proposition
	})
	limit, ok := bestBuyLimit(routeable[0])
	if !ok {
		limit = 0.60
	}
	return routeable[0].ClusterID, limit
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
	sort.Strings(out)
	return strings.Join(out, ", ")
}

func sortedRouteableProps(props []model.PropositionCluster) []model.PropositionCluster {
	out := append([]model.PropositionCluster(nil), props...)
	filtered := out[:0]
	for _, p := range out {
		if p.Routeability == model.Routeable {
			filtered = append(filtered, p)
		}
	}
	out = filtered
	sort.Slice(out, func(i, j int) bool {
		return out[i].ClusterID < out[j].ClusterID
	})
	return out
}

func sortedEvaluationRows(labels map[string]string) []evaluationRow {
	rows := make([]evaluationRow, 0, len(labels))
	for label, id := range labels {
		rows = append(rows, evaluationRow{Label: label, ID: id})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Label < rows[j].Label
	})
	return rows
}

func earliestDeadline(prop model.PropositionCluster) (time.Time, bool) {
	var best time.Time
	found := false
	for _, instance := range prop.MarketInstances {
		if instance.DeadlineUTC == nil {
			continue
		}
		if !found || instance.DeadlineUTC.Before(best) {
			best = *instance.DeadlineUTC
			found = true
		}
	}
	return best, found
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

func sourceLabel(source string) string {
	switch source {
	case "live-epl":
		return "Live Premier League"
	default:
		return "Fixture snapshot"
	}
}

func sourceDescription(source string) string {
	switch source {
	case "live-epl":
		return "This local UI is backed by a live Premier League snapshot fetched from the current public Polymarket and Kalshi APIs."
	default:
		return "This local UI is fixture-backed for deterministic review and is built on the same Go engine as the CLI."
	}
}
