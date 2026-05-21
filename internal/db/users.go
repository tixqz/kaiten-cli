package db

import (
	"database/sql"
	"fmt"
)

// User represents a Kaiten user as stored in the local database.
type User struct {
	ID        int    `json:"id"`
	FullName  string `json:"full_name"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
	RawJSON   string `json:"-" yaml:"-"`
}

const userColumns = `id, full_name, name, email, avatar_url, raw_json`

func userScanners(u *User) []any {
	return []any{&u.ID, &u.FullName, &u.Name, &u.Email, &u.AvatarURL, &u.RawJSON}
}

const userUpsertSQL = `INSERT INTO users (id, full_name, name, email, avatar_url, raw_json)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	full_name  = excluded.full_name,
	name       = excluded.name,
	email      = excluded.email,
	avatar_url = excluded.avatar_url,
	raw_json   = excluded.raw_json`

// UpsertUser inserts or updates a user.
func (db *DB) UpsertUser(user User) error {
	_, err := db.SQL.Exec(userUpsertSQL,
		user.ID, user.FullName, user.Name, user.Email, user.AvatarURL, user.RawJSON)
	return err
}

// GetUser retrieves a user by ID. Returns nil if not found.
func (db *DB) GetUser(id int) (*User, error) {
	query := "SELECT " + userColumns + " FROM users WHERE id = ?"
	row := db.SQL.QueryRow(query, id)
	var u User
	err := row.Scan(userScanners(&u)...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// SetCardMembers atomically replaces the member list for a card.
func (db *DB) SetCardMembers(cardID int, userIDs []int) error {
	tx, err := db.SQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM card_members WHERE card_id = ?", cardID); err != nil {
		return fmt.Errorf("delete card members: %w", err)
	}

	for _, uid := range userIDs {
		if _, err := tx.Exec("INSERT INTO card_members (card_id, user_id) VALUES (?, ?)", cardID, uid); err != nil {
			return fmt.Errorf("insert card member: %w", err)
		}
	}

	return tx.Commit()
}

// GetCardMembers returns the user IDs of all members on a card.
func (db *DB) GetCardMembers(cardID int) ([]int, error) {
	rows, err := db.SQL.Query("SELECT user_id FROM card_members WHERE card_id = ? ORDER BY user_id", cardID)
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
