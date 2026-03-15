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

func parseTimeOrNow(raw string) *time.Time {
	if t := parseTime(raw); t != nil {
		return t
	}
	n := time.Now().UTC()
	return &n
}

func normalizePhrase(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	repl := strings.NewReplacer("?", "", ",", " ", ".", " ", ":", " ", "-", " ")
	s = repl.Replace(s)
	parts := strings.Fields(s)
	stop := map[string]bool{"the": true, "will": true, "at": true, "a": true, "an": true, "if": true, "on": true, "fc": true}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if stop[p] {
			continue
		}
		if isNumericToken(p) {
			continue
		}
		if p == "tie" {
			p = "draw"
		}
		if strings.HasSuffix(p, "s") && len(p) > 4 {
			p = strings.TrimSuffix(p, "s")
		}
		out = append(out, p)
	}
	return strings.Join(out, " ")
}

func inferNormalizedTarget(question string, outcomes []string, selection string) string {
	if sportsTarget := inferSportsTarget(question, selection); sportsTarget != "" {
		return sportsTarget
	}
	q := normalizePhrase(question)
	if len(outcomes) > 2 {
		q = q + " multi-outcome"
	}
	return strings.TrimSpace(q)
}

func inferSportsTarget(question string, selection string) string {
	lower := strings.ToLower(strings.TrimSpace(question))

	if strings.Contains(lower, "both teams score") {
		return "both teams score"
	}
	if strings.Contains(lower, "end in a draw") {
		return "draw"
	}
	if strings.Contains(lower, " winner") && strings.TrimSpace(selection) != "" {
		canonical := canonicalSelection(selection)
		if canonical == "draw" {
			return "draw"
		}
		return canonical + " win"
	}
	if strings.HasPrefix(lower, "will ") && strings.Contains(lower, " win") {
		subject := lower[len("will "):]
		subject = strings.Split(subject, " win")[0]
		canonical := canonicalSelection(subject)
		if canonical != "" {
			return canonical + " win"
		}
	}
	return ""
}

func canonicalSelection(selection string) string {
	s := normalizePhrase(selection)
	switch s {
	case "tie":
		return "draw"
	case "draw":
		return "draw"
	}
	repl := strings.NewReplacer(
		"arsenal fc", "arsenal",
		"aston villa fc", "aston villa",
		"bournemouth fc", "bournemouth",
		"brentford fc", "brentford",
		"brighton hove albion", "brighton",
		"chelsea fc", "chelsea",
		"crystal palace fc", "crystal palace",
		"everton fc", "everton",
		"fulham fc", "fulham",
		"ipswich town", "ipswich",
		"leeds united", "leeds united",
		"tottenham hotspur", "tottenham",
		"tottenham hotspur fc", "tottenham",
		"spurs", "tottenham",
		"manchester city fc", "manchester city",
		"manchester united fc", "manchester united",
		"liverpool football club", "liverpool",
		"liverpool fc", "liverpool",
		"newcastle united", "newcastle",
		"newcastle united fc", "newcastle",
		"nottingham forest", "nottingham forest",
		"nottingham forest fc", "nottingham forest",
		"nottingham", "nottingham forest",
		"west ham united", "west ham",
		"west ham united fc", "west ham",
		"wolverhampton wanderers", "wolverhampton",
		"wolverhampton wanderers fc", "wolverhampton",
		"wolves", "wolverhampton",
	)
	s = repl.Replace(s)
	return strings.TrimSpace(s)
}

func isNumericToken(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func inferUnsupported(binaryLike bool, hasOther bool, marketType string, outcomes []string, text string) bool {
	if !binaryLike || hasOther {
		return true
	}
	mt := strings.ToLower(marketType)
	if strings.Contains(mt, "scalar") || strings.Contains(mt, "range") || strings.Contains(mt, "multi") {
		return true
	}
	if len(outcomes) > 2 {
		return true
	}
	lower := strings.ToLower(text)
	return containsWord(lower, "other") || containsWord(lower, "bucket") || strings.Contains(lower, "multi-outcome")
}

func inferBinaryLike(outcomes []string, marketType string) (bool, bool) {
	if len(outcomes) == 0 {
		mt := strings.ToLower(marketType)
		return strings.Contains(mt, "binary"), false
	}
	hasOther := false
	for _, o := range outcomes {
		if strings.EqualFold(strings.TrimSpace(o), "other") {
			hasOther = true
		}
	}
	return len(outcomes) == 2 && !hasOther, hasOther
}

func inferAmbiguity(marketTitle, rules string, deadline *time.Time) []string {
	var notes []string
	text := strings.ToLower(marketTitle + " " + rules)
	excludesExtraTime := strings.Contains(text, "does not include extra time") ||
		strings.Contains(text, "does not include extra time or penalties") ||
		strings.Contains(text, "no extra time or penalties")
	if strings.Contains(text, "advance") ||
		((strings.Contains(text, "extra time") || strings.Contains(text, "penalties")) && !excludesExtraTime) {
		notes = append(notes, "qualification/extra-time semantics may diverge from regulation-only contracts")
	}
	if strings.Contains(text, "subject to") || strings.Contains(text, "discretion") {
		notes = append(notes, "resolution governance contains discretionary wording")
	}
	if deadline == nil {
		notes = append(notes, "deadline missing or unparsable")
	}
	return dedupe(notes)
}

func inferDeadlineProvenance(deadlineRaw string, rules string) string {
	if parseTime(deadlineRaw) != nil {
		return "explicit_market_deadline"
	}
	if strings.TrimSpace(rules) != "" {
		return "rules_text_only"
	}
	return "missing"
}

func inferResolutionSource(primary, rules string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	if strings.TrimSpace(rules) != "" {
		return "derived_from_rules_text"
	}
	return "unknown"
}

func FromPolymarket(rows []polymarket.RawMarket) []model.VenueMarketInstance {
	out := make([]model.VenueMarketInstance, 0, len(rows))
	for _, r := range rows {
		eventKey := fmt.Sprintf("%s:%s", slug(r.EventFamily), slug(r.EventTitle))
		binaryLike, hasOther := inferBinaryLike(r.Outcomes, r.MarketType)
		normTarget := inferNormalizedTarget(r.Question, r.Outcomes, "")
		deadline := parseTime(r.EndDateISO)
		unsupported := inferUnsupported(binaryLike, hasOther, r.MarketType, r.Outcomes, r.Question+" "+r.RulesText)
		amb := inferAmbiguity(r.Question, r.RulesText, deadline)
		prop := fmt.Sprintf("%s:%s", eventKey, slug(normTarget))
		out = append(out, model.VenueMarketInstance{
			InstanceID:          fmt.Sprintf("polymarket:%s", r.MarketID),
			Venue:               model.VenuePolymarket,
			VenueMarketID:       r.MarketID,
			VenueEventID:        r.EventID,
			EventTitle:          r.EventTitle,
			MarketTitle:         r.Question,
			EventFamily:         r.EventFamily,
			Category:            r.Category,
			BinaryLike:          binaryLike,
			HasOtherOutcome:     hasOther,
			UnsupportedShape:    unsupported,
			AmbiguityNotes:      amb,
			ResolutionSource:    inferResolutionSource(r.RulesPrimary, r.RulesText),
			DeadlineUTC:         deadline,
			DeadlineProvenance:  inferDeadlineProvenance(r.EndDateISO, r.RulesText),
			CanonicalEventKey:   eventKey,
			CanonicalPropHint:   prop,
			NormalizedYesTarget: normTarget,
			Quote:               model.QuoteView{YesBid: r.QuoteYesBid, YesAsk: r.QuoteYesAsk, NoBid: r.QuoteNoBid, NoAsk: r.QuoteNoAsk, DepthNotional: r.DepthNotional, FreshAt: *parseTimeOrNow(r.QuoteFreshAt), Observed: r.QuoteObserved},
			RawRef:              fmt.Sprintf("fixture://polymarket/%s", r.MarketID),
		})
	}
	return out
}

func FromKalshi(rows []kalshi.RawMarket) []model.VenueMarketInstance {
	out := make([]model.VenueMarketInstance, 0, len(rows))
	for _, r := range rows {
		eventKey := fmt.Sprintf("%s:%s", slug(r.EventFamily), slug(r.EventTitle))
		binaryLike, hasOther := inferBinaryLike(r.Outcomes, r.MarketType)
		normTarget := inferNormalizedTarget(r.Title, r.Outcomes, r.YesSubTitle)
		deadline := parseTime(r.CloseTimeISO)
		unsupported := inferUnsupported(binaryLike, hasOther, r.MarketType, r.Outcomes, r.Title+" "+r.RulesText)
		amb := inferAmbiguity(r.Title, r.RulesText+" "+r.SettlementNotes, deadline)
		prop := fmt.Sprintf("%s:%s", eventKey, slug(normTarget))
		out = append(out, model.VenueMarketInstance{
			InstanceID:          fmt.Sprintf("kalshi:%s", r.MarketTicker),
			Venue:               model.VenueKalshi,
			VenueMarketID:       r.MarketTicker,
			VenueEventID:        r.EventID,
			EventTitle:          r.EventTitle,
			MarketTitle:         r.Title,
			EventFamily:         r.EventFamily,
			Category:            r.Category,
			BinaryLike:          binaryLike,
			HasOtherOutcome:     hasOther,
			UnsupportedShape:    unsupported,
			AmbiguityNotes:      amb,
			ResolutionSource:    inferResolutionSource(r.RulesPrimary, r.RulesText),
			DeadlineUTC:         deadline,
			DeadlineProvenance:  inferDeadlineProvenance(r.CloseTimeISO, r.RulesText),
			CanonicalEventKey:   eventKey,
			CanonicalPropHint:   prop,
			NormalizedYesTarget: normTarget,
			Quote:               model.QuoteView{YesBid: r.YesBidCents / 100.0, YesAsk: r.YesAskCents / 100.0, NoBid: r.NoBidCents / 100.0, NoAsk: r.NoAskCents / 100.0, DepthNotional: r.DepthNotional, FreshAt: *parseTimeOrNow(r.QuoteFreshAt), Observed: r.QuoteObserved},
			RawRef:              fmt.Sprintf("fixture://kalshi/%s", r.MarketTicker),
		})
	}
	return out
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func containsWord(text, word string) bool {
	repl := strings.NewReplacer("/", " ", "-", " ", "_", " ", ":", " ", ",", " ", ".", " ", "?", " ", "!", " ", "\"", " ", "'", " ", "(", " ", ")", " ")
	parts := strings.Fields(repl.Replace(strings.ToLower(text)))
	for _, part := range parts {
		if part == word {
			return true
		}
	}
	return false
}
