package artifacts

import (
	"encoding/json"
	"os"
	"path/filepath"

	"equinox/internal/model"
)

type Bundle struct {
	Instances   []model.VenueMarketInstance   `json:"instances"`
	Events      []model.EventCluster          `json:"event_clusters"`
	Props       []model.PropositionCluster    `json:"proposition_clusters"`
	Assessments []model.EquivalenceAssessment `json:"assessments"`
	Decisions   []model.RoutingDecision       `json:"routing_decisions"`
	Evaluation  map[string]string             `json:"evaluation_labels"`
}

func Write(dir string, b Bundle) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return writeJSON(filepath.Join(dir, "bundle.json"), b)
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
