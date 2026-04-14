package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var spacesCmd = &cobra.Command{
	Use:   "spaces",
	Short: "Manage spaces",
}

var spacesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all spaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		spaces, err := apiClient.ListSpaces()
		if err != nil {
			return err
		}
		outputJSON(spaces)
		return nil
	},
}

var spacesGetCmd = &cobra.Command{
	Use:   "get [space_id]",
	Short: "Get a space by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid space ID: %s", args[0])
		}
		space, err := apiClient.GetSpace(id)
		if err != nil {
			return err
		}
		outputJSON(space)
		return nil
	},
}

func init() {
	spacesCmd.AddCommand(spacesListCmd)
	spacesCmd.AddCommand(spacesGetCmd)
}
