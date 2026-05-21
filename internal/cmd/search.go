package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/db"
	"gitlab.life-pay.ru/ai/kaiten-cli/internal/output"
)

const searchTimeLayout = "2006-01-02T15:04:05.000000000Z"

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search local Kaiten data",
	Long: `Search synced Kaiten data stored in the local SQLite database.

Data must be synced first using "kaiten sync" before searching.`,
}

var searchCardsCmd = &cobra.Command{
	Use:   "cards",
	Short: "Search cards in the local database",
	Long: `Search cards in the local SQLite database using various filters.

All filtering is performed against the local database; no API calls are made.
Use "kaiten sync" to populate the database before searching.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check DB file exists.
		if _, err := os.Stat(cfg.DBPath); os.IsNotExist(err) {
			return fmt.Errorf("local database not found at %s\nhint: run 'kaiten sync --all' or 'kaiten sync --board-id <id>' to create and populate the local database", cfg.DBPath)
		}

		database, err := db.Open(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer database.Close()

		q, err := buildSearchQuery(cmd)
		if err != nil {
			return err
		}

		cards, err := database.SearchCards(q)
		if err != nil {
			return fmt.Errorf("search cards: %w", err)
		}

		format, _ := cmd.Flags().GetString("output")
		fieldsStr, _ := cmd.Flags().GetString("fields")
		noDescriptions, _ := cmd.Flags().GetBool("no-descriptions")
		quiet, _ := cmd.Flags().GetBool("quiet")

		var fields []string
		if fieldsStr != "" {
			fields = strings.Split(fieldsStr, ",")
			for i := range fields {
				fields[i] = strings.TrimSpace(fields[i])
			}
		}

		opts := output.Options{
			Fields:         fields,
			NoDescriptions: noDescriptions,
			Quiet:          quiet,
		}

		if err := output.Format(cmd.OutOrStdout(), cards, format, opts); err != nil {
			return fmt.Errorf("output: %w", err)
		}

		return nil
	},
}

// buildSearchQuery constructs a db.SearchQuery from cobra command flags.
// Uses cmd.Flags().Changed() so only explicitly-provided flags are applied;
// leftover flag state from prior command invocations is ignored for filter flags.
func buildSearchQuery(cmd *cobra.Command) (db.SearchQuery, error) {
	var q db.SearchQuery

	if cmd.Flags().Changed("text") {
		v, _ := cmd.Flags().GetString("text")
		q.Text = v
	}

	if cmd.Flags().Changed("board-id") {
		v, _ := cmd.Flags().GetInt("board-id")
		if v != 0 {
			q.BoardID = intPtr(v)
		}
	}

	if cmd.Flags().Changed("space-id") {
		v, _ := cmd.Flags().GetInt("space-id")
		if v != 0 {
			q.SpaceID = intPtr(v)
		}
	}

	if cmd.Flags().Changed("owner-id") {
		v, _ := cmd.Flags().GetInt("owner-id")
		if v != 0 {
			q.OwnerID = intPtr(v)
		}
	}

	if cmd.Flags().Changed("member-id") {
		v, _ := cmd.Flags().GetInt("member-id")
		if v != 0 {
			q.MemberID = intPtr(v)
		}
	}

	if cmd.Flags().Changed("tag-id") {
		v, _ := cmd.Flags().GetInt("tag-id")
		if v != 0 {
			q.TagID = intPtr(v)
		}
	}

	if cmd.Flags().Changed("condition") {
		v, _ := cmd.Flags().GetString("condition")
		if v != "" {
			cond, err := parseCondition(v)
			if err != nil {
				return q, err
			}
			q.Condition = intPtr(cond)
		}
	}

	if cmd.Flags().Changed("created-from") {
		v, _ := cmd.Flags().GetString("created-from")
		if v != "" {
			parsed, err := parseSearchDate(v, false)
			if err != nil {
				return q, fmt.Errorf("--created-from: %w", err)
			}
			q.CreatedFrom = strPtr(parsed)
		}
	}

	if cmd.Flags().Changed("created-to") {
		v, _ := cmd.Flags().GetString("created-to")
		if v != "" {
			parsed, err := parseSearchDate(v, true)
			if err != nil {
				return q, fmt.Errorf("--created-to: %w", err)
			}
			q.CreatedTo = strPtr(parsed)
		}
	}

	if cmd.Flags().Changed("updated-from") {
		v, _ := cmd.Flags().GetString("updated-from")
		if v != "" {
			parsed, err := parseSearchDate(v, false)
			if err != nil {
				return q, fmt.Errorf("--updated-from: %w", err)
			}
			q.UpdatedFrom = strPtr(parsed)
		}
	}

	if cmd.Flags().Changed("updated-to") {
		v, _ := cmd.Flags().GetString("updated-to")
		if v != "" {
			parsed, err := parseSearchDate(v, true)
			if err != nil {
				return q, fmt.Errorf("--updated-to: %w", err)
			}
			q.UpdatedTo = strPtr(parsed)
		}
	}

	if cmd.Flags().Changed("completed-from") {
		v, _ := cmd.Flags().GetString("completed-from")
		if v != "" {
			parsed, err := parseSearchDate(v, false)
			if err != nil {
				return q, fmt.Errorf("--completed-from: %w", err)
			}
			q.CompletedFrom = strPtr(parsed)
		}
	}

	if cmd.Flags().Changed("completed-to") {
		v, _ := cmd.Flags().GetString("completed-to")
		if v != "" {
			parsed, err := parseSearchDate(v, true)
			if err != nil {
				return q, fmt.Errorf("--completed-to: %w", err)
			}
			q.CompletedTo = strPtr(parsed)
		}
	}

	limit, _ := cmd.Flags().GetInt("limit")
	if !cmd.Flags().Changed("limit") || limit > 0 {
		q.Limit = limit
	}

	if cmd.Flags().Changed("offset") {
		v, _ := cmd.Flags().GetInt("offset")
		if v > 0 {
			q.Offset = v
		}
	}

	if cmd.Flags().Changed("sort") {
		v, _ := cmd.Flags().GetString("sort")
		if v != "" {
			q.Sort = v
		}
	}

	return q, nil
}

// parseCondition converts a condition flag value to an integer.
// Accepts numeric strings ("0", "1", "2", "3") and named aliases.
func parseCondition(s string) (int, error) {
	aliases := map[string]int{
		"active":   1,
		"archived": 2,
		"done":     3,
	}
	if v, ok := aliases[strings.ToLower(s)]; ok {
		return v, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid condition %q: expected numeric (0-3) or alias (active, archived, done)", s)
	}
	if v < 0 || v > 3 {
		return 0, fmt.Errorf("invalid condition %d: expected 0-3", v)
	}
	return v, nil
}

// parseSearchDate parses a date string, accepting YYYY-MM-DD or RFC3339.
// Date-only values are normalized to start-of-day for lower bounds and
// end-of-day for upper bounds, so --completed-to 2025-12-31 includes that day.
func parseSearchDate(s string, endOfDay bool) (string, error) {
	// Try RFC3339 first.
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t.UTC().Format(searchTimeLayout), nil
	}

	// Try YYYY-MM-DD.
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		if endOfDay {
			t = t.Add(24*time.Hour - time.Nanosecond)
		}
		return t.UTC().Format(searchTimeLayout), nil
	}

	return "", fmt.Errorf("invalid date %q, expected YYYY-MM-DD or RFC3339", s)
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func init() {
	searchCmd.AddCommand(searchCardsCmd)

	// Filters
	searchCardsCmd.Flags().String("text", "", "Full-text search over card titles and descriptions")
	searchCardsCmd.Flags().Int("board-id", 0, "Filter by board ID")
	searchCardsCmd.Flags().Int("space-id", 0, "Filter by space ID")
	searchCardsCmd.Flags().Int("owner-id", 0, "Filter by owner user ID")
	searchCardsCmd.Flags().Int("member-id", 0, "Filter by member user ID")
	searchCardsCmd.Flags().Int("tag-id", 0, "Filter by tag ID")
	searchCardsCmd.Flags().String("condition", "", `Filter by condition: numeric (0-3) or alias (active, archived, done)`)
	searchCardsCmd.Flags().String("created-from", "", "Filter cards created on or after this date (YYYY-MM-DD or RFC3339)")
	searchCardsCmd.Flags().String("created-to", "", "Filter cards created on or before this date (YYYY-MM-DD or RFC3339)")
	searchCardsCmd.Flags().String("updated-from", "", "Filter cards updated on or after this date (YYYY-MM-DD or RFC3339)")
	searchCardsCmd.Flags().String("updated-to", "", "Filter cards updated on or before this date (YYYY-MM-DD or RFC3339)")
	searchCardsCmd.Flags().String("completed-from", "", "Filter cards completed on or after this date (YYYY-MM-DD or RFC3339)")
	searchCardsCmd.Flags().String("completed-to", "", "Filter cards completed on or before this date (YYYY-MM-DD or RFC3339)")

	// Pagination / sort
	searchCardsCmd.Flags().Int("limit", 100, "Maximum number of cards to return (0 for all)")
	searchCardsCmd.Flags().Int("offset", 0, "Number of cards to skip before returning results")
	searchCardsCmd.Flags().String("sort", "updated_desc", `Sort order: created, created_desc, updated, updated_desc, completed, completed_desc, id, id_desc`)

	// Output controls
	searchCardsCmd.Flags().StringP("output", "o", "json", "Output format: json, jsonl, yaml, table, csv")
	searchCardsCmd.Flags().String("fields", "", "Comma-separated list of fields to include")
	searchCardsCmd.Flags().Bool("no-descriptions", false, "Exclude description fields from output")
	searchCardsCmd.Flags().BoolP("quiet", "q", false, "Print only card IDs, one per line")
}
