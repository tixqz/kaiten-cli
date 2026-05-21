package db

import (
	"database/sql"
	"fmt"
)

// Tag represents a Kaiten tag as stored in the local database.
type Tag struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Color   int    `json:"color"`
	RawJSON string `json:"-" yaml:"-"`
}

const tagColumns = `id, name, color, raw_json`

func tagScanners(t *Tag) []any {
	return []any{&t.ID, &t.Name, &t.Color, &t.RawJSON}
}

const tagUpsertSQL = `INSERT INTO tags (id, name, color, raw_json)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	name     = excluded.name,
	color    = excluded.color,
	raw_json = excluded.raw_json`

// UpsertTag inserts or updates a tag.
func (db *DB) UpsertTag(tag Tag) error {
	_, err := db.SQL.Exec(tagUpsertSQL, tag.ID, tag.Name, tag.Color, tag.RawJSON)
	return err
}

// GetTag retrieves a tag by ID. Returns nil if not found.
func (db *DB) GetTag(id int) (*Tag, error) {
	row := db.SQL.QueryRow("SELECT "+tagColumns+" FROM tags WHERE id = ?", id)
	var t Tag
	err := row.Scan(tagScanners(&t)...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// SetCardTags atomically replaces the tag list for a card.
func (db *DB) SetCardTags(cardID int, tagIDs []int) error {
	tx, err := db.SQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM card_tags WHERE card_id = ?", cardID); err != nil {
		return fmt.Errorf("delete card tags: %w", err)
	}

	for _, tid := range tagIDs {
		if _, err := tx.Exec("INSERT INTO card_tags (card_id, tag_id) VALUES (?, ?)", cardID, tid); err != nil {
			return fmt.Errorf("insert card tag: %w", err)
		}
	}

	return tx.Commit()
}

// GetCardTags returns the tag IDs of all tags on a card.
func (db *DB) GetCardTags(cardID int) ([]int, error) {
	rows, err := db.SQL.Query("SELECT tag_id FROM card_tags WHERE card_id = ? ORDER BY tag_id", cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
