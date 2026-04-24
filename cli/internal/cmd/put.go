package cmd

import (
	"fmt"
	"os"
	"strings"

	"codeberg.org/gadonias/vimgolf/cli/internal/challenge"
	"codeberg.org/gadonias/vimgolf/cli/internal/config"
	"codeberg.org/gadonias/vimgolf/cli/internal/ui"
	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put <challenge-id>",
	Short: "Launch a vimgolf.com challenge",
	Long:  "Download and launch a VimGolf session for the specified challenge ID.\nFind active challenges at vimgolf.com.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPut(args[0])
	},
}

func init() {
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
