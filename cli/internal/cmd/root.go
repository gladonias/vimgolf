package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vimgolf",
	Short: "CLI client for VimGolf",
	Long: `VimGolf - Real Vim ninjas count every keystroke.

Download challenges from vimgolf.com, edit the input file to match
the output using Vim, and compete for the fewest keystrokes.`,
}

func Execute() error {
	return rootCmd.Execute()
}
