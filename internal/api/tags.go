package api

import (
	"encoding/json"
	"fmt"
)

// Tag represents a Kaiten tag.
type Tag struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Color     int    `json:"color"`
	CompanyID int    `json:"company_id,omitempty"`
	Archived  bool   `json:"archived,omitempty"`
	CreatedAt string `json:"created,omitempty"`
	UpdatedAt string `json:"updated,omitempty"`
}

// CardTag represents a tag attached to a card.
type CardTag struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Color  int    `json:"color"`
	CardID int    `json:"card_id"`
	TagID  int    `json:"tag_id"`
}

// AddTagRequest is the payload for adding a tag to a card.
type AddTagRequest struct {
	Name string `json:"name"`
}

// ListTags returns all company tags.
// GET /tags
func (c *Client) ListTags() ([]Tag, error) {
	data, err := c.Get("/tags")
	if err != nil {
		return nil, err
	}
	var tags []Tag
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("parsing tags: %w", err)
	}
	return tags, nil
}

// ListCardTags returns tags on a card.
// GET /cards/{card_id}/tags
func (c *Client) ListCardTags(cardID int) ([]CardTag, error) {
	data, err := c.Get(fmt.Sprintf("/cards/%d/tags", cardID))
	if err != nil {
		return nil, err
	}
	var tags []CardTag
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("parsing card tags: %w", err)
	}
	return tags, nil
}

// AddCardTag adds a tag to a card.
// POST /cards/{card_id}/tags
func (c *Client) AddCardTag(cardID int, req *AddTagRequest) (*Tag, error) {
	data, err := c.Post(fmt.Sprintf("/cards/%d/tags", cardID), req)
	if err != nil {
		return nil, err
	}
	var tag Tag
	if err := json.Unmarshal(data, &tag); err != nil {
		return nil, fmt.Errorf("parsing tag: %w", err)
	}
	return &tag, nil
}

// RemoveCardTag removes a tag from a card.
// DELETE /cards/{card_id}/tags/{tag_id}
func (c *Client) RemoveCardTag(cardID, tagID int) error {
	_, err := c.Delete(fmt.Sprintf("/cards/%d/tags/%d", cardID, tagID))
	return err
}
