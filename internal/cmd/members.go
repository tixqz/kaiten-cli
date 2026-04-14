package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var membersCmd = &cobra.Command{
	Use:   "members",
	Short: "Manage card members",
}

var membersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List members of a card",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		members, err := apiClient.ListCardMembers(cardID)
		if err != nil {
			return err
		}
		outputJSON(members)
		return nil
	},
}

func init() {
	membersCmd.AddCommand(membersListCmd)
	membersListCmd.Flags().Int("card-id", 0, "Card ID (required)")
	membersListCmd.MarkFlagRequired("card-id")
}
