package store

import (
	"context"
	"database/sql"
	"encoding/json"

	"equinox/internal/model"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS event_clusters(id TEXT PRIMARY KEY, payload TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS proposition_clusters(id TEXT PRIMARY KEY, payload TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS assessments(id TEXT PRIMARY KEY, payload TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS routing_decisions(id TEXT PRIMARY KEY, payload TEXT NOT NULL);
`)
	return err
}

func (s *Store) PersistRun(ctx context.Context, events []model.EventCluster, props []model.PropositionCluster, assessments []model.EquivalenceAssessment, decisions []model.RoutingDecision) error {
	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		for _, e := range events {
			if err := upsert(tx, "event_clusters", e.ClusterID, e); err != nil {
				return err
			}
		}
		for _, p := range props {
			if err := upsert(tx, "proposition_clusters", p.ClusterID, p); err != nil {
				return err
			}
		}
		for _, a := range assessments {
			if err := upsert(tx, "assessments", a.AssessmentID, a); err != nil {
				return err
			}
		}
		for _, d := range decisions {
			if err := upsert(tx, "routing_decisions", d.DecisionID, d); err != nil {
				return err
			}
		}
		return nil
	})
}

func upsert(tx *sql.Tx, table, id string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO "+table+"(id,payload) VALUES(?,?) ON CONFLICT(id) DO UPDATE SET payload=excluded.payload", id, string(b))
	return err
}

func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
