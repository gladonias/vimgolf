package cmd

import (
	"fmt"

	"codeberg.org/gadonias/vimgolf/cli/internal/config"
	"codeberg.org/gadonias/vimgolf/cli/internal/ui"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure your VimGolf credentials",
	Long: `To participate in challenges, go to vimgolf.com and register an account.
Once signed in, you will get your API key, which you need to set up the client.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetup()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup() error {
	ui.Info("\nLet's setup your VimGolf key...")
	ui.Warn("1) Open vimgolf.com in your browser.")
	ui.Warn("2) Click \"Sign in\".")
	ui.Warn("3) Sign in with Twitter or GitHub.")
	ui.Warn("4) Once signed in, copy your key (black box, top right).")

	key, err := ui.Ask("\nPaste your VimGolf key:")
	if err != nil {
		return err
	}

	if !config.ValidateKey(key) {
		return fmt.Errorf("invalid key, please double check your key on vimgolf.com")
	}

	username, err := ui.Ask("Your vimgolf.com username (for leaderboard, optional):")
	if err != nil {
		return err
	}

	if err := config.Save(&config.Config{Key: key, Username: username}); err != nil {
		return err
	}

	ui.Info("Saved. Happy golfing!")
	return nil
}
