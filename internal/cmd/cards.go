package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/api"
)

var cardsCmd = &cobra.Command{
	Use:   "cards",
	Short: "Manage cards",
}

var cardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cards on a board",
	RunE: func(cmd *cobra.Command, args []string) error {
		boardID, err := cmd.Flags().GetInt("board-id")
		if err != nil {
			return err
		}
		if boardID == 0 {
			return fmt.Errorf("--board-id is required")
		}
		condition := 1 // active by default
		if archived, _ := cmd.Flags().GetBool("archived"); archived {
			condition = 2
		}
		if all, _ := cmd.Flags().GetBool("all"); all {
			condition = 0
		}
		cards, err := apiClient.ListCards(boardID, condition)
		if err != nil {
			return err
		}
		outputJSON(cards)
		return nil
	},
}

var cardsGetCmd = &cobra.Command{
	Use:   "get [card_id]",
	Short: "Get a card by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid card ID: %s", args[0])
		}
		card, err := apiClient.GetCard(id)
		if err != nil {
			return err
		}
		outputJSON(card)
		return nil
	},
}

var cardsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new card",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		boardID, _ := cmd.Flags().GetInt("board-id")
		columnID, _ := cmd.Flags().GetInt("column-id")
		laneID, _ := cmd.Flags().GetInt("lane-id")
		columnName, _ := cmd.Flags().GetString("column-name")
		laneName, _ := cmd.Flags().GetString("lane-name")
		typeID, _ := cmd.Flags().GetInt("type-id")
		position, _ := cmd.Flags().GetInt("position")
		description, _ := cmd.Flags().GetString("description")
		sizeText := ""
		if size, _ := cmd.Flags().GetInt("size"); size > 0 {
			sizeText = strconv.Itoa(size)
		}
		dueDate, _ := cmd.Flags().GetString("due-date")

		resolvedColumnID, err := resolveColumnID(boardID, columnID, columnName)
		if err != nil {
			return err
		}
		resolvedLaneID, err := resolveLaneID(boardID, laneID, laneName)
		if err != nil {
			return err
		}

		req := &api.CreateCardRequest{
			Title:       title,
			BoardID:     boardID,
			ColumnID:    resolvedColumnID,
			LaneID:      resolvedLaneID,
			TypeID:      typeID,
			Position:    position,
			Description: description,
			SizeText:    sizeText,
			DueDate:     dueDate,
		}

		card, err := apiClient.CreateCard(req)
		if err != nil {
			return err
		}
		outputJSON(card)
		return nil
	},
}

var cardsUpdateCmd = &cobra.Command{
	Use:   "update [card_id]",
	Short: "Update an existing card",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid card ID: %s", args[0])
		}

		req := &api.UpdateCardRequest{}
		boardID, _ := cmd.Flags().GetInt("board-id")

		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			req.Title = &v
		}
		if cmd.Flags().Changed("column-id") {
			v, _ := cmd.Flags().GetInt("column-id")
			req.ColumnID = &v
		}
		if cmd.Flags().Changed("column-name") {
			if boardID == 0 {
				return fmt.Errorf("--board-id is required when using --column-name")
			}
			columnName, _ := cmd.Flags().GetString("column-name")
			resolvedColumnID, err := resolveColumnID(boardID, 0, columnName)
			if err != nil {
				return err
			}
			req.ColumnID = &resolvedColumnID
		}
		if cmd.Flags().Changed("lane-id") {
			v, _ := cmd.Flags().GetInt("lane-id")
			req.LaneID = &v
		}
		if cmd.Flags().Changed("lane-name") {
			if boardID == 0 {
				return fmt.Errorf("--board-id is required when using --lane-name")
			}
			laneName, _ := cmd.Flags().GetString("lane-name")
			resolvedLaneID, err := resolveLaneID(boardID, 0, laneName)
			if err != nil {
				return err
			}
			req.LaneID = &resolvedLaneID
		}
		if cmd.Flags().Changed("type-id") {
			v, _ := cmd.Flags().GetInt("type-id")
			req.TypeID = &v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			req.Description = &v
		}
		if cmd.Flags().Changed("size") {
			v, _ := cmd.Flags().GetInt("size")
			s := strconv.Itoa(v)
			req.SizeText = &s
		}
		if cmd.Flags().Changed("due-date") {
			v, _ := cmd.Flags().GetString("due-date")
			req.DueDate = &v
		}

		card, err := apiClient.UpdateCard(id, req)
		if err != nil {
			return err
		}
		outputJSON(card)
		return nil
	},
}

var cardsArchiveCmd = &cobra.Command{
	Use:   "archive [card_id]",
	Short: "Archive a card",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid card ID: %s", args[0])
		}
		condition := 2
		req := &api.UpdateCardRequest{Condition: &condition}
		card, err := apiClient.UpdateCard(id, req)
		if err != nil {
			return err
		}
		outputJSON(card)
		return nil
	},
}

var cardsUnarchiveCmd = &cobra.Command{
	Use:   "unarchive [card_id]",
	Short: "Unarchive a card",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid card ID: %s", args[0])
		}
		condition := 1
		req := &api.UpdateCardRequest{Condition: &condition}
		card, err := apiClient.UpdateCard(id, req)
		if err != nil {
			return err
		}
		outputJSON(card)
		return nil
	},
}

var cardsDeleteCmd = &cobra.Command{
	Use:   "delete [card_id]",
	Short: "Delete a card",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid card ID: %s", args[0])
		}
		if err := apiClient.DeleteCard(id); err != nil {
			return err
		}
		fmt.Printf("{\"deleted\": %d}\n", id)
		return nil
	},
}

func init() {
	cardsCmd.AddCommand(cardsListCmd)
	cardsCmd.AddCommand(cardsGetCmd)
	cardsCmd.AddCommand(cardsCreateCmd)
	cardsCmd.AddCommand(cardsUpdateCmd)
	cardsCmd.AddCommand(cardsDeleteCmd)
	cardsCmd.AddCommand(cardsArchiveCmd)
	cardsCmd.AddCommand(cardsUnarchiveCmd)

	// list flags
	cardsListCmd.Flags().Int("board-id", 0, "Board ID (required)")
	cardsListCmd.Flags().Bool("archived", false, "Show archived cards instead of active")
	cardsListCmd.Flags().Bool("all", false, "Show all cards (active + archived)")
	cardsListCmd.MarkFlagRequired("board-id")

	// create flags
	cardsCreateCmd.Flags().String("title", "", "Card title (required)")
	cardsCreateCmd.Flags().Int("board-id", 0, "Board ID (required)")
	cardsCreateCmd.Flags().Int("column-id", 0, "Column ID")
	cardsCreateCmd.Flags().String("column-name", "", "Column name (alternative to --column-id)")
	cardsCreateCmd.Flags().Int("lane-id", 0, "Lane ID")
	cardsCreateCmd.Flags().String("lane-name", "", "Lane name (alternative to --lane-id)")
	cardsCreateCmd.Flags().Int("type-id", 0, "Card type ID")
	cardsCreateCmd.Flags().Int("position", 0, "Position: 1=top, 2=bottom")
	cardsCreateCmd.Flags().String("description", "", "Card description")
	cardsCreateCmd.Flags().Int("size", 0, "Card size")
	cardsCreateCmd.Flags().String("due-date", "", "Due date (ISO 8601)")
	cardsCreateCmd.MarkFlagRequired("title")
	cardsCreateCmd.MarkFlagRequired("board-id")

	// update flags
	cardsUpdateCmd.Flags().Int("board-id", 0, "Board ID (required for --column-name/--lane-name)")
	cardsUpdateCmd.Flags().String("title", "", "New title")
	cardsUpdateCmd.Flags().Int("column-id", 0, "New column ID")
	cardsUpdateCmd.Flags().String("column-name", "", "New column name (alternative to --column-id)")
	cardsUpdateCmd.Flags().Int("lane-id", 0, "New lane ID")
	cardsUpdateCmd.Flags().String("lane-name", "", "New lane name (alternative to --lane-id)")
	cardsUpdateCmd.Flags().Int("type-id", 0, "New card type ID")
	cardsUpdateCmd.Flags().String("description", "", "New description")
	cardsUpdateCmd.Flags().Int("size", 0, "New card size")
	cardsUpdateCmd.Flags().String("due-date", "", "New due date (ISO 8601)")
}
