package cmd

import (
	"fmt"
	"strings"

	"codeberg.org/gadonias/vimgolf/cli/internal/config"
	"codeberg.org/gadonias/vimgolf/cli/internal/leaderboard"
	"codeberg.org/gadonias/vimgolf/cli/internal/ui"
	"github.com/spf13/cobra"
)

var leaderboardCmd = &cobra.Command{
	Use:     "leaderboard <challenge-id>",
	Aliases: []string{"lb"},
	Short:   "Show the leaderboard for a challenge",
	Long:    "Display the top 10 entries for a challenge and your position if you have submitted.",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLeaderboard(args[0])
	},
}

func init() {
	rootCmd.AddCommand(leaderboardCmd)
}

func runLeaderboard(challengeID string) error {
	ui.Warn(fmt.Sprintf("Fetching leaderboard for challenge: %s", challengeID))

	var username string
	cfg, err := config.Load()
	if err == nil && cfg.Key != "" {
		username = cfg.Username
	}

	result, err := leaderboard.Fetch(challengeID, username)
	if err != nil {
		return err
	}

	if result.Stats.Title != "" {
		ui.Info(fmt.Sprintf("\n  %s", result.Stats.Title))
	}
	if result.Stats.Golfers > 0 {
		fmt.Printf("  %d golfers, %d entries\n", result.Stats.Golfers, result.Stats.Entries)
	}

	fmt.Println()
	fmt.Println("  Rank  Score  Player")
	fmt.Println("  " + strings.Repeat("-", 40))

	limit := 10
	if len(result.Top) < limit {
		limit = len(result.Top)
	}

	for i := 0; i < limit; i++ {
		e := result.Top[i]
		marker := "  "
		if username != "" && strings.EqualFold(e.Username, username) {
			marker = "> "
		}
		fmt.Printf("%s%4d  %5d  %s (@%s)\n", marker, e.Rank, e.Score, e.Name, e.Username)
	}

	if username != "" && result.UserEntry != nil && result.UserEntry.Rank > 10 {
		fmt.Println("  " + strings.Repeat(".", 40))
		e := result.UserEntry
		fmt.Printf("> %4d  %5d  You (@%s)\n", e.Rank, e.Score, e.Username)
	}

	fmt.Println()
	return nil
}
