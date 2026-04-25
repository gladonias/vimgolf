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
	existing, _ := config.Load()

	ui.Info("\nLet's setup your VimGolf key...")
	ui.Warn("1) Open vimgolf.com in your browser.")
	ui.Warn("2) Click \"Sign in\".")
	ui.Warn("3) Sign in with Twitter or GitHub.")
	ui.Warn("4) Once signed in, copy your key (black box, top right).")

	keyPrompt := "\nPaste your VimGolf key:"
	if existing != nil && existing.Key != "" {
		keyPrompt = "\nPaste your VimGolf key (Enter to keep current):"
	}

	key, err := ui.Ask(keyPrompt)
	if err != nil {
		return err
	}

	if key == "" && existing != nil && existing.Key != "" {
		key = existing.Key
	}

	if !config.ValidateKey(key) {
		return fmt.Errorf("invalid key, please double check your key on vimgolf.com")
	}

	usernamePrompt := "Your vimgolf.com username (the @handle shown on the site after sign-in, optional):"
	if existing != nil && existing.Username != "" {
		usernamePrompt = fmt.Sprintf("Your vimgolf.com username [%s] (Enter to keep current):", existing.Username)
	}

	username, err := ui.Ask(usernamePrompt)
	if err != nil {
		return err
	}

	if username == "" && existing != nil {
		username = existing.Username
	}

	if err := config.Save(&config.Config{Key: key, Username: username}); err != nil {
		return err
	}

	ui.Info("Saved. Happy golfing!")
	return nil
}
