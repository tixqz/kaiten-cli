package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/api"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/config"
)

var (
	// Version is set at build time via -ldflags.
	Version = "dev"

	cfg       *config.Config
	apiClient *api.Client
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:     "kaiten",
	Short:   "Kaiten CLI — command-line client for Kaiten project management",
	Long:    "A CLI tool to interact with the Kaiten API for managing spaces, boards, columns, lanes, and cards.",
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if isLocalOnlyCommand(cmd) {
			cfg, err = config.LoadLocal()
			return err
		}

		cfg, err = config.Load()
		if err != nil {
			return err
		}
		apiClient = api.NewClient(cfg)
		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(spacesCmd)
	rootCmd.AddCommand(boardsCmd)
	rootCmd.AddCommand(columnsCmd)
	rootCmd.AddCommand(lanesCmd)
	rootCmd.AddCommand(cardsCmd)
	rootCmd.AddCommand(commentsCmd)
	rootCmd.AddCommand(blockersCmd)
	rootCmd.AddCommand(membersCmd)
	rootCmd.AddCommand(tagsCmd)
	rootCmd.AddCommand(checklistsCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(searchCmd)
}

func isLocalOnlyCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "db", "search":
			return true
		}
	}
	return false
}

// outputJSON marshals v to indented JSON and prints to stdout.
// On error it prints to stderr and exits.
func outputJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
