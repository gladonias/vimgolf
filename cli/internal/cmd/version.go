package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of the VimGolf client",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("VimGolf CLI v%s (Go)\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
