package db

import (
	"database/sql"
	"fmt"
)

// SyncState tracks the synchronization status for a scope (e.g., "all" or a specific board).
type SyncState struct {
	ScopeType   string `json:"scope_type"`
	ScopeID     int    `json:"scope_id"`
	StartedAt   string `json:"started_at"`
	FinishedAt  string `json:"finished_at"`
	Status      string `json:"status"`
	CardsSynced int    `json:"cards_synced"`
	Error       string `json:"error,omitempty"`
}

const syncColumns = `scope_type, scope_id, started_at, finished_at, status, cards_synced, error`

func syncScanners(s *SyncState) []any {
	return []any{&s.ScopeType, &s.ScopeID, &s.StartedAt, &s.FinishedAt, &s.Status, &s.CardsSynced, &s.Error}
}

const syncUpsertSQL = `INSERT INTO sync_state (scope_type, scope_id, started_at, finished_at, status, cards_synced, error)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(scope_type, scope_id) DO UPDATE SET
	started_at   = excluded.started_at,
	finished_at  = excluded.finished_at,
	status       = excluded.status,
	cards_synced = excluded.cards_synced,
	error        = excluded.error`

// SetSyncState inserts or updates the sync state for a scope.
func (db *DB) SetSyncState(state SyncState) error {
	_, err := db.SQL.Exec(syncUpsertSQL,
		state.ScopeType, state.ScopeID, state.StartedAt, state.FinishedAt,
		state.Status, state.CardsSynced, state.Error)
	return err
}

// GetSyncState retrieves the sync state for a scope. Returns nil if not found.
func (db *DB) GetSyncState(scopeType string, scopeID int) (*SyncState, error) {
	row := db.SQL.QueryRow("SELECT "+syncColumns+" FROM sync_state WHERE scope_type = ? AND scope_id = ?",
		scopeType, scopeID)
	var s SyncState
	err := row.Scan(syncScanners(&s)...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// StatusString returns a human-readable summary of sync states.
func (db *DB) StatusString() (string, error) {
	rows, err := db.SQL.Query("SELECT " + syncColumns + " FROM sync_state ORDER BY scope_type, scope_id")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var states []SyncState
	for rows.Next() {
		var s SyncState
		if err := rows.Scan(syncScanners(&s)...); err != nil {
			return "", err
		}
		states = append(states, s)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	if len(states) == 0 {
		return "no sync history", nil
	}

	var result string
	for _, s := range states {
		result += fmt.Sprintf("  %s/%d: %s (synced %d cards)\n", s.ScopeType, s.ScopeID, s.Status, s.CardsSynced)
	}
	return result, nil
}

// ResetSync clears all sync state records.
func (db *DB) ResetSync() error {
	_, err := db.SQL.Exec("DELETE FROM sync_state")
	return err
}

// Vacuum reclaims unused space in the database.
func (db *DB) Vacuum() error {
	_, err := db.SQL.Exec("VACUUM")
	return err
}
