package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/api"
)

var commentsCmd = &cobra.Command{
	Use:   "comments",
	Short: "Manage card comments",
}

var commentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List comments on a card",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		comments, err := apiClient.ListComments(cardID)
		if err != nil {
			return err
		}
		outputJSON(comments)
		return nil
	},
}

var commentsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a comment to a card",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		text, err := cmd.Flags().GetString("text")
		if err != nil {
			return err
		}
		if text == "" {
			return fmt.Errorf("--text is required")
		}
		commentType, err := cmd.Flags().GetInt("type")
		if err != nil {
			return err
		}
		req := &api.CreateCommentRequest{
			Text: text,
			Type: commentType,
		}
		comment, err := apiClient.AddComment(cardID, req)
		if err != nil {
			return err
		}
		outputJSON(comment)
		return nil
	},
}

func init() {
	commentsCmd.AddCommand(commentsListCmd)
	commentsCmd.AddCommand(commentsAddCmd)

	commentsListCmd.Flags().Int("card-id", 0, "Card ID (required)")
	commentsListCmd.MarkFlagRequired("card-id")

	commentsAddCmd.Flags().Int("card-id", 0, "Card ID (required)")
	commentsAddCmd.Flags().String("text", "", "Comment text (required)")
	commentsAddCmd.Flags().Int("type", 1, "Comment type: 1=markdown, 2=html")
	commentsAddCmd.MarkFlagRequired("card-id")
	commentsAddCmd.MarkFlagRequired("text")
}
