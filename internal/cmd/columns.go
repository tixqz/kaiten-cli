package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var columnsCmd = &cobra.Command{
	Use:   "columns",
	Short: "Manage columns",
}

var columnsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List columns on a board",
	RunE: func(cmd *cobra.Command, args []string) error {
		boardID, err := cmd.Flags().GetInt("board-id")
		if err != nil {
			return err
		}
		if boardID == 0 {
			return fmt.Errorf("--board-id is required")
		}
		columns, err := apiClient.ListColumns(boardID)
		if err != nil {
			return err
		}
		outputJSON(columns)
		return nil
	},
}

func init() {
	columnsCmd.AddCommand(columnsListCmd)
	columnsListCmd.Flags().Int("board-id", 0, "Board ID (required)")
	columnsListCmd.MarkFlagRequired("board-id")
}
