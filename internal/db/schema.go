package db

import "fmt"

const schemaVersion = 1

// migrate creates all tables, indexes, and triggers idempotently.
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS schema_meta (
		version INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS spaces (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL DEFAULT '',
		raw_json TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS boards (
		id INTEGER PRIMARY KEY,
		space_id INTEGER NOT NULL DEFAULT 0,
		title TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL DEFAULT '',
		raw_json TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS cards (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		board_id INTEGER NOT NULL DEFAULT 0,
		space_id INTEGER NOT NULL DEFAULT 0,
		column_id INTEGER NOT NULL DEFAULT 0,
		lane_id INTEGER NOT NULL DEFAULT 0,
		type_id INTEGER,
		owner_id INTEGER NOT NULL DEFAULT 0,
		updater_id INTEGER NOT NULL DEFAULT 0,
		condition INTEGER NOT NULL DEFAULT 0,
		size_text TEXT NOT NULL DEFAULT '',
		due_date TEXT NOT NULL DEFAULT '',
		blocked INTEGER NOT NULL DEFAULT 0,
		block_reason TEXT NOT NULL DEFAULT '',
		created TEXT NOT NULL DEFAULT '',
		updated TEXT NOT NULL DEFAULT '',
		completed_at TEXT NOT NULL DEFAULT '',
		column_changed_at TEXT NOT NULL DEFAULT '',
		comments_total INTEGER NOT NULL DEFAULT 0,
		sprint_id INTEGER NOT NULL DEFAULT 0,
		sort_order REAL NOT NULL DEFAULT 0,
		raw_json TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		full_name TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL DEFAULT '',
		email TEXT NOT NULL DEFAULT '',
		avatar_url TEXT NOT NULL DEFAULT '',
		raw_json TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS card_members (
		card_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		PRIMARY KEY (card_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL DEFAULT '',
		color INTEGER NOT NULL DEFAULT 0,
		raw_json TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS card_tags (
		card_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		PRIMARY KEY (card_id, tag_id)
	);

	CREATE TABLE IF NOT EXISTS sync_state (
		scope_type TEXT NOT NULL,
		scope_id INTEGER NOT NULL,
		started_at TEXT NOT NULL DEFAULT '',
		finished_at TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT '',
		cards_synced INTEGER NOT NULL DEFAULT 0,
		error TEXT NOT NULL DEFAULT '',
		PRIMARY KEY (scope_type, scope_id)
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS cards_fts USING fts5(
		title, description,
		content='cards',
		tokenize='porter unicode61'
	);

	-- FTS triggers to keep cards_fts in sync with cards table
	CREATE TRIGGER IF NOT EXISTS cards_ai AFTER INSERT ON cards BEGIN
		INSERT INTO cards_fts(rowid, title, description) VALUES (new.id, new.title, new.description);
	END;

	CREATE TRIGGER IF NOT EXISTS cards_ad AFTER DELETE ON cards BEGIN
		INSERT INTO cards_fts(cards_fts, rowid, title, description) VALUES('delete', old.id, old.title, old.description);
	END;

	CREATE TRIGGER IF NOT EXISTS cards_au AFTER UPDATE ON cards BEGIN
		INSERT INTO cards_fts(cards_fts, rowid, title, description) VALUES('delete', old.id, old.title, old.description);
		INSERT INTO cards_fts(rowid, title, description) VALUES (new.id, new.title, new.description);
	END;

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_cards_board_id ON cards(board_id);
	CREATE INDEX IF NOT EXISTS idx_cards_space_id ON cards(space_id);
	CREATE INDEX IF NOT EXISTS idx_cards_condition ON cards(condition);
	CREATE INDEX IF NOT EXISTS idx_cards_owner_id ON cards(owner_id);
	CREATE INDEX IF NOT EXISTS idx_cards_created ON cards(created);
	CREATE INDEX IF NOT EXISTS idx_cards_updated ON cards(updated);
	CREATE INDEX IF NOT EXISTS idx_cards_completed_at ON cards(completed_at);
	CREATE INDEX IF NOT EXISTS idx_cards_column_id ON cards(column_id);
	CREATE INDEX IF NOT EXISTS idx_cards_lane_id ON cards(lane_id);
	CREATE INDEX IF NOT EXISTS idx_cards_type_id ON cards(type_id);
	CREATE INDEX IF NOT EXISTS idx_boards_space_id ON boards(space_id);
	CREATE INDEX IF NOT EXISTS idx_card_members_card_id ON card_members(card_id);
	CREATE INDEX IF NOT EXISTS idx_card_members_user_id ON card_members(user_id);
	CREATE INDEX IF NOT EXISTS idx_card_tags_card_id ON card_tags(card_id);
	CREATE INDEX IF NOT EXISTS idx_card_tags_tag_id ON card_tags(tag_id);
	`

	if _, err := db.SQL.Exec(schema); err != nil {
		return fmt.Errorf("creating schema: %w", err)
	}

	// Rebuild FTS index in case cards_fts table was just created with content='cards'
	if _, err := db.SQL.Exec("INSERT INTO cards_fts(cards_fts) VALUES('rebuild')"); err != nil {
		return fmt.Errorf("rebuilding fts: %w", err)
	}

	// Record schema version idempotently
	if _, err := db.SQL.Exec(
		`INSERT INTO schema_meta (version) SELECT ? WHERE NOT EXISTS (SELECT 1 FROM schema_meta)`,
		schemaVersion); err != nil {
		return fmt.Errorf("recording schema version: %w", err)
	}

	return nil
}
