package api

import (
	"encoding/json"
	"fmt"
)

// ChecklistItem represents an item in a checklist.
type ChecklistItem struct {
	ID            int     `json:"id"`
	Text          string  `json:"text"`
	Checked       bool    `json:"checked"`
	CheckerID     *int    `json:"checker_id,omitempty"`
	CheckedAt     string  `json:"checked_at,omitempty"`
	SortOrder     float64 `json:"sort_order"`
	ResponsibleID *int    `json:"responsible_id,omitempty"`
	DueDate       string  `json:"due_date,omitempty"`
	Deleted       bool    `json:"deleted,omitempty"`
	CreatedAt     string  `json:"created"`
	UpdatedAt     string  `json:"updated"`
}

// Checklist represents a card checklist.
type Checklist struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	PolicyID  *int            `json:"policy_id,omitempty"`
	SortOrder float64         `json:"sort_order,omitempty"`
	Deleted   bool            `json:"deleted,omitempty"`
	Items     []ChecklistItem `json:"items,omitempty"`
	CreatedAt string          `json:"created"`
	UpdatedAt string          `json:"updated"`
}

// CreateChecklistRequest is the payload for creating a checklist.
type CreateChecklistRequest struct {
	Name string `json:"name,omitempty"`
}

// AddChecklistItemRequest is the payload for adding an item to a checklist.
type AddChecklistItemRequest struct {
	Text string `json:"text"`
}

// UpdateChecklistItemRequest is the payload for updating a checklist item.
type UpdateChecklistItemRequest struct {
	Text    *string `json:"text,omitempty"`
	Checked *bool   `json:"checked,omitempty"`
}

// CreateChecklist creates a new checklist on a card.
// POST /cards/{card_id}/checklists
func (c *Client) CreateChecklist(cardID int, req *CreateChecklistRequest) (*Checklist, error) {
	data, err := c.Post(fmt.Sprintf("/cards/%d/checklists", cardID), req)
	if err != nil {
		return nil, err
	}
	var checklist Checklist
	if err := json.Unmarshal(data, &checklist); err != nil {
		return nil, fmt.Errorf("parsing checklist: %w", err)
	}
	return &checklist, nil
}

// GetChecklist retrieves a checklist with its items.
// GET /cards/{card_id}/checklists/{id}
func (c *Client) GetChecklist(cardID, checklistID int) (*Checklist, error) {
	data, err := c.Get(fmt.Sprintf("/cards/%d/checklists/%d", cardID, checklistID))
	if err != nil {
		return nil, err
	}
	var checklist Checklist
	if err := json.Unmarshal(data, &checklist); err != nil {
		return nil, fmt.Errorf("parsing checklist: %w", err)
	}
	return &checklist, nil
}

// DeleteChecklist removes a checklist from a card.
// DELETE /cards/{card_id}/checklists/{id}
func (c *Client) DeleteChecklist(cardID, checklistID int) error {
	_, err := c.Delete(fmt.Sprintf("/cards/%d/checklists/%d", cardID, checklistID))
	return err
}

// AddChecklistItem adds an item to a checklist.
// POST /cards/{card_id}/checklists/{checklist_id}/items
func (c *Client) AddChecklistItem(cardID, checklistID int, req *AddChecklistItemRequest) (*ChecklistItem, error) {
	data, err := c.Post(fmt.Sprintf("/cards/%d/checklists/%d/items", cardID, checklistID), req)
	if err != nil {
		return nil, err
	}
	var item ChecklistItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("parsing checklist item: %w", err)
	}
	return &item, nil
}

// UpdateChecklistItem updates a checklist item (e.g. toggle checked, change text).
// PATCH /cards/{card_id}/checklists/{checklist_id}/items/{id}
func (c *Client) UpdateChecklistItem(cardID, checklistID, itemID int, req *UpdateChecklistItemRequest) (*ChecklistItem, error) {
	data, err := c.Patch(fmt.Sprintf("/cards/%d/checklists/%d/items/%d", cardID, checklistID, itemID), req)
	if err != nil {
		return nil, err
	}
	var item ChecklistItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("parsing checklist item: %w", err)
	}
	return &item, nil
}

// DeleteChecklistItem removes an item from a checklist.
// DELETE /cards/{card_id}/checklists/{checklist_id}/items/{id}
func (c *Client) DeleteChecklistItem(cardID, checklistID, itemID int) error {
	_, err := c.Delete(fmt.Sprintf("/cards/%d/checklists/%d/items/%d", cardID, checklistID, itemID))
	return err
}
