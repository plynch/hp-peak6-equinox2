package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type RawMarket struct {
	EventID            string   `json:"event_id"`
	EventTitle         string   `json:"event_title"`
	EventFamily        string   `json:"event_family"`
	MarketID           string   `json:"market_id"`
	Question           string   `json:"question"`
	YesOutcomeTarget   string   `json:"yes_outcome_target"`
	Category           string   `json:"category"`
	BinaryLike         bool     `json:"binary_like"`
	HasOtherOutcome    bool     `json:"has_other_outcome"`
	UnsupportedShape   bool     `json:"unsupported_shape"`
	AmbiguityNotes     []string `json:"ambiguity_notes"`
	ResolutionSource   string   `json:"resolution_source"`
	DeadlineUTC        string   `json:"deadline_utc"`
	DeadlineProvenance string   `json:"deadline_provenance"`
	QuoteYesBid        float64  `json:"quote_yes_bid"`
	QuoteYesAsk        float64  `json:"quote_yes_ask"`
	QuoteNoBid         float64  `json:"quote_no_bid"`
	QuoteNoAsk         float64  `json:"quote_no_ask"`
	DepthNotional      float64  `json:"depth_notional"`
	QuoteObserved      bool     `json:"quote_observed"`
	QuoteFreshAt       string   `json:"quote_fresh_at"`
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
			BinaryLike:    true,
			QuoteObserved: false,
			QuoteFreshAt:  now,
		})
	}
	return results, nil
}
