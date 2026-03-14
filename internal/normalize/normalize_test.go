package normalize

import (
	"testing"

	"equinox/internal/adapters/kalshi"
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

func TestDerivesComparableSportsWinnerTargetsAcrossVenues(t *testing.T) {
	pmRows := []polymarket.RawMarket{
		{
			EventID:      "pm-epl-1",
			EventTitle:   "Liverpool FC vs. Tottenham Hotspur FC",
			EventFamily:  "soccer_big_five",
			MarketID:     "pm-liv",
			Question:     "Will Liverpool FC win on 2026-03-15?",
			Category:     "sports",
			MarketType:   "binary",
			Outcomes:     []string{"Yes", "No"},
			RulesText:    "90 minutes plus stoppage time only.",
			EndDateISO:   "2026-03-15T16:30:00Z",
			QuoteYesAsk:  0.76,
			QuoteFreshAt: "2026-03-14T23:40:00Z",
		},
	}
	kalshiRows := []kalshi.RawMarket{
		{
			EventID:      "k-epl-1",
			EventTitle:   "Liverpool vs Tottenham",
			EventFamily:  "soccer_big_five",
			MarketTicker: "KXEPLGAME-26MAR15LFCTOT-LFC",
			Title:        "Liverpool vs Tottenham Winner?",
			YesSubTitle:  "Liverpool",
			Category:     "sports",
			MarketType:   "binary",
			Outcomes:     []string{"Yes", "No"},
			RulesText:    "90 minutes plus stoppage time only.",
			CloseTimeISO: "2026-03-15T16:30:00Z",
			YesAskCents:  76,
			QuoteFreshAt: "2026-03-14T23:40:00Z",
		},
	}

	pm := FromPolymarket(pmRows)
	ka := FromKalshi(kalshiRows)
	if pm[0].NormalizedYesTarget != "liverpool win" {
		t.Fatalf("expected polymarket normalized target to be liverpool win, got %q", pm[0].NormalizedYesTarget)
	}
	if ka[0].NormalizedYesTarget != "liverpool win" {
		t.Fatalf("expected kalshi normalized target to be liverpool win, got %q", ka[0].NormalizedYesTarget)
	}
}
