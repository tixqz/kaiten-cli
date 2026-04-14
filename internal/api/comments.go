package api

import (
	"encoding/json"
	"fmt"
)

// Comment represents a card comment.
type Comment struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Type      int    `json:"type"`
	CardID    int    `json:"card_id"`
	AuthorID  int    `json:"author_id"`
	CreatedAt string `json:"created"`
	UpdatedAt string `json:"updated"`
}

// CreateCommentRequest is the payload for creating a comment.
type CreateCommentRequest struct {
	Text string `json:"text"`
	Type int    `json:"type,omitempty"`
}

// ListComments returns comments for a card.
// GET /cards/{card_id}/comments
func (c *Client) ListComments(cardID int) ([]Comment, error) {
	data, err := c.Get(fmt.Sprintf("/cards/%d/comments", cardID))
	if err != nil {
		return nil, err
	}
	var comments []Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("parsing comments: %w", err)
	}
	return comments, nil
}

// AddComment creates a new comment for a card.
// POST /cards/{card_id}/comments
func (c *Client) AddComment(cardID int, req *CreateCommentRequest) (*Comment, error) {
	data, err := c.Post(fmt.Sprintf("/cards/%d/comments", cardID), req)
	if err != nil {
		return nil, err
	}
	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("parsing comment: %w", err)
	}
	return &comment, nil
}
