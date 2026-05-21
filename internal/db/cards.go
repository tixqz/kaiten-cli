package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Card represents a Kaiten card as stored in the local database.
type Card struct {
	ID              int     `json:"id"`
	Title           string  `json:"title"`
	Description     string  `json:"description,omitempty"`
	BoardID         int     `json:"board_id"`
	SpaceID         int     `json:"space_id"`
	ColumnID        int     `json:"column_id"`
	LaneID          int     `json:"lane_id,omitempty"`
	TypeID          *int    `json:"type_id,omitempty"`
	OwnerID         int     `json:"owner_id,omitempty"`
	UpdaterID       int     `json:"updater_id,omitempty"`
	Condition       int     `json:"condition,omitempty"`
	SizeText        string  `json:"size_text,omitempty"`
	DueDate         string  `json:"due_date,omitempty"`
	Blocked         bool    `json:"blocked"`
	BlockReason     string  `json:"block_reason,omitempty"`
	Created         string  `json:"created"`
	Updated         string  `json:"updated"`
	CompletedAt     string  `json:"completed_at,omitempty"`
	ColumnChangedAt string  `json:"column_changed_at,omitempty"`
	CommentsTotal   int     `json:"comments_total,omitempty"`
	SprintID        int     `json:"sprint_id,omitempty"`
	SortOrder       float64 `json:"sort_order,omitempty"`
	RawJSON         string  `json:"-" yaml:"-"`
}

const searchableTimeLayout = "2006-01-02T15:04:05.000000000Z"

// SearchQuery defines filters and pagination for searching cards.
type SearchQuery struct {
	Text          string
	BoardID       *int
	SpaceID       *int
	OwnerID       *int
	MemberID      *int
	TagID         *int
	Condition     *int
	CreatedFrom   *string
	CreatedTo     *string
	UpdatedFrom   *string
	UpdatedTo     *string
	CompletedFrom *string
	CompletedTo   *string
	Limit         int
	Offset        int
	Sort          string
}

const cardColumns = `id, title, description, board_id, space_id, column_id, lane_id, type_id,
	owner_id, updater_id, condition, size_text, due_date, blocked, block_reason,
	created, updated, completed_at, column_changed_at, comments_total, sprint_id, sort_order, raw_json`

func cardScanners(c *Card) []any {
	return []any{
		&c.ID, &c.Title, &c.Description, &c.BoardID, &c.SpaceID, &c.ColumnID, &c.LaneID, &c.TypeID,
		&c.OwnerID, &c.UpdaterID, &c.Condition, &c.SizeText, &c.DueDate, &c.Blocked, &c.BlockReason,
		&c.Created, &c.Updated, &c.CompletedAt, &c.ColumnChangedAt, &c.CommentsTotal, &c.SprintID, &c.SortOrder, &c.RawJSON,
	}
}

const cardUpsertSQL = `INSERT INTO cards (
	id, title, description, board_id, space_id, column_id, lane_id, type_id,
	owner_id, updater_id, condition, size_text, due_date, blocked, block_reason,
	created, updated, completed_at, column_changed_at, comments_total, sprint_id, sort_order, raw_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	title          = excluded.title,
	description    = excluded.description,
	board_id       = excluded.board_id,
	space_id       = excluded.space_id,
	column_id      = excluded.column_id,
	lane_id        = excluded.lane_id,
	type_id        = excluded.type_id,
	owner_id       = excluded.owner_id,
	updater_id     = excluded.updater_id,
	condition      = excluded.condition,
	size_text      = excluded.size_text,
	due_date       = excluded.due_date,
	blocked        = excluded.blocked,
	block_reason   = excluded.block_reason,
	created        = excluded.created,
	updated        = excluded.updated,
	completed_at   = excluded.completed_at,
	column_changed_at = excluded.column_changed_at,
	comments_total = excluded.comments_total,
	sprint_id      = excluded.sprint_id,
	sort_order     = excluded.sort_order,
	raw_json       = excluded.raw_json`

// UpsertCard inserts or updates a single card.
func (db *DB) UpsertCard(card Card) error {
	card = normalizeCardTimes(card)
	_, err := db.SQL.Exec(cardUpsertSQL,
		card.ID, card.Title, card.Description, card.BoardID, card.SpaceID,
		card.ColumnID, card.LaneID, card.TypeID, card.OwnerID, card.UpdaterID,
		card.Condition, card.SizeText, card.DueDate, boolToInt(card.Blocked),
		card.BlockReason, card.Created, card.Updated, card.CompletedAt,
		card.ColumnChangedAt, card.CommentsTotal, card.SprintID, card.SortOrder, card.RawJSON,
	)
	return err
}

// BulkUpsertCards inserts or updates multiple cards in a single transaction.
func (db *DB) BulkUpsertCards(cards []Card) error {
	tx, err := db.SQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(cardUpsertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, card := range cards {
		card = normalizeCardTimes(card)
		_, err := stmt.Exec(
			card.ID, card.Title, card.Description, card.BoardID, card.SpaceID,
			card.ColumnID, card.LaneID, card.TypeID, card.OwnerID, card.UpdaterID,
			card.Condition, card.SizeText, card.DueDate, boolToInt(card.Blocked),
			card.BlockReason, card.Created, card.Updated, card.CompletedAt,
			card.ColumnChangedAt, card.CommentsTotal, card.SprintID, card.SortOrder, card.RawJSON,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SearchCards searches cards using the provided query filters.
// When Text is set, an FTS5 match is performed on title/description.
// Filters compose safely using placeholders.
func (db *DB) SearchCards(q SearchQuery) ([]Card, error) {
	var clauses []string
	var args []any
	argN := 0

	nextArg := func(v any) string {
		argN++
		args = append(args, v)
		return fmt.Sprintf("?%d", argN)
	}

	if q.Text != "" {
		clauses = append(clauses, fmt.Sprintf("cards.id IN (SELECT rowid FROM cards_fts WHERE cards_fts MATCH %s)", nextArg(q.Text)))
	}
	if q.BoardID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.board_id = %s", nextArg(*q.BoardID)))
	}
	if q.SpaceID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.space_id = %s", nextArg(*q.SpaceID)))
	}
	if q.OwnerID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.owner_id = %s", nextArg(*q.OwnerID)))
	}
	if q.Condition != nil {
		clauses = append(clauses, fmt.Sprintf("cards.condition = %s", nextArg(*q.Condition)))
	}
	if q.MemberID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.id IN (SELECT card_id FROM card_members WHERE user_id = %s)", nextArg(*q.MemberID)))
	}
	if q.TagID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.id IN (SELECT card_id FROM card_tags WHERE tag_id = %s)", nextArg(*q.TagID)))
	}
	if q.CreatedFrom != nil {
		clauses = append(clauses, fmt.Sprintf("cards.created >= %s", nextArg(normalizeQueryTime(*q.CreatedFrom))))
	}
	if q.CreatedTo != nil {
		clauses = append(clauses, fmt.Sprintf("cards.created <= %s", nextArg(normalizeQueryTime(*q.CreatedTo))))
	}
	if q.UpdatedFrom != nil {
		clauses = append(clauses, fmt.Sprintf("cards.updated >= %s", nextArg(normalizeQueryTime(*q.UpdatedFrom))))
	}
	if q.UpdatedTo != nil {
		clauses = append(clauses, fmt.Sprintf("cards.updated <= %s", nextArg(normalizeQueryTime(*q.UpdatedTo))))
	}
	if q.CompletedFrom != nil || q.CompletedTo != nil {
		clauses = append(clauses, "cards.completed_at != ''")
	}
	if q.CompletedFrom != nil {
		clauses = append(clauses, fmt.Sprintf("cards.completed_at >= %s", nextArg(normalizeQueryTime(*q.CompletedFrom))))
	}
	if q.CompletedTo != nil {
		clauses = append(clauses, fmt.Sprintf("cards.completed_at <= %s", nextArg(normalizeQueryTime(*q.CompletedTo))))
	}

	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}

	sortClause := buildSort(q.Sort)

	var limitClause string
	if q.Limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", q.Limit)
		if q.Offset > 0 {
			limitClause += fmt.Sprintf(" OFFSET %d", q.Offset)
		}
	}

	query := fmt.Sprintf("SELECT %s FROM cards%s%s%s", cardColumns, where, sortClause, limitClause)

	rows, err := db.SQL.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search cards: %w", err)
	}
	defer rows.Close()

	cards := make([]Card, 0)
	for rows.Next() {
		var c Card
		if err := rows.Scan(cardScanners(&c)...); err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		cards = append(cards, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return cards, nil
}

// GetCard retrieves a single card by ID. Returns nil if not found.
func (db *DB) GetCard(id int) (*Card, error) {
	query := fmt.Sprintf("SELECT %s FROM cards WHERE id = ?", cardColumns)
	row := db.SQL.QueryRow(query, id)
	var c Card
	err := row.Scan(cardScanners(&c)...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CountCards returns the total number of cards matching the query filters.
func (db *DB) CountCards(q SearchQuery) (int, error) {
	// Build same WHERE clause as SearchCards but without sort/limit/offset.
	var clauses []string
	var args []any
	argN := 0

	nextArg := func(v any) string {
		argN++
		args = append(args, v)
		return fmt.Sprintf("?%d", argN)
	}

	if q.Text != "" {
		clauses = append(clauses, fmt.Sprintf("cards.id IN (SELECT rowid FROM cards_fts WHERE cards_fts MATCH %s)", nextArg(q.Text)))
	}
	if q.BoardID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.board_id = %s", nextArg(*q.BoardID)))
	}
	if q.SpaceID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.space_id = %s", nextArg(*q.SpaceID)))
	}
	if q.OwnerID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.owner_id = %s", nextArg(*q.OwnerID)))
	}
	if q.Condition != nil {
		clauses = append(clauses, fmt.Sprintf("cards.condition = %s", nextArg(*q.Condition)))
	}
	if q.MemberID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.id IN (SELECT card_id FROM card_members WHERE user_id = %s)", nextArg(*q.MemberID)))
	}
	if q.TagID != nil {
		clauses = append(clauses, fmt.Sprintf("cards.id IN (SELECT card_id FROM card_tags WHERE tag_id = %s)", nextArg(*q.TagID)))
	}
	if q.CreatedFrom != nil {
		clauses = append(clauses, fmt.Sprintf("cards.created >= %s", nextArg(normalizeQueryTime(*q.CreatedFrom))))
	}
	if q.CreatedTo != nil {
		clauses = append(clauses, fmt.Sprintf("cards.created <= %s", nextArg(normalizeQueryTime(*q.CreatedTo))))
	}
	if q.UpdatedFrom != nil {
		clauses = append(clauses, fmt.Sprintf("cards.updated >= %s", nextArg(normalizeQueryTime(*q.UpdatedFrom))))
	}
	if q.UpdatedTo != nil {
		clauses = append(clauses, fmt.Sprintf("cards.updated <= %s", nextArg(normalizeQueryTime(*q.UpdatedTo))))
	}
	if q.CompletedFrom != nil || q.CompletedTo != nil {
		clauses = append(clauses, "cards.completed_at != ''")
	}
	if q.CompletedFrom != nil {
		clauses = append(clauses, fmt.Sprintf("cards.completed_at >= %s", nextArg(normalizeQueryTime(*q.CompletedFrom))))
	}
	if q.CompletedTo != nil {
		clauses = append(clauses, fmt.Sprintf("cards.completed_at <= %s", nextArg(normalizeQueryTime(*q.CompletedTo))))
	}

	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}

	query := "SELECT COUNT(*) FROM cards" + where
	var count int
	if err := db.SQL.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count cards: %w", err)
	}
	return count, nil
}

// buildSort converts a sort key to an ORDER BY clause.
func buildSort(sort string) string {
	switch sort {
	case "created":
		return " ORDER BY cards.created ASC, cards.id ASC"
	case "created_desc":
		return " ORDER BY cards.created DESC, cards.id DESC"
	case "updated":
		return " ORDER BY cards.updated ASC, cards.id ASC"
	case "updated_desc":
		return " ORDER BY cards.updated DESC, cards.id DESC"
	case "completed":
		return " ORDER BY cards.completed_at ASC, cards.id ASC"
	case "completed_desc":
		return " ORDER BY cards.completed_at DESC, cards.id DESC"
	case "id":
		return " ORDER BY cards.id ASC"
	case "id_desc":
		return " ORDER BY cards.id DESC"
	default:
		return " ORDER BY cards.updated DESC, cards.id DESC"
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func normalizeCardTimes(card Card) Card {
	card.Created = normalizeQueryTime(card.Created)
	card.Updated = normalizeQueryTime(card.Updated)
	card.CompletedAt = normalizeQueryTime(card.CompletedAt)
	card.ColumnChangedAt = normalizeQueryTime(card.ColumnChangedAt)
	return card
}

func normalizeQueryTime(value string) string {
	if value == "" {
		return value
	}
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t.UTC().Format(searchableTimeLayout)
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.UTC().Format(searchableTimeLayout)
	}
	return value
}
