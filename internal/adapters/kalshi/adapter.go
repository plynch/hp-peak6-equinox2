package kalshi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type RawMarket struct {
	EventID         string   `json:"event_id"`
	EventTitle      string   `json:"event_title"`
	EventFamily     string   `json:"event_family"`
	MarketTicker    string   `json:"market_ticker"`
	Title           string   `json:"title"`
	YesSubTitle     string   `json:"yes_sub_title"`
	Category        string   `json:"category"`
	MarketType      string   `json:"market_type"`
	Outcomes        []string `json:"outcomes"`
	RulesPrimary    string   `json:"rules_primary_source"`
	RulesText       string   `json:"rules_text"`
	CloseTimeISO    string   `json:"close_time_iso"`
	YesBidCents     float64  `json:"yes_bid_cents"`
	YesAskCents     float64  `json:"yes_ask_cents"`
	NoBidCents      float64  `json:"no_bid_cents"`
	NoAskCents      float64  `json:"no_ask_cents"`
	DepthNotional   float64  `json:"depth_notional"`
	QuoteObserved   bool     `json:"quote_observed"`
	QuoteFreshAt    string   `json:"quote_fresh_at"`
	SettlementNotes string   `json:"settlement_notes"`
}

type Adapter struct{}

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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.elections.kalshi.com/trade-api/v2/markets?limit=%d&status=open", limit), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kalshi status %d", resp.StatusCode)
	}
	var payload struct {
		Markets []map[string]any `json:"markets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	results := make([]RawMarket, 0, len(payload.Markets))
	for _, r := range payload.Markets {
		ticker, _ := r["ticker"].(string)
		title, _ := r["title"].(string)
		results = append(results, RawMarket{
			EventID:      fmt.Sprintf("kalshi-live-%s", ticker),
			EventTitle:   title,
			EventFamily:  "live-inspect",
			MarketTicker: ticker,
			Title:        title,
			Category:     "live",
			MarketType:   "binary",
			Outcomes:     []string{"Yes", "No"},
			QuoteFreshAt: now,
		})
	}
	return results, nil
}

func (a Adapter) LivePremierLeague(ctx context.Context, eventLimit int) ([]RawMarket, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.elections.kalshi.com/trade-api/v2/events?series_ticker=KXEPLGAME&limit=200", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kalshi events status %d", resp.StatusCode)
	}

	var payload struct {
		Events []liveEventSummary `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	filtered := make([]liveEventSummary, 0, len(payload.Events))
	for _, event := range payload.Events {
		eventTime := parseTickerDate(event.EventTicker)
		if eventTime == nil {
			continue
		}
		if eventTime.Before(now.Add(-24 * time.Hour)) {
			continue
		}
		filtered = append(filtered, event)
	}

	sort.Slice(filtered, func(i, j int) bool {
		ti := parseTickerDate(filtered[i].EventTicker)
		tj := parseTickerDate(filtered[j].EventTicker)
		if ti == nil || tj == nil {
			return filtered[i].EventTicker < filtered[j].EventTicker
		}
		return ti.Before(*tj)
	})
	if eventLimit > 0 && len(filtered) > eventLimit {
		filtered = filtered[:eventLimit]
	}

	results := make([]RawMarket, 0, len(filtered)*3)
	for _, event := range filtered {
		rows, err := a.loadPremierLeagueEvent(ctx, event.EventTicker)
		if err != nil {
			return nil, err
		}
		results = append(results, rows...)
	}
	return results, nil
}

func (a Adapter) loadPremierLeagueEvent(ctx context.Context, ticker string) ([]RawMarket, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.elections.kalshi.com/trade-api/v2/events/"+ticker, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kalshi event %s status %d", ticker, resp.StatusCode)
	}

	var payload struct {
		Event   liveEventDetail   `json:"event"`
		Markets []liveMarketEntry `json:"markets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	rows := make([]RawMarket, 0, len(payload.Markets))
	for _, market := range payload.Markets {
		if market.Status != "" && market.Status != "active" {
			continue
		}
		closeISO := coalesceTime(market.ExpectedExpirationTime, market.CloseTime)
		rulesText := market.RulesSecondary
		if rulesText == "" {
			rulesText = market.EarlyCloseCondition
		}
		rows = append(rows, RawMarket{
			EventID:         payload.Event.EventTicker,
			EventTitle:      payload.Event.Title,
			EventFamily:     "soccer_big_five",
			MarketTicker:    market.Ticker,
			Title:           market.Title,
			YesSubTitle:     market.YesSubTitle,
			Category:        "sports",
			MarketType:      market.MarketType,
			Outcomes:        []string{"Yes", "No"},
			RulesPrimary:    market.RulesPrimary,
			RulesText:       rulesText,
			CloseTimeISO:    closeISO,
			YesBidCents:     dollarsToCents(market.YesBidDollars),
			YesAskCents:     dollarsToCents(market.YesAskDollars),
			NoBidCents:      dollarsToCents(market.NoBidDollars),
			NoAskCents:      dollarsToCents(market.NoAskDollars),
			DepthNotional:   parseFloat(market.VolumeFP),
			QuoteObserved:   market.YesBidDollars != "" || market.YesAskDollars != "",
			QuoteFreshAt:    coalesceTime(market.UpdatedTime, market.CloseTime),
			SettlementNotes: market.RulesSecondary,
		})
	}
	return rows, nil
}

type liveEventSummary struct {
	EventTicker string `json:"event_ticker"`
}

type liveEventDetail struct {
	EventTicker string `json:"event_ticker"`
	Title       string `json:"title"`
}

type liveMarketEntry struct {
	Ticker                 string `json:"ticker"`
	Title                  string `json:"title"`
	YesSubTitle            string `json:"yes_sub_title"`
	MarketType             string `json:"market_type"`
	Status                 string `json:"status"`
	RulesPrimary           string `json:"rules_primary"`
	RulesSecondary         string `json:"rules_secondary"`
	EarlyCloseCondition    string `json:"early_close_condition"`
	CloseTime              string `json:"close_time"`
	ExpectedExpirationTime string `json:"expected_expiration_time"`
	UpdatedTime            string `json:"updated_time"`
	YesBidDollars          string `json:"yes_bid_dollars"`
	YesAskDollars          string `json:"yes_ask_dollars"`
	NoBidDollars           string `json:"no_bid_dollars"`
	NoAskDollars           string `json:"no_ask_dollars"`
	VolumeFP               string `json:"volume_fp"`
}

func parseFloat(raw string) float64 {
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	return v
}

func dollarsToCents(raw string) float64 {
	return parseFloat(raw) * 100
}

func coalesceTime(values ...string) string {
	for _, value := range values {
		if value != "" && value != "0001-01-01T00:00:00Z" {
			return value
		}
	}
	return time.Now().UTC().Format(time.RFC3339)
}

func parseTickerDate(ticker string) *time.Time {
	const prefix = "KXEPLGAME-"
	if !strings.HasPrefix(ticker, prefix) || len(ticker) < len(prefix)+7 {
		return nil
	}
	datePart := ticker[len(prefix) : len(prefix)+7]
	month, ok := map[string]string{
		"JAN": "Jan",
		"FEB": "Feb",
		"MAR": "Mar",
		"APR": "Apr",
		"MAY": "May",
		"JUN": "Jun",
		"JUL": "Jul",
		"AUG": "Aug",
		"SEP": "Sep",
		"OCT": "Oct",
		"NOV": "Nov",
		"DEC": "Dec",
	}[strings.ToUpper(datePart[2:5])]
	if !ok {
		return nil
	}
	t, err := time.Parse("06Jan02", datePart[0:2]+month+datePart[5:7])
	if err != nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}
