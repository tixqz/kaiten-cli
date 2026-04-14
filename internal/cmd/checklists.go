package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/api"
)

var checklistsCmd = &cobra.Command{
	Use:   "checklists",
	Short: "Manage card checklists",
}

var checklistsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a checklist with its items",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		checklistID, err := cmd.Flags().GetInt("checklist-id")
		if err != nil {
			return err
		}
		if checklistID == 0 {
			return fmt.Errorf("--checklist-id is required")
		}
		checklist, err := apiClient.GetChecklist(cardID, checklistID)
		if err != nil {
			return err
		}
		outputJSON(checklist)
		return nil
	},
}

var checklistsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a checklist on a card",
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
		req := &api.CreateChecklistRequest{Name: name}
		checklist, err := apiClient.CreateChecklist(cardID, req)
		if err != nil {
			return err
		}
		outputJSON(checklist)
		return nil
	},
}

var checklistsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a checklist",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		checklistID, err := cmd.Flags().GetInt("checklist-id")
		if err != nil {
			return err
		}
		if checklistID == 0 {
			return fmt.Errorf("--checklist-id is required")
		}
		if err := apiClient.DeleteChecklist(cardID, checklistID); err != nil {
			return err
		}
		outputJSON(map[string]int{"deleted": checklistID})
		return nil
	},
}

var checklistsAddItemCmd = &cobra.Command{
	Use:   "add-item",
	Short: "Add an item to a checklist",
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := cmd.Flags().GetInt("card-id")
		if err != nil {
			return err
		}
		if cardID == 0 {
			return fmt.Errorf("--card-id is required")
		}
		checklistID, err := cmd.Flags().GetInt("checklist-id")
		if err != nil {
			return err
		}
		if checklistID == 0 {
			return fmt.Errorf("--checklist-id is required")
		}
		text, err := cmd.Flags().GetString("text")
		if err != nil {
			return err
		}
		if text == "" {
			return fmt.Errorf("--text is required")
		}
		req := &api.AddChecklistItemRequest{Text: text}
		item, err := apiClient.AddChecklistItem(cardID, checklistID, req)
		if err != nil {
			return err
		}
		outputJSON(item)
		return nil
	},
}

var checklistsCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Mark a checklist item as checked",
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateChecklistItemChecked(cmd, true)
	},
}

var checklistsUncheckCmd = &cobra.Command{
	Use:   "uncheck",
	Short: "Mark a checklist item as unchecked",
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateChecklistItemChecked(cmd, false)
	},
}

func updateChecklistItemChecked(cmd *cobra.Command, checked bool) error {
	cardID, err := cmd.Flags().GetInt("card-id")
	if err != nil {
		return err
	}
	if cardID == 0 {
		return fmt.Errorf("--card-id is required")
	}
	checklistID, err := cmd.Flags().GetInt("checklist-id")
	if err != nil {
		return err
	}
	if checklistID == 0 {
		return fmt.Errorf("--checklist-id is required")
	}
	itemID, err := cmd.Flags().GetInt("item-id")
	if err != nil {
		return err
	}
	if itemID == 0 {
		return fmt.Errorf("--item-id is required")
	}
	req := &api.UpdateChecklistItemRequest{Checked: &checked}
	item, err := apiClient.UpdateChecklistItem(cardID, checklistID, itemID, req)
	if err != nil {
		return err
	}
	outputJSON(item)
	return nil
}

func init() {
	checklistsCmd.AddCommand(checklistsGetCmd)
	checklistsCmd.AddCommand(checklistsCreateCmd)
	checklistsCmd.AddCommand(checklistsDeleteCmd)
	checklistsCmd.AddCommand(checklistsAddItemCmd)
	checklistsCmd.AddCommand(checklistsCheckCmd)
	checklistsCmd.AddCommand(checklistsUncheckCmd)

	checklistsGetCmd.Flags().Int("card-id", 0, "Card ID (required)")
	checklistsGetCmd.Flags().Int("checklist-id", 0, "Checklist ID (required)")
	checklistsGetCmd.MarkFlagRequired("card-id")
	checklistsGetCmd.MarkFlagRequired("checklist-id")

	checklistsCreateCmd.Flags().Int("card-id", 0, "Card ID (required)")
	checklistsCreateCmd.Flags().String("name", "", "Checklist name")
	checklistsCreateCmd.MarkFlagRequired("card-id")

	checklistsDeleteCmd.Flags().Int("card-id", 0, "Card ID (required)")
	checklistsDeleteCmd.Flags().Int("checklist-id", 0, "Checklist ID (required)")
	checklistsDeleteCmd.MarkFlagRequired("card-id")
	checklistsDeleteCmd.MarkFlagRequired("checklist-id")

	checklistsAddItemCmd.Flags().Int("card-id", 0, "Card ID (required)")
	checklistsAddItemCmd.Flags().Int("checklist-id", 0, "Checklist ID (required)")
	checklistsAddItemCmd.Flags().String("text", "", "Checklist item text (required)")
	checklistsAddItemCmd.MarkFlagRequired("card-id")
	checklistsAddItemCmd.MarkFlagRequired("checklist-id")
	checklistsAddItemCmd.MarkFlagRequired("text")

	checklistsCheckCmd.Flags().Int("card-id", 0, "Card ID (required)")
	checklistsCheckCmd.Flags().Int("checklist-id", 0, "Checklist ID (required)")
	checklistsCheckCmd.Flags().Int("item-id", 0, "Checklist Item ID (required)")
	checklistsCheckCmd.MarkFlagRequired("card-id")
	checklistsCheckCmd.MarkFlagRequired("checklist-id")
	checklistsCheckCmd.MarkFlagRequired("item-id")

	checklistsUncheckCmd.Flags().Int("card-id", 0, "Card ID (required)")
	checklistsUncheckCmd.Flags().Int("checklist-id", 0, "Checklist ID (required)")
	checklistsUncheckCmd.Flags().Int("item-id", 0, "Checklist Item ID (required)")
	checklistsUncheckCmd.MarkFlagRequired("card-id")
	checklistsUncheckCmd.MarkFlagRequired("checklist-id")
	checklistsUncheckCmd.MarkFlagRequired("item-id")
}
