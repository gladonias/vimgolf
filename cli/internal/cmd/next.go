package cmd

import (
	"fmt"
	"slices"

	"codeberg.org/gadonias/vimgolf/cli/internal/config"
	"codeberg.org/gadonias/vimgolf/cli/internal/leaderboard"
	"codeberg.org/gadonias/vimgolf/cli/internal/listing"
	"codeberg.org/gadonias/vimgolf/cli/internal/ui"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Play challenges sequentially from oldest to newest",
	Long: `Play through VimGolf challenges in order, from oldest to newest.
Each challenge is repeated until you match the best score on the leaderboard.
Once beaten, the next challenge in sequence is selected.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNext()
	},
}

func init() {
	rootCmd.AddCommand(nextCmd)
}

func runNext() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config required for 'next': %w", err)
	}
	if cfg.Username == "" {
		return fmt.Errorf("username required — run 'vimgolf setup' to set it")
	}

	state, _ := config.LoadNextState()

	if state != nil && state.ChallengeID != "" {
		result, err := leaderboard.Fetch(state.ChallengeID, cfg.Username)
		if err != nil {
			ui.Warn(fmt.Sprintf("Could not check challenge %s: %v", state.ChallengeID, err))
			ui.Warn("Advancing to next challenge...")
			return advanceNext(state.ChallengeID)
		}

		// Top[0] is rank #1 (lowest score). Beaten means user matched it.
		beaten := len(result.Top) > 0 && result.UserEntry != nil &&
			result.UserEntry.Score <= result.Top[0].Score

		if !beaten {
			if result.Stats.Title != "" {
				ui.Info(fmt.Sprintf("Resuming: %q", result.Stats.Title))
			}
			if len(result.Top) > 0 {
				msg := fmt.Sprintf("Best score: %d", result.Top[0].Score)
				if result.UserEntry != nil {
					msg += fmt.Sprintf(" | Your best: %d", result.UserEntry.Score)
				}
				ui.Warn(msg)
			}
			return runPut(state.ChallengeID)
		}

		ui.Info("Challenge beaten! Advancing to next...")
		return advanceNext(state.ChallengeID)
	}

	return advanceNext("")
}

func advanceNext(currentID string) error {
	ui.Warn("Fetching all challenges from vimgolf.com...")
	challenges, err := listing.FetchAll(0)
	if err != nil {
		return fmt.Errorf("failed to fetch challenges: %w", err)
	}

	if len(challenges) == 0 {
		return fmt.Errorf("no challenges found on vimgolf.com")
	}

	slices.Reverse(challenges)

	var next *listing.Challenge
	if currentID != "" {
		for i, c := range challenges {
			if c.ID == currentID {
				if i+1 < len(challenges) {
					next = &challenges[i+1]
				} else {
					ui.Info("Congratulations! You've completed every challenge on VimGolf!")
					return nil
				}
				break
			}
		}
	}

	if next == nil {
		next = &challenges[0]
	}

	if err := config.SaveNextState(&config.NextState{ChallengeID: next.ID}); err != nil {
		return fmt.Errorf("failed to save progress: %w", err)
	}

	ui.Info(fmt.Sprintf("Next challenge: %q (%d entries)", next.Title, next.Entries))
	return runPut(next.ID)
}
