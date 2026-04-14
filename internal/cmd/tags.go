package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/api"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage tags",
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List company tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		tags, err := apiClient.ListTags()
		if err != nil {
			return err
		}
		outputJSON(tags)
		return nil
	},
}

var tagsCardTagsCmd = &cobra.Command{
	Use:   "card-tags",
	Short: "List tags on a card",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		tags, err := apiClient.ListCardTags(cardID)
		if err != nil {
			return err
		}
		outputJSON(tags)
		return nil
	},
}

var tagsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a tag to a card",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		req := &api.AddTagRequest{Name: name}
		tag, err := apiClient.AddCardTag(cardID, req)
		if err != nil {
			return err
		}
		outputJSON(tag)
		return nil
	},
}

var tagsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a tag from a card",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		tagID, err := cmd.Flags().GetInt("tag-id")
		if err != nil {
			return err
		}
		if tagID == 0 {
			return fmt.Errorf("--tag-id is required")
		}
		if err := apiClient.RemoveCardTag(cardID, tagID); err != nil {
			return err
		}
		outputJSON(map[string]int{"removed_tag_id": tagID})
		return nil
	},
}

func init() {
	tagsCmd.AddCommand(tagsListCmd)
	tagsCmd.AddCommand(tagsCardTagsCmd)
	tagsCmd.AddCommand(tagsAddCmd)
	tagsCmd.AddCommand(tagsRemoveCmd)

	tagsCardTagsCmd.Flags().Int("card-id", 0, "Card ID (required)")
	tagsCardTagsCmd.MarkFlagRequired("card-id")

	tagsAddCmd.Flags().Int("card-id", 0, "Card ID (required)")
	tagsAddCmd.Flags().String("name", "", "Tag name (required)")
	tagsAddCmd.MarkFlagRequired("card-id")
	tagsAddCmd.MarkFlagRequired("name")

	tagsRemoveCmd.Flags().Int("card-id", 0, "Card ID (required)")
	tagsRemoveCmd.Flags().Int("tag-id", 0, "Tag ID (required)")
	tagsRemoveCmd.MarkFlagRequired("card-id")
	tagsRemoveCmd.MarkFlagRequired("tag-id")
}
