package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var boardsCmd = &cobra.Command{
	Use:   "boards",
	Short: "Manage boards",
}

var boardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List boards in a space",
	RunE: func(cmd *cobra.Command, args []string) error {
		spaceID, err := cmd.Flags().GetInt("space-id")
		if err != nil {
			return err
		}
		if spaceID == 0 {
			return fmt.Errorf("--space-id is required")
		}
		boards, err := apiClient.ListBoards(spaceID)
		if err != nil {
			return err
		}
		outputJSON(boards)
		return nil
	},
}

var boardsGetCmd = &cobra.Command{
	Use:   "get [board_id]",
	Short: "Get a board by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid board ID: %s", args[0])
		}
		board, err := apiClient.GetBoard(id)
		if err != nil {
			return err
		}
		outputJSON(board)
		return nil
	},
}

func init() {
	boardsCmd.AddCommand(boardsListCmd)
	boardsCmd.AddCommand(boardsGetCmd)

	boardsListCmd.Flags().Int("space-id", 0, "Space ID (required)")
	boardsListCmd.MarkFlagRequired("space-id")
}
