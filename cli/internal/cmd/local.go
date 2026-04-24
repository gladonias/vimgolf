package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/gadonias/vimgolf/cli/internal/challenge"
	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local <infile> <outfile>",
	Short: "Launch a local challenge",
	Long:  "Launch a local VimGolf challenge. A temporary copy of the input file is made;\nthe original files will not be touched.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLocal(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
}

func runLocal(infile, outfile string) error {
	for _, f := range []string{infile, outfile} {
		info, err := os.Stat(f)
		if err != nil || info.IsDir() {
			return fmt.Errorf("%s must exist and be a regular file", f)
		}
	}

	id := filepath.Base(infile)
	ch := challenge.New(id, "")
	if err := ch.SetupLocal(infile, outfile); err != nil {
		return err
	}
	defer ch.Cleanup()

	return play(ch)
}
