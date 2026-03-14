package normalize

import (
	"fmt"
	"strings"
	"time"

	"equinox/internal/adapters/kalshi"
	"equinox/internal/adapters/polymarket"
	"equinox/internal/model"
)

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	repl := strings.NewReplacer(" ", "-", "/", "-", ":", "", "?", "", ",", "", ".", "")
	return repl.Replace(s)
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

func FromPolymarket(rows []polymarket.RawMarket) []model.VenueMarketInstance {
	out := make([]model.VenueMarketInstance, 0, len(rows))
	for _, r := range rows {
		eventKey := fmt.Sprintf("%s:%s", slug(r.EventFamily), slug(r.EventTitle))
		prop := fmt.Sprintf("%s:%s", eventKey, slug(r.YesOutcomeTarget))
		out = append(out, model.VenueMarketInstance{
			InstanceID:          fmt.Sprintf("polymarket:%s", r.MarketID),
			Venue:               model.VenuePolymarket,
			VenueMarketID:       r.MarketID,
			VenueEventID:        r.EventID,
			EventTitle:          r.EventTitle,
			MarketTitle:         r.Question,
			EventFamily:         r.EventFamily,
			Category:            r.Category,
			BinaryLike:          r.BinaryLike,
			HasOtherOutcome:     r.HasOtherOutcome,
			UnsupportedShape:    r.UnsupportedShape,
			AmbiguityNotes:      append([]string{}, r.AmbiguityNotes...),
			ResolutionSource:    r.ResolutionSource,
			DeadlineUTC:         parseTime(r.DeadlineUTC),
			DeadlineProvenance:  r.DeadlineProvenance,
			CanonicalEventKey:   eventKey,
			CanonicalPropHint:   prop,
			NormalizedYesTarget: strings.ToLower(r.YesOutcomeTarget),
			Quote:               model.QuoteView{YesBid: r.QuoteYesBid, YesAsk: r.QuoteYesAsk, NoBid: r.QuoteNoBid, NoAsk: r.QuoteNoAsk, DepthNotional: r.DepthNotional, FreshAt: *parseTimeOrNow(r.QuoteFreshAt), Observed: r.QuoteObserved},
			RawRef:              fmt.Sprintf("fixture://polymarket/%s", r.MarketID),
		})
	}
	return out
}

func parseTimeOrNow(raw string) *time.Time {
	if t := parseTime(raw); t != nil {
		return t
	}
	n := time.Now().UTC()
	return &n
}

func FromKalshi(rows []kalshi.RawMarket) []model.VenueMarketInstance {
	out := make([]model.VenueMarketInstance, 0, len(rows))
	for _, r := range rows {
		eventKey := fmt.Sprintf("%s:%s", slug(r.EventFamily), slug(r.EventTitle))
		prop := fmt.Sprintf("%s:%s", eventKey, slug(r.YesOutcomeTarget))
		out = append(out, model.VenueMarketInstance{
			InstanceID:          fmt.Sprintf("kalshi:%s", r.MarketTicker),
			Venue:               model.VenueKalshi,
			VenueMarketID:       r.MarketTicker,
			VenueEventID:        r.EventID,
			EventTitle:          r.EventTitle,
			MarketTitle:         r.Title,
			EventFamily:         r.EventFamily,
			Category:            r.Category,
			BinaryLike:          r.BinaryLike,
			HasOtherOutcome:     r.HasOtherOutcome,
			UnsupportedShape:    r.UnsupportedShape,
			AmbiguityNotes:      append([]string{}, r.AmbiguityNotes...),
			ResolutionSource:    r.ResolutionSource,
			DeadlineUTC:         parseTime(r.DeadlineUTC),
			DeadlineProvenance:  r.DeadlineProvenance,
			CanonicalEventKey:   eventKey,
			CanonicalPropHint:   prop,
			NormalizedYesTarget: strings.ToLower(r.YesOutcomeTarget),
			Quote:               model.QuoteView{YesBid: r.YesBidCents / 100.0, YesAsk: r.YesAskCents / 100.0, NoBid: r.NoBidCents / 100.0, NoAsk: r.NoAskCents / 100.0, DepthNotional: r.DepthNotional, FreshAt: *parseTimeOrNow(r.QuoteFreshAt), Observed: r.QuoteObserved},
			RawRef:              fmt.Sprintf("fixture://kalshi/%s", r.MarketTicker),
		})
	}
	return out
}
