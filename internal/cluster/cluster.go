package cluster

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"equinox/internal/model"
)

func BuildEventClusters(instances []model.VenueMarketInstance) []model.EventCluster {
	sorted := append([]model.VenueMarketInstance(nil), instances...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].InstanceID < sorted[j].InstanceID })

	type working struct {
		members []model.VenueMarketInstance
	}
	var groups []working

	for _, in := range sorted {
		bestIdx := -1
		bestScore := 0.0
		for i := range groups {
			s := eventSimilarity(in, groups[i].members)
			if s > bestScore {
				bestScore = s
				bestIdx = i
			}
		}
		if bestIdx >= 0 && bestScore >= 0.72 {
			groups[bestIdx].members = append(groups[bestIdx].members, in)
		} else {
			groups = append(groups, working{members: []model.VenueMarketInstance{in}})
		}
	}

	events := make([]model.EventCluster, 0, len(groups))
	for _, g := range groups {
		canonical := eventCanonicalKey(g.members)
		title := representativeTitle(g.members)
		events = append(events, model.EventCluster{
			CanonicalKey:    canonical,
			Title:           title,
			Confidence:      confidenceEvent(g.members),
			AmbiguityNotes:  collectAmbiguity(g.members),
			MarketInstances: g.members,
		})
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].CanonicalKey == events[j].CanonicalKey {
			return events[i].Title < events[j].Title
		}
		return events[i].CanonicalKey < events[j].CanonicalKey
	})
	for i := range events {
		events[i].ClusterID = fmt.Sprintf("event-%03d", i+1)
	}
	return events
}

func BuildPropositionClusters(events []model.EventCluster) ([]model.PropositionCluster, []model.EquivalenceAssessment) {
	var pcs []model.PropositionCluster
	var assessments []model.EquivalenceAssessment

	for _, e := range events {
		groups := groupPropositions(e.MarketInstances)
		for _, members := range groups {
			routeability, reasons, amb, conf := classify(members)
			pc := model.PropositionCluster{
				EventClusterID:  e.ClusterID,
				CanonicalKey:    propositionCanonicalKey(e.CanonicalKey, members),
				Proposition:     representativeProposition(members),
				Confidence:      conf,
				Routeability:    routeability,
				RefusalReasons:  reasons,
				AmbiguityNotes:  amb,
				MarketInstances: members,
			}
			pcs = append(pcs, pc)
		}
	}

	sort.Slice(pcs, func(i, j int) bool {
		if pcs[i].EventClusterID == pcs[j].EventClusterID {
			return pcs[i].CanonicalKey < pcs[j].CanonicalKey
		}
		return pcs[i].EventClusterID < pcs[j].EventClusterID
	})
	for i := range pcs {
		pcs[i].ClusterID = fmt.Sprintf("prop-%03d", i+1)
	}

	asID := 1
	for _, p := range pcs {
		ids := make([]string, 0, len(p.MarketInstances))
		for _, m := range p.MarketInstances {
			ids = append(ids, m.InstanceID)
		}
		classification := string(p.Routeability)
		if p.Routeability == model.Routeable {
			classification = "strong_proposition_match"
		}
		reasons := append([]string{}, p.RefusalReasons...)
		reasons = append(reasons, p.AmbiguityNotes...)
		if len(reasons) == 0 {
			reasons = []string{"high proposition text similarity with compatible semantics"}
		}
		assessments = append(assessments, model.EquivalenceAssessment{
			AssessmentID:         fmt.Sprintf("assess-%03d", asID),
			EventClusterID:       p.EventClusterID,
			PropositionClusterID: p.ClusterID,
			CandidateInstanceIDs: ids,
			Classification:       classification,
			Confidence:           p.Confidence,
			Reasons:              reasons,
		})
		asID++
	}
	for _, e := range events {
		count := 0
		for _, p := range pcs {
			if p.EventClusterID == e.ClusterID {
				count++
			}
		}
		if count > 1 {
			assessments = append(assessments, model.EquivalenceAssessment{
				AssessmentID:   fmt.Sprintf("assess-%03d", asID),
				EventClusterID: e.ClusterID,
				Classification: "event_only_match",
				Confidence:     e.Confidence - 0.2,
				Reasons:        []string{"same event cluster but proposition semantics diverge"},
			})
			asID++
		}

		// Emit explicit paired non-match assessments across venues inside the same event when proposition similarity is low.
		for i := 0; i < len(e.MarketInstances); i++ {
			for j := i + 1; j < len(e.MarketInstances); j++ {
				a := e.MarketInstances[i]
				b := e.MarketInstances[j]
				if a.Venue == b.Venue {
					continue
				}
				sim := propositionSimilarity(a, []model.VenueMarketInstance{b})
				if sim < 0.55 {
					assessments = append(assessments, model.EquivalenceAssessment{
						AssessmentID:         fmt.Sprintf("assess-%03d", asID),
						EventClusterID:       e.ClusterID,
						CandidateInstanceIDs: []string{a.InstanceID, b.InstanceID},
						Classification:       "explicit_non_match",
						Confidence:           0.85,
						Reasons:              []string{"same event family but proposition semantics diverge strongly", fmt.Sprintf("proposition similarity %.2f below non-match threshold", sim)},
					})
					asID++
				}
			}
		}
	}
	return pcs, assessments
}

func groupPropositions(instances []model.VenueMarketInstance) [][]model.VenueMarketInstance {
	sorted := append([]model.VenueMarketInstance(nil), instances...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].InstanceID < sorted[j].InstanceID })
	var groups [][]model.VenueMarketInstance
	for _, in := range sorted {
		best := -1
		bestScore := 0.0
		for i := range groups {
			s := propositionSimilarity(in, groups[i])
			if s > bestScore {
				bestScore = s
				best = i
			}
		}
		if best >= 0 && bestScore >= 0.68 {
			groups[best] = append(groups[best], in)
		} else {
			groups = append(groups, []model.VenueMarketInstance{in})
		}
	}
	return groups
}

func classify(members []model.VenueMarketInstance) (model.Routeability, []string, []string, float64) {
	reasons := []string{}
	amb := []string{}
	venues := map[model.Venue]bool{}
	for _, m := range members {
		venues[m.Venue] = true
		if m.UnsupportedShape {
			reasons = append(reasons, "unsupported market shape")
		}
		if m.HasOtherOutcome {
			reasons = append(reasons, "contains Other/placeholder style outcomes")
		}
		if !m.BinaryLike {
			reasons = append(reasons, "not simple binary yes/no")
		}
		if len(m.AmbiguityNotes) > 0 {
			amb = append(amb, m.AmbiguityNotes...)
		}
	}
	if len(reasons) > 0 {
		return model.Unsupported, dedupe(reasons), dedupe(amb), 0.45
	}
	if len(amb) > 0 {
		return model.Ambiguous, []string{"ambiguous resolution/deadline semantics"}, dedupe(amb), 0.5
	}
	if len(venues) < 2 {
		return model.EventOnly, []string{"single venue proposition instance"}, nil, 0.6
	}
	return model.Routeable, nil, nil, 0.82
}

func eventSimilarity(a model.VenueMarketInstance, members []model.VenueMarketInstance) float64 {
	if len(members) == 0 {
		return 0
	}
	score := 0.0
	for _, m := range members {
		s := 0.0
		if strings.EqualFold(a.EventFamily, m.EventFamily) {
			s += 0.18
		}
		if strings.EqualFold(a.Category, m.Category) {
			s += 0.07
		}
		s += 0.65 * jaccard(tokens(a.EventTitle), tokens(m.EventTitle))
		if closeDeadline(a.DeadlineUTC, m.DeadlineUTC, 72*time.Hour) {
			s += 0.10
		}
		if s > score {
			score = s
		}
	}
	return score
}

func propositionSimilarity(a model.VenueMarketInstance, members []model.VenueMarketInstance) float64 {
	if len(members) == 0 {
		return 0
	}
	best := 0.0
	for _, m := range members {
		s := 0.0
		s += 0.65 * jaccard(tokens(a.NormalizedYesTarget), tokens(m.NormalizedYesTarget))
		s += 0.20 * jaccard(tokens(a.MarketTitle), tokens(m.MarketTitle))
		if closeDeadline(a.DeadlineUTC, m.DeadlineUTC, 8*time.Hour) {
			s += 0.15
		}
		if s > best {
			best = s
		}
	}
	return best
}

func closeDeadline(a, b *time.Time, maxDelta time.Duration) bool {
	if a == nil || b == nil {
		return false
	}
	d := a.Sub(*b)
	if d < 0 {
		d = -d
	}
	return d <= maxDelta
}

func eventCanonicalKey(members []model.VenueMarketInstance) string {
	if len(members) == 0 {
		return "event:unknown"
	}
	family := strings.ToLower(strings.TrimSpace(members[0].EventFamily))
	if family == "" {
		family = "unknown"
	}
	base := tokenFingerprint(tokens(representativeTitle(members)), 4)
	return fmt.Sprintf("%s:%s", family, base)
}

func propositionCanonicalKey(eventKey string, members []model.VenueMarketInstance) string {
	base := tokenFingerprint(tokens(representativeProposition(members)), 5)
	return eventKey + ":" + base
}

func representativeTitle(members []model.VenueMarketInstance) string {
	best := members[0].EventTitle
	for _, m := range members[1:] {
		if len(m.EventTitle) < len(best) {
			best = m.EventTitle
		}
	}
	return strings.ToLower(best)
}

func representativeProposition(members []model.VenueMarketInstance) string {
	best := members[0].NormalizedYesTarget
	for _, m := range members[1:] {
		if len(m.NormalizedYesTarget) < len(best) {
			best = m.NormalizedYesTarget
		}
	}
	return best
}

func collectAmbiguity(members []model.VenueMarketInstance) []string {
	var notes []string
	for _, m := range members {
		for _, n := range m.AmbiguityNotes {
			notes = append(notes, fmt.Sprintf("%s: %s", m.InstanceID, n))
		}
	}
	return dedupe(notes)
}

func confidenceEvent(ms []model.VenueMarketInstance) float64 {
	if len(ms) <= 1 {
		return 0.55
	}
	venues := map[model.Venue]bool{}
	for _, m := range ms {
		venues[m.Venue] = true
	}
	c := 0.65 + float64(len(venues)-1)*0.15
	if c > 0.95 {
		c = 0.95
	}
	return c
}

func tokens(s string) []string {
	s = strings.ToLower(strings.TrimSpace(s))
	repl := strings.NewReplacer("/", " ", "-", " ", "_", " ", ":", " ", ",", " ", ".", " ", "?", " ", "!", " ", "&", " ")
	s = repl.Replace(s)
	parts := strings.Fields(s)
	stop := map[string]bool{"the": true, "a": true, "an": true, "will": true, "at": true, "in": true, "vs": true, "fc": true, "afc": true}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if stop[p] {
			continue
		}
		if isNumericToken(p) {
			continue
		}
		if month := canonicalMonthToken(p); month != "" {
			p = month
		}
		out = append(out, p)
	}
	return out
}

func jaccard(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	ma := map[string]bool{}
	mb := map[string]bool{}
	for _, x := range a {
		ma[x] = true
	}
	for _, x := range b {
		mb[x] = true
	}
	inter := 0
	union := map[string]bool{}
	for k := range ma {
		union[k] = true
		if mb[k] {
			inter++
		}
	}
	for k := range mb {
		union[k] = true
	}
	return float64(inter) / float64(len(union))
}

func tokenFingerprint(tok []string, n int) string {
	if len(tok) == 0 {
		return "unknown"
	}
	unique := dedupe(tok)
	sort.Strings(unique)
	if len(unique) > n {
		unique = unique[:n]
	}
	return strings.Join(unique, "-")
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

func canonicalMonthToken(token string) string {
	switch token {
	case "january", "jan":
		return "jan"
	case "february", "feb":
		return "feb"
	case "march", "mar":
		return "mar"
	case "april", "apr":
		return "apr"
	case "may":
		return "may"
	case "june", "jun":
		return "jun"
	case "july", "jul":
		return "jul"
	case "august", "aug":
		return "aug"
	case "september", "sep":
		return "sep"
	case "october", "oct":
		return "oct"
	case "november", "nov":
		return "nov"
	case "december", "dec":
		return "dec"
	default:
		return ""
	}
}

func isNumericToken(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
