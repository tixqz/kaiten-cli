package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/api"
)

var blockersCmd = &cobra.Command{
	Use:   "blockers",
	Short: "Manage card blockers",
}

var blockersBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block a card",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		reason, err := cmd.Flags().GetString("reason")
		if err != nil {
			return err
		}
		req := &api.BlockCardRequest{Reason: reason}
		blocker, err := apiClient.BlockCard(cardID, req)
		if err != nil {
			return err
		}
		outputJSON(blocker)
		return nil
	},
}

var blockersUnblockCmd = &cobra.Command{
	Use:   "unblock",
	Short: "Unblock a card (removes the blocker)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		blockerID, err := cmd.Flags().GetInt("blocker-id")
		if err != nil {
			return err
		}
		if blockerID == 0 {
			return fmt.Errorf("--blocker-id is required")
		}
		if err := apiClient.DeleteBlocker(cardID, blockerID); err != nil {
			return err
		}
		outputJSON(map[string]int{"unblocked": blockerID})
		return nil
	},
}

var blockersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a blocker",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		blockerID, err := cmd.Flags().GetInt("blocker-id")
		if err != nil {
			return err
		}
		if blockerID == 0 {
			return fmt.Errorf("--blocker-id is required")
		}
		if err := apiClient.DeleteBlocker(cardID, blockerID); err != nil {
			return err
		}
		outputJSON(map[string]int{"deleted": blockerID})
		return nil
	},
}

func init() {
	blockersCmd.AddCommand(blockersBlockCmd)
	blockersCmd.AddCommand(blockersUnblockCmd)
	blockersCmd.AddCommand(blockersDeleteCmd)

	blockersBlockCmd.Flags().Int("card-id", 0, "Card ID (required)")
	blockersBlockCmd.Flags().String("reason", "", "Reason for blocking")
	blockersBlockCmd.MarkFlagRequired("card-id")

	blockersUnblockCmd.Flags().Int("card-id", 0, "Card ID (required)")
	blockersUnblockCmd.Flags().Int("blocker-id", 0, "Blocker ID (required)")
	blockersUnblockCmd.MarkFlagRequired("card-id")
	blockersUnblockCmd.MarkFlagRequired("blocker-id")

	blockersDeleteCmd.Flags().Int("card-id", 0, "Card ID (required)")
	blockersDeleteCmd.Flags().Int("blocker-id", 0, "Blocker ID (required)")
	blockersDeleteCmd.MarkFlagRequired("card-id")
	blockersDeleteCmd.MarkFlagRequired("blocker-id")
}
