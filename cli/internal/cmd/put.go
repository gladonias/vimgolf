package cmd

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"

	"codeberg.org/gadonias/vimgolf/cli/internal/challenge"
	"codeberg.org/gadonias/vimgolf/cli/internal/config"
	"codeberg.org/gadonias/vimgolf/cli/internal/leaderboard"
	"codeberg.org/gadonias/vimgolf/cli/internal/listing"
	"codeberg.org/gadonias/vimgolf/cli/internal/ui"
	"github.com/spf13/cobra"
)

var minEntries int

var putCmd = &cobra.Command{
	Use:   "put [challenge-id]",
	Short: "Launch a vimgolf.com challenge",
	Long: `Download and launch a VimGolf session for the specified challenge ID.
If no challenge ID is given, pick a random unplayed challenge.
Find active challenges at vimgolf.com.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return runPut(args[0])
		}
		return runRandomPut(minEntries)
	},
}

func init() {
	putCmd.Flags().IntVar(&minEntries, "min-entries", 100,
		"minimum number of entries for random challenge selection")
	rootCmd.AddCommand(putCmd)
}

func runPut(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("challenge ID cannot be empty")
	}

	putDir := config.PutDir()
	if err := os.MkdirAll(putDir, 0o755); err != nil {
		return fmt.Errorf("failed to create put directory: %w", err)
	}

	ui.Warn(fmt.Sprintf("Downloading VimGolf challenge: %s", id))
	ch := challenge.New(id, putDir)
	if err := ch.Download(); err != nil {
		return err
	}
	defer ch.Cleanup()

	return play(ch)
}

const maxCompletionChecks = 20

func runRandomPut(minEntries int) error {
	var username string
	cfg, err := config.Load()
	if err == nil {
		username = cfg.Username
	}
	if username == "" {
		ui.Warn("No username configured — cannot filter completed challenges.")
		ui.Warn("Run 'vimgolf setup' to set your username.")
	}

	ui.Warn(fmt.Sprintf("Fetching challenges from vimgolf.com (min entries: %d)...", minEntries))
	challenges, err := listing.FetchAll(minEntries)
	if err != nil {
		return fmt.Errorf("failed to fetch challenges: %w", err)
	}

	if len(challenges) == 0 {
		return fmt.Errorf("no challenges found with at least %d entries", minEntries)
	}

	ui.Info(fmt.Sprintf("Found %d challenges with at least %d entries.", len(challenges), minEntries))

	rand.Shuffle(len(challenges), func(i, j int) {
		challenges[i], challenges[j] = challenges[j], challenges[i]
	})

	var selected *listing.Challenge
	if username != "" {
		ui.Warn("Picking a random unplayed challenge...")
		checks := maxCompletionChecks
		if checks > len(challenges) {
			checks = len(challenges)
		}
		for i := 0; i < checks; i++ {
			c := challenges[i]
			entry, err := leaderboard.FetchUserEntry(c.ID, username)
			if err != nil {
				ui.Debug(fmt.Sprintf("Error checking %q: %v", c.Title, err))
				continue
			}
			if entry == nil {
				selected = &c
				break
			}
			ui.Debug(fmt.Sprintf("Skipping %q (already completed)", c.Title))
		}
		if selected == nil {
			ui.Warn("Could not find an unplayed challenge. Picking a random one.")
			selected = &challenges[0]
		}
	} else {
		selected = &challenges[0]
	}

	ui.Info(fmt.Sprintf("Selected: %q (%d entries)", selected.Title, selected.Entries))
	return runPut(selected.ID)
}
