package db

// Space represents a Kaiten space as stored in the local database.
type Space struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Name      string `json:"name,omitempty"`
	RawJSON   string `json:"-" yaml:"-"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

const spaceColumns = `id, title, name, raw_json, updated_at`

func spaceScanners(s *Space) []any {
	return []any{&s.ID, &s.Title, &s.Name, &s.RawJSON, &s.UpdatedAt}
}

const spaceUpsertSQL = `INSERT INTO spaces (id, title, name, raw_json, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	title      = excluded.title,
	name       = excluded.name,
	raw_json   = excluded.raw_json,
	updated_at = excluded.updated_at`

// UpsertSpace inserts or updates a space.
func (db *DB) UpsertSpace(space Space) error {
	_, err := db.SQL.Exec(spaceUpsertSQL,
		space.ID, space.Title, space.Name, space.RawJSON, space.UpdatedAt)
	return err
}

// GetSpace retrieves a space by ID. Returns nil if not found.
func (db *DB) GetSpace(id int) (*Space, error) {
	row := db.SQL.QueryRow("SELECT "+spaceColumns+" FROM spaces WHERE id = ?", id)
	var s Space
	err := row.Scan(spaceScanners(&s)...)
	if err != nil {
		return nil, nil
	}
	return &s, nil
}

// ListSpaces returns all spaces, ordered by id.
func (db *DB) ListSpaces() ([]Space, error) {
	rows, err := db.SQL.Query("SELECT " + spaceColumns + " FROM spaces ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spaces []Space
	for rows.Next() {
		var s Space
		if err := rows.Scan(spaceScanners(&s)...); err != nil {
			return nil, err
		}
		spaces = append(spaces, s)
	}
	return spaces, rows.Err()
}
