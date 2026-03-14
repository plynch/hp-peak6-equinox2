package model

import "time"

type Venue string

const (
	VenuePolymarket Venue = "Polymarket"
	VenueKalshi     Venue = "Kalshi"
)

type Routeability string

const (
	Routeable    Routeability = "routeable"
	EventOnly    Routeability = "event_only"
	Unsupported  Routeability = "unsupported"
	Ambiguous    Routeability = "ambiguous"
	Insufficient Routeability = "insufficient_data"
)

type QuoteView struct {
	YesBid        float64   `json:"yes_bid"`
	YesAsk        float64   `json:"yes_ask"`
	NoBid         float64   `json:"no_bid"`
	NoAsk         float64   `json:"no_ask"`
	DepthNotional float64   `json:"depth_notional"`
	FreshAt       time.Time `json:"fresh_at"`
	Observed      bool      `json:"observed"`
	Notes         []string  `json:"notes,omitempty"`
}

type VenueMarketInstance struct {
	InstanceID          string     `json:"instance_id"`
	Venue               Venue      `json:"venue"`
	VenueMarketID       string     `json:"venue_market_id"`
	VenueEventID        string     `json:"venue_event_id"`
	EventTitle          string     `json:"event_title"`
	MarketTitle         string     `json:"market_title"`
	EventFamily         string     `json:"event_family"`
	Sport               string     `json:"sport,omitempty"`
	Category            string     `json:"category"`
	BinaryLike          bool       `json:"binary_like"`
	HasOtherOutcome     bool       `json:"has_other_outcome"`
	UnsupportedShape    bool       `json:"unsupported_shape"`
	AmbiguityNotes      []string   `json:"ambiguity_notes,omitempty"`
	ResolutionSource    string     `json:"resolution_source"`
	DeadlineUTC         *time.Time `json:"deadline_utc,omitempty"`
	DeadlineProvenance  string     `json:"deadline_provenance"`
	CanonicalEventKey   string     `json:"canonical_event_key"`
	CanonicalPropHint   string     `json:"canonical_prop_hint"`
	NormalizedYesTarget string     `json:"normalized_yes_target"`
	Quote               QuoteView  `json:"quote"`
	RawRef              string     `json:"raw_ref"`
}

type EventCluster struct {
	ClusterID       string                `json:"cluster_id"`
	CanonicalKey    string                `json:"canonical_key"`
	Title           string                `json:"title"`
	Confidence      float64               `json:"confidence"`
	AmbiguityNotes  []string              `json:"ambiguity_notes,omitempty"`
	MarketInstances []VenueMarketInstance `json:"market_instances"`
}

type PropositionCluster struct {
	ClusterID       string                `json:"cluster_id"`
	EventClusterID  string                `json:"event_cluster_id"`
	CanonicalKey    string                `json:"canonical_key"`
	Proposition     string                `json:"proposition"`
	Confidence      float64               `json:"confidence"`
	Routeability    Routeability          `json:"routeability"`
	RefusalReasons  []string              `json:"refusal_reasons,omitempty"`
	AmbiguityNotes  []string              `json:"ambiguity_notes,omitempty"`
	MarketInstances []VenueMarketInstance `json:"market_instances"`
}

type EquivalenceAssessment struct {
	AssessmentID         string   `json:"assessment_id"`
	EventClusterID       string   `json:"event_cluster_id"`
	PropositionClusterID string   `json:"proposition_cluster_id,omitempty"`
	CandidateInstanceIDs []string `json:"candidate_instance_ids"`
	Classification       string   `json:"classification"`
	Confidence           float64  `json:"confidence"`
	Reasons              []string `json:"reasons"`
}

type HypotheticalOrder struct {
	OrderID              string  `json:"order_id"`
	PropositionClusterID string  `json:"proposition_cluster_id"`
	Side                 string  `json:"side"`
	LimitProbability     float64 `json:"limit_probability"`
	SizeNotional         float64 `json:"size_notional"`
}

type RoutingDecision struct {
	DecisionID         string            `json:"decision_id"`
	Order              HypotheticalOrder `json:"order"`
	SelectedInstanceID string            `json:"selected_instance_id,omitempty"`
	SelectedVenue      Venue             `json:"selected_venue,omitempty"`
	Action             string            `json:"action"`
	Reasons            []string          `json:"reasons"`
	RankedCandidates   []string          `json:"ranked_candidates,omitempty"`
}
