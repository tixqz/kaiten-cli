package db

// Board represents a Kaiten board as stored in the local database.
type Board struct {
	ID        int    `json:"id"`
	SpaceID   int    `json:"space_id"`
	Title     string `json:"title"`
	Name      string `json:"name,omitempty"`
	RawJSON   string `json:"-" yaml:"-"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

const boardColumns = `id, space_id, title, name, raw_json, updated_at`

func boardScanners(b *Board) []any {
	return []any{&b.ID, &b.SpaceID, &b.Title, &b.Name, &b.RawJSON, &b.UpdatedAt}
}

const boardUpsertSQL = `INSERT INTO boards (id, space_id, title, name, raw_json, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	space_id   = excluded.space_id,
	title      = excluded.title,
	name       = excluded.name,
	raw_json   = excluded.raw_json,
	updated_at = excluded.updated_at`

// UpsertBoard inserts or updates a board.
func (db *DB) UpsertBoard(board Board) error {
	_, err := db.SQL.Exec(boardUpsertSQL,
		board.ID, board.SpaceID, board.Title, board.Name, board.RawJSON, board.UpdatedAt)
	return err
}

// GetBoard retrieves a board by ID. Returns nil if not found.
func (db *DB) GetBoard(id int) (*Board, error) {
	row := db.SQL.QueryRow("SELECT "+boardColumns+" FROM boards WHERE id = ?", id)
	var b Board
	err := row.Scan(boardScanners(&b)...)
	if err != nil {
		return nil, nil
	}
	return &b, nil
}

// ListBoardsBySpace returns all boards for a given space, ordered by id.
func (db *DB) ListBoardsBySpace(spaceID int) ([]Board, error) {
	rows, err := db.SQL.Query("SELECT "+boardColumns+" FROM boards WHERE space_id = ? ORDER BY id", spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []Board
	for rows.Next() {
		var b Board
		if err := rows.Scan(boardScanners(&b)...); err != nil {
			return nil, err
		}
		boards = append(boards, b)
	}
	return boards, rows.Err()
}
