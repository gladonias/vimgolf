package challenge

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func vimBinary() string {
	if v := os.Getenv("GOLFVIM"); v != "" {
		return v
	}
	return "vim"
}

func diffCommand() string {
	if d := os.Getenv("GOLFSHOWDIFF"); d != "" {
		return d
	}
	return "vim -d -n"
}

func (c *Challenge) LaunchVim() error {
	bin := vimBinary()
	parts := strings.Fields(bin)

	args := append(parts[1:],
		"-Z", "-n", "--noplugin",
		"-i", "NONE",
		"+0",
		"-u", c.VimrcPath,
		"-U", "NONE",
		"-W", c.LogPath,
		c.WorkPath,
	)

	if parts[0] == "gvim" || strings.HasSuffix(parts[0], "/gvim") {
		args = append(args, "--nofork")
	}

	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("vim exited with status %d", exitErr.ExitCode())
		}
		return fmt.Errorf("failed to launch vim: %w", err)
	}

	return nil
}

func (c *Challenge) ShowDiff() error {
	parts := strings.Fields(diffCommand())
	args := append(parts[1:], c.WorkPath, c.OutputPath)

	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
