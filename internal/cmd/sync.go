package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/db"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/output"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/syncer"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync Kaiten data to local SQLite database",
	Long: `Fetch boards, cards, members, and tags from the Kaiten API and store them
in a local SQLite database for offline querying and searching.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		boardID, _ := cmd.Flags().GetInt("board-id")
		spaceID, _ := cmd.Flags().GetInt("space-id")
		all, _ := cmd.Flags().GetBool("all")
		format, _ := cmd.Flags().GetString("output")

		// Validate exactly one scope.
		scopes := 0
		if boardID > 0 {
			scopes++
		}
		if spaceID > 0 {
			scopes++
		}
		if all {
			scopes++
		}
		if scopes != 1 {
			return fmt.Errorf("exactly one of --board-id, --space-id, or --all must be specified")
		}

		database, err := db.Open(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer database.Close()

		s := syncer.New(apiClient, database)

		var result syncer.SyncResult
		switch {
		case boardID > 0:
			result, err = s.SyncBoard(boardID)
		case spaceID > 0:
			result, err = s.SyncSpace(spaceID)
		case all:
			result, err = s.SyncAll()
		}

		// Always output the result (even partial on error).
		if outErr := output.Format(os.Stdout, result, format, output.Options{}); outErr != nil {
			return fmt.Errorf("output: %w", outErr)
		}

		return err
	},
}

func init() {
	syncCmd.Flags().Int("board-id", 0, "Sync a single board by ID")
	syncCmd.Flags().Int("space-id", 0, "Sync all boards in a space by space ID")
	syncCmd.Flags().Bool("all", false, "Sync all accessible spaces, boards, and cards")
	syncCmd.Flags().StringP("output", "o", "json", "Output format (json, jsonl, yaml, table, csv)")
}
