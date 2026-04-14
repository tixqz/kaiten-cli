package cmd

import (
	"fmt"
	"strings"
)

// resolveColumnID resolves column name to ID for a given board.
func resolveColumnID(boardID, columnID int, columnName string) (int, error) {
	if columnID != 0 {
		return columnID, nil
	}
	if columnName == "" {
		return 0, fmt.Errorf("either --column-id or --column-name is required")
	}
	columns, err := apiClient.ListColumns(boardID)
	if err != nil {
		return 0, fmt.Errorf("fetching columns: %w", err)
	}
	for _, col := range columns {
		if strings.EqualFold(col.Title, columnName) {
			return col.ID, nil
		}
	}
	return 0, fmt.Errorf("column %q not found on board %d", columnName, boardID)
}

// resolveLaneID resolves lane name to ID for a given board.
func resolveLaneID(boardID, laneID int, laneName string) (int, error) {
	if laneID != 0 {
		return laneID, nil
	}
	if laneName == "" {
		return 0, nil // lane is optional
	}
	lanes, err := apiClient.ListLanes(boardID)
	if err != nil {
		return 0, fmt.Errorf("fetching lanes: %w", err)
	}
	for _, lane := range lanes {
		if strings.EqualFold(lane.Title, laneName) {
			return lane.ID, nil
		}
	}
	return 0, fmt.Errorf("lane %q not found on board %d", laneName, boardID)
}
