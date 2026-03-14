package kalshi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
