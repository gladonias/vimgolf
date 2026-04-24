package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"codeberg.org/gadonias/vimgolf/cli/internal/challenge"
	"codeberg.org/gadonias/vimgolf/cli/internal/config"
	"codeberg.org/gadonias/vimgolf/cli/internal/ui"
)

func play(ch *challenge.Challenge) error {
	for {
		if err := ch.Start(); err != nil {
			return err
		}

		ui.Warn(fmt.Sprintf("Launching VimGolf session for challenge: %s", ch.ID))

		if err := ch.LaunchVim(); err != nil {
			ui.Error("Vim did not exit properly.")
			ui.Error("Please ensure you can execute 'vim' from the command line.")
			return err
		}

		logData, err := os.ReadFile(ch.LogPath)
		if err != nil {
			return fmt.Errorf("failed to read keylog: %w", err)
		}

		score := countKeystrokes(logData)

		ui.Info("\nHere are your keystrokes:")
		printKeystrokes(logData)
		fmt.Println()

		correct, err := ch.Correct()
		if err != nil {
			return err
		}

		if !correct {
			ui.Error(fmt.Sprintf("\nUh oh, looks like your entry does not match the desired output."))
			ui.Error(fmt.Sprintf("Your score for this failed attempt was: %d", score))

			action, err := promptFailedAttempt(ch)
			if err != nil {
				return err
			}
			if action == "retry" {
				continue
			}
			ui.Info("\nThanks for playing!")
			return nil
		}

		ui.Info(fmt.Sprintf("\nSuccess! Your output matches. Your score: %d", score))

		action, err := promptSuccess(ch)
		if err != nil {
			return err
		}
		if action == "retry" {
			continue
		}
		ui.Info("\nThanks for playing!")
		return nil
	}
}

func promptFailedAttempt(ch *challenge.Challenge) (string, error) {
	for {
		choice, err := ui.AskChoice("Choice> ", []string{
			"[d] Show diff",
			"[r] Retry the current challenge",
			"[q] Quit vimgolf",
		})
		if err != nil {
			return "quit", err
		}

		switch choice {
		case "d":
			ui.Warn("Showing vimdiff of your attempt (left) and correct output (right)")
			ch.ShowDiff()
		case "r":
			ui.Warn("Retrying current challenge...")
			return "retry", nil
		case "q":
			return "quit", nil
		}
	}
}

func promptSuccess(ch *challenge.Challenge) (string, error) {
	for {
		var choices []string

		canUpload := false
		if ch.Remote {
			cfg, err := config.Load()
			if err == nil && cfg.Key != "" {
				canUpload = true
				choices = append(choices,
					"[w] Upload result and retry the challenge",
					"[x] Upload result and quit",
				)
			} else {
				choices = append(choices,
					"[s] Set up vimgolf.com key to submit result",
				)
			}
		}

		choices = append(choices,
			"[r] Do not upload result and retry the challenge",
			"[q] Do not upload result and quit",
		)

		choice, err := ui.AskChoice("Choice> ", choices)
		if err != nil {
			return "quit", err
		}

		switch choice {
		case "w":
			if canUpload {
				if upload(ch) {
					return "retry", nil
				}
			}
		case "x":
			if canUpload {
				upload(ch)
				return "quit", nil
			}
		case "s":
			if err := runSetup(); err != nil {
				ui.Error(fmt.Sprintf("Setup failed: %v", err))
			}
		case "r":
			return "retry", nil
		case "q":
			return "quit", nil
		}
	}
}

func upload(ch *challenge.Challenge) bool {
	ui.Warn("Uploading to VimGolf...")

	cfg, err := config.Load()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load config: %v", err))
		return false
	}

	logData, err := os.ReadFile(ch.LogPath)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to read keylog: %v", err))
		return false
	}

	form := url.Values{
		"challenge_id": {ch.ID},
		"apikey":       {cfg.Key},
		"entry":        {string(logData)},
	}

	uploadURL := fmt.Sprintf("%s/entry.json", config.Host())
	resp, err := http.PostForm(uploadURL, form)
	if err != nil {
		ui.Error(fmt.Sprintf("Upload failed: %v", err))
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		ui.Warn("Uploaded entry.")
		ui.Warn(fmt.Sprintf("View the leaderboard: %s/challenges/%s", config.Host(), ch.ID))
		return true
	}

	ui.Error("Uh oh, upload failed. You're not cheating are you? :-)")
	return false
}

// Placeholder keystroke counting — counts raw bytes as keystrokes.
// The full keylog parser (Stage 3) will replace this with proper
// multi-byte Vim keycode handling.
func countKeystrokes(data []byte) int {
	count := 0
	i := 0
	for i < len(data) {
		if data[i] == 0x80 && i+2 < len(data) {
			i += 3
		} else {
			i++
		}
		count++
	}
	return count
}

func printKeystrokes(data []byte) {
	i := 0
	for i < len(data) {
		if data[i] == 0x80 && i+2 < len(data) {
			ui.PrintKeystroke(fmt.Sprintf("<0x%02x%02x>", data[i+1], data[i+2]))
			i += 3
		} else {
			b := data[i]
			switch {
			case b == 0x1b:
				ui.PrintKeystroke("<Esc>")
			case b == 0x0d:
				ui.PrintKeystroke("<CR>")
			case b == 0x0a:
				ui.PrintKeystroke("<NL>")
			case b == 0x09:
				ui.PrintKeystroke("<Tab>")
			case b >= 32 && b <= 126:
				fmt.Print(string(b))
			default:
				ui.PrintKeystroke(fmt.Sprintf("<C-%c>", b^0x40))
			}
			i++
		}
	}
}
