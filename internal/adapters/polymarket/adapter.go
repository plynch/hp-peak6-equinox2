package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"time"

	"equinox/internal/schedule"
)

type RawMarket struct {
	EventID       string   `json:"event_id"`
	EventTitle    string   `json:"event_title"`
	EventFamily   string   `json:"event_family"`
	MarketID      string   `json:"market_id"`
	Question      string   `json:"question"`
	Category      string   `json:"category"`
	MarketType    string   `json:"market_type"`
	Outcomes      []string `json:"outcomes"`
	RulesPrimary  string   `json:"rules_primary_source"`
	RulesText     string   `json:"rules_text"`
	EndDateISO    string   `json:"end_date_iso"`
	QuoteYesBid   float64  `json:"quote_yes_bid"`
	QuoteYesAsk   float64  `json:"quote_yes_ask"`
	QuoteNoBid    float64  `json:"quote_no_bid"`
	QuoteNoAsk    float64  `json:"quote_no_ask"`
	DepthNotional float64  `json:"depth_notional"`
	QuoteObserved bool     `json:"quote_observed"`
	QuoteFreshAt  string   `json:"quote_fresh_at"`
}

type Adapter struct{}

var eplEventSlugPattern = regexp.MustCompile(`^epl-[a-z0-9]+-[a-z0-9]+-\d{4}-\d{2}-\d{2}$`)
var fedDecisionSlugPattern = regexp.MustCompile(`^fed-decision-in-[a-z0-9-]+$`)

func (a Adapter) LoadFixture(path string) ([]RawMarket, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []RawMarket
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (a Adapter) LiveInspect(ctx context.Context, limit int) ([]RawMarket, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://gamma-api.polymarket.com/markets?closed=false&limit=%d", limit), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("polymarket status %d", resp.StatusCode)
	}
	var rows []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	results := make([]RawMarket, 0, len(rows))
	now := time.Now().UTC().Format(time.RFC3339)
	for _, r := range rows {
		id, _ := r["id"].(string)
		q, _ := r["question"].(string)
		results = append(results, RawMarket{
			EventID:       fmt.Sprintf("pm-live-%s", id),
			EventTitle:    q,
			EventFamily:   "live-inspect",
			MarketID:      id,
			Question:      q,
			Category:      "live",
			MarketType:    "binary",
			Outcomes:      []string{"Yes", "No"},
			QuoteObserved: false,
			QuoteFreshAt:  now,
		})
	}
	return results, nil
}

func (a Adapter) LivePremierLeague(ctx context.Context, matchweekLimit int) ([]RawMarket, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://gamma-api.polymarket.com/events?tag_slug=premier-league&closed=false&limit=200", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("polymarket status %d", resp.StatusCode)
	}

	var events []liveEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	filtered := make([]liveEvent, 0, len(events))
	for _, event := range events {
		if len(event.Markets) == 0 {
			continue
		}
		if !eplEventSlugPattern.MatchString(event.Slug) {
			continue
		}
		eventTime := parseTime(event.Markets[0].EndDate)
		if eventTime != nil && eventTime.Before(now.Add(-2*time.Hour)) {
			continue
		}
		filtered = append(filtered, event)
	}

	sort.Slice(filtered, func(i, j int) bool {
		ti := parseTime(filtered[i].Markets[0].EndDate)
		tj := parseTime(filtered[j].Markets[0].EndDate)
		if ti == nil || tj == nil {
			return filtered[i].Slug < filtered[j].Slug
		}
		return ti.Before(*tj)
	})
	filtered = schedule.KeepUpcomingWindows(filtered, func(event liveEvent) *time.Time {
		if len(event.Markets) == 0 {
			return nil
		}
		return parseTime(event.Markets[0].EndDate)
	}, matchweekLimit, 72*time.Hour)

	results := make([]RawMarket, 0, len(filtered)*3)
	for _, event := range filtered {
		for _, market := range event.Markets {
			outcomes := parseOutcomeString(market.Outcomes)
			if len(outcomes) == 0 {
				outcomes = []string{"Yes", "No"}
			}
			bestBid := valueOrZero(market.BestBid)
			bestAsk := valueOrZero(market.BestAsk)
			results = append(results, RawMarket{
				EventID:       event.Slug,
				EventTitle:    event.Title,
				EventFamily:   "soccer_big_five",
				MarketID:      market.SlugOrID(),
				Question:      market.Question,
				Category:      "sports",
				MarketType:    "binary",
				Outcomes:      outcomes,
				RulesPrimary:  market.ResolutionSource,
				RulesText:     market.Description,
				EndDateISO:    market.EndDate,
				QuoteYesBid:   bestBid,
				QuoteYesAsk:   bestAsk,
				QuoteNoBid:    complement(bestAsk),
				QuoteNoAsk:    complement(bestBid),
				DepthNotional: valueOrZero(market.LiquidityNum),
				QuoteObserved: market.BestBid != nil || market.BestAsk != nil,
				QuoteFreshAt:  coalesceString(market.UpdatedAt, time.Now().UTC().Format(time.RFC3339)),
			})
		}
	}
	return results, nil
}

func (a Adapter) LiveFedDecision(ctx context.Context, meetingLimit int) ([]RawMarket, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://gamma-api.polymarket.com/events?closed=false&limit=1000", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("polymarket status %d", resp.StatusCode)
	}

	var events []liveEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	filtered := make([]liveEvent, 0, len(events))
	for _, event := range events {
		if len(event.Markets) == 0 {
			continue
		}
		if !fedDecisionSlugPattern.MatchString(event.Slug) {
			continue
		}
		eventTime := parseTime(event.EndDate)
		if eventTime == nil {
			eventTime = parseTime(event.Markets[0].EndDate)
		}
		if eventTime != nil && eventTime.Before(now.Add(-24*time.Hour)) {
			continue
		}
		filtered = append(filtered, event)
	}

	sort.Slice(filtered, func(i, j int) bool {
		ti := parseTime(filtered[i].EndDate)
		tj := parseTime(filtered[j].EndDate)
		if ti == nil || tj == nil {
			return filtered[i].Slug < filtered[j].Slug
		}
		return ti.Before(*tj)
	})
	if meetingLimit > 0 && len(filtered) > meetingLimit {
		filtered = filtered[:meetingLimit]
	}

	results := make([]RawMarket, 0, len(filtered)*4)
	for _, event := range filtered {
		for _, market := range event.Markets {
			outcomes := parseOutcomeString(market.Outcomes)
			if len(outcomes) == 0 {
				outcomes = []string{"Yes", "No"}
			}
			bestBid := valueOrZero(market.BestBid)
			bestAsk := valueOrZero(market.BestAsk)
			results = append(results, RawMarket{
				EventID:       event.Slug,
				EventTitle:    event.Title,
				EventFamily:   "fed_fomc",
				MarketID:      market.SlugOrID(),
				Question:      market.Question,
				Category:      "economy",
				MarketType:    "binary",
				Outcomes:      outcomes,
				RulesPrimary:  coalesceString(market.ResolutionSource, event.ResolutionSource),
				RulesText:     coalesceString(market.Description, event.Description),
				EndDateISO:    coalesceString(market.EndDate, event.EndDate),
				QuoteYesBid:   bestBid,
				QuoteYesAsk:   bestAsk,
				QuoteNoBid:    complement(bestAsk),
				QuoteNoAsk:    complement(bestBid),
				DepthNotional: valueOrZero(market.LiquidityNum),
				QuoteObserved: market.BestBid != nil || market.BestAsk != nil,
				QuoteFreshAt:  coalesceString(market.UpdatedAt, time.Now().UTC().Format(time.RFC3339)),
			})
		}
	}
	return results, nil
}

func parseTime(raw string) *time.Time {
	if raw == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &t
}

type liveEvent struct {
	Slug             string       `json:"slug"`
	Title            string       `json:"title"`
	EndDate          string       `json:"endDate"`
	Description      string       `json:"description"`
	ResolutionSource string       `json:"resolutionSource"`
	Markets          []liveMarket `json:"markets"`
}

type liveMarket struct {
	ID               any      `json:"id"`
	Slug             string   `json:"slug"`
	Question         string   `json:"question"`
	Description      string   `json:"description"`
	ResolutionSource string   `json:"resolutionSource"`
	EndDate          string   `json:"endDate"`
	UpdatedAt        string   `json:"updatedAt"`
	Outcomes         string   `json:"outcomes"`
	BestBid          *float64 `json:"bestBid"`
	BestAsk          *float64 `json:"bestAsk"`
	LiquidityNum     *float64 `json:"liquidityNum"`
}

func (m liveMarket) SlugOrID() string {
	if m.Slug != "" {
		return m.Slug
	}
	return fmt.Sprint(m.ID)
}

func parseOutcomeString(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

func valueOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func complement(v float64) float64 {
	if v <= 0 {
		return 0
	}
	return 1 - v
}

func coalesceString(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
