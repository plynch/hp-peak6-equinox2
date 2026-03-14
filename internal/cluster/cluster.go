package cluster

import (
	"fmt"
	"sort"
	"strings"

	"equinox/internal/model"
)

func BuildEventClusters(instances []model.VenueMarketInstance) []model.EventCluster {
	byKey := map[string][]model.VenueMarketInstance{}
	for _, in := range instances {
		byKey[in.CanonicalEventKey] = append(byKey[in.CanonicalEventKey], in)
	}
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]model.EventCluster, 0, len(keys))
	for i, k := range keys {
		m := byKey[k]
		amb := []string{}
		for _, inst := range m {
			if len(inst.AmbiguityNotes) > 0 {
				amb = append(amb, fmt.Sprintf("%s: %s", inst.InstanceID, strings.Join(inst.AmbiguityNotes, "; ")))
			}
		}
		out = append(out, model.EventCluster{ClusterID: fmt.Sprintf("event-%03d", i+1), CanonicalKey: k, Title: m[0].EventTitle, Confidence: confidenceEvent(m), AmbiguityNotes: amb, MarketInstances: m})
	}
	return out
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

func BuildPropositionClusters(events []model.EventCluster) ([]model.PropositionCluster, []model.EquivalenceAssessment) {
	var pcs []model.PropositionCluster
	var assessments []model.EquivalenceAssessment
	pcID := 1
	asID := 1
	for _, e := range events {
		byProp := map[string][]model.VenueMarketInstance{}
		for _, m := range e.MarketInstances {
			byProp[m.CanonicalPropHint] = append(byProp[m.CanonicalPropHint], m)
		}
		for key, members := range byProp {
			routeability, reasons, amb, conf := classify(members)
			pc := model.PropositionCluster{
				ClusterID:       fmt.Sprintf("prop-%03d", pcID),
				EventClusterID:  e.ClusterID,
				CanonicalKey:    key,
				Proposition:     members[0].NormalizedYesTarget,
				Confidence:      conf,
				Routeability:    routeability,
				RefusalReasons:  reasons,
				AmbiguityNotes:  amb,
				MarketInstances: members,
			}
			pcs = append(pcs, pc)
			ids := make([]string, 0, len(members))
			for _, m := range members {
				ids = append(ids, m.InstanceID)
			}
			classification := string(routeability)
			if routeability == model.Routeable {
				classification = "strong_proposition_match"
			}
			assessments = append(assessments, model.EquivalenceAssessment{
				AssessmentID:         fmt.Sprintf("assess-%03d", asID),
				EventClusterID:       e.ClusterID,
				PropositionClusterID: pc.ClusterID,
				CandidateInstanceIDs: ids,
				Classification:       classification,
				Confidence:           conf,
				Reasons:              append(append([]string{}, reasons...), amb...),
			})
			asID++
			pcID++
		}
		if len(byProp) > 1 {
			assessments = append(assessments, model.EquivalenceAssessment{
				AssessmentID:   fmt.Sprintf("assess-%03d", asID),
				EventClusterID: e.ClusterID,
				Classification: "event_only_match",
				Confidence:     e.Confidence - 0.2,
				Reasons:        []string{"same canonical event, proposition hints diverge"},
			})
			asID++
		}
	}
	sort.Slice(pcs, func(i, j int) bool { return pcs[i].ClusterID < pcs[j].ClusterID })
	return pcs, assessments
}

func classify(members []model.VenueMarketInstance) (model.Routeability, []string, []string, float64) {
	reasons := []string{}
	amb := []string{}
	venues := map[model.Venue]bool{}
	conf := 0.8
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
		return model.Unsupported, dedupe(reasons), amb, 0.45
	}
	if len(venues) < 2 {
		return model.EventOnly, []string{"single venue proposition instance"}, amb, 0.6
	}
	if len(amb) > 0 {
		return model.Ambiguous, []string{"ambiguous resolution/deadline semantics"}, dedupe(amb), 0.5
	}
	return model.Routeable, nil, nil, conf
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
