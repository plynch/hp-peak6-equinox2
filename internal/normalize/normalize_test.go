package normalize

import (
	"testing"

	"equinox/internal/adapters/polymarket"
)

func TestDerivesNormalizationSignalsFromVenueStyleFields(t *testing.T) {
	rows := []polymarket.RawMarket{
		{
			EventID:      "e1",
			EventTitle:   "Barcelona vs Real Madrid next week",
			EventFamily:  "soccer_big_five",
			MarketID:     "m1",
			Question:     "How many goals will be scored in El Clasico?",
			Category:     "sports",
			MarketType:   "range",
			Outcomes:     []string{"0-1", "2-3", "4+", "Other"},
			RulesText:    "Bucketed market with Other outcome",
			EndDateISO:   "",
			QuoteYesAsk:  0.55,
			QuoteNoAsk:   0.45,
			QuoteFreshAt: "2026-03-14T00:00:00Z",
		},
	}
	instances := FromPolymarket(rows)
	if len(instances) != 1 {
		t.Fatalf("expected one instance")
	}
	in := instances[0]
	if !in.UnsupportedShape {
		t.Fatalf("expected unsupported shape to be derived")
	}
	if !in.HasOtherOutcome {
		t.Fatalf("expected other-outcome detection to be derived")
	}
	if in.BinaryLike {
		t.Fatalf("expected non-binary inference from outcomes")
	}
	if in.DeadlineProvenance != "rules_text_only" {
		t.Fatalf("expected rules_text_only provenance, got %s", in.DeadlineProvenance)
	}
}
