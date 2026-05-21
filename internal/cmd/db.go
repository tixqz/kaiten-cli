package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/db"
)

// dbCmd represents the "kaiten db" parent command.
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the local SQLite database",
}

var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database status and sync information",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If the database file does not exist, return a helpful hint.
		if _, statErr := os.Stat(cfg.DBPath); os.IsNotExist(statErr) {
			outputJSON(map[string]string{
				"error":   fmt.Sprintf("local database not found at %s", cfg.DBPath),
				"hint":    "run 'kaiten sync --all' or 'kaiten sync --board-id <id>' to create and populate the local database",
				"db_path": cfg.DBPath,
			})
			return nil
		}

		database, err := db.Open(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer database.Close()

		// Gather counts.
		cardCount, _ := database.CountCards(db.SearchQuery{})

		spaces, _ := database.ListSpaces()
		spaceCount := len(spaces)

		boardCount := 0
		for _, s := range spaces {
			boards, err := database.ListBoardsBySpace(s.ID)
			if err == nil {
				boardCount += len(boards)
			}
		}

		// Sync state summary.
		syncInfo, _ := database.StatusString()

		outputJSON(map[string]interface{}{
			"db_path":   cfg.DBPath,
			"cards":     cardCount,
			"spaces":    spaceCount,
			"boards":    boardCount,
			"sync_info": syncInfo,
		})
		return nil
	},
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete the local database file",
	Long: `Remove the local SQLite database file. All synced data will be lost.
Use --yes to confirm the operation (non-interactive, safe for automation).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			return fmt.Errorf("this operation requires confirmation; use --yes to confirm")
		}

		if err := os.Remove(cfg.DBPath); err != nil {
			if os.IsNotExist(err) {
				outputJSON(map[string]string{
					"status":  "ok",
					"message": "database does not exist, nothing to reset",
				})
				return nil
			}
			return fmt.Errorf("removing database: %w", err)
		}

		outputJSON(map[string]string{
			"status":  "ok",
			"message": "database removed",
		})
		return nil
	},
}

var dbVacuumCmd = &cobra.Command{
	Use:   "vacuum",
	Short: "Vacuum the local database to reclaim unused space",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer database.Close()

		if err := database.Vacuum(); err != nil {
			return fmt.Errorf("vacuum: %w", err)
		}

		outputJSON(map[string]string{
			"status":  "ok",
			"message": "database vacuumed",
		})
		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbStatusCmd)
	dbCmd.AddCommand(dbResetCmd)
	dbCmd.AddCommand(dbVacuumCmd)

	dbResetCmd.Flags().Bool("yes", false, "Confirm reset without prompting")
}
