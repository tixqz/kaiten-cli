package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var lanesCmd = &cobra.Command{
	Use:   "lanes",
	Short: "Manage lanes",
}

var lanesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List lanes on a board",
	RunE: func(cmd *cobra.Command, args []string) error {
		boardID, err := cmd.Flags().GetInt("board-id")
		if err != nil {
			return err
		}
		if boardID == 0 {
			return fmt.Errorf("--board-id is required")
		}
		lanes, err := apiClient.ListLanes(boardID)
		if err != nil {
			return err
		}
		outputJSON(lanes)
		return nil
	},
}

func init() {
	lanesCmd.AddCommand(lanesListCmd)
	lanesListCmd.Flags().Int("board-id", 0, "Board ID (required)")
	lanesListCmd.MarkFlagRequired("board-id")
}
