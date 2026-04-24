package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"codeberg.org/gadonias/vimgolf/cli/internal/challenge"
	"codeberg.org/gadonias/vimgolf/cli/internal/config"
	"codeberg.org/gadonias/vimgolf/cli/internal/keylog"
	"codeberg.org/gadonias/vimgolf/cli/internal/ui"
)

var uploadClient = &http.Client{Timeout: 30 * time.Second}

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

		strokes := keylog.Parse(logData)
		score := len(strokes)

		ui.Info("\nHere are your keystrokes:")
		for _, s := range strokes {
			ui.PrintKeystroke(s.Display)
		}
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
			if err := ch.ShowDiff(); err != nil {
				ui.Error(fmt.Sprintf("Failed to show diff: %v", err))
			}
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
				if upload(ch) {
					return "quit", nil
				}
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

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("challenge_id", ch.ID)
	writer.WriteField("apikey", cfg.Key)

	part, err := writer.CreateFormField("entry")
	if err != nil {
		ui.Error(fmt.Sprintf("Upload failed: %v", err))
		return false
	}
	part.Write(logData)
	writer.Close()

	uploadURL := fmt.Sprintf("%s/entry.json", config.Host())
	req, err := http.NewRequest("POST", uploadURL, &body)
	if err != nil {
		ui.Error(fmt.Sprintf("Upload failed: %v", err))
		return false
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := uploadClient.Do(req)
	if err != nil {
		ui.Error(fmt.Sprintf("Upload failed: %v", err))
		return false
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to read server response: %v", err))
		return false
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		ui.Error(fmt.Sprintf("Failed to parse server response: %v", err))
		return false
	}

	if result.Status == "ok" {
		ui.Warn("Uploaded entry.")
		ui.Warn(fmt.Sprintf("View the leaderboard: %s/challenges/%s", config.Host(), ch.ID))
		return true
	}

	ui.Error("Uh oh, upload failed. You're not cheating are you? :-)")
	return false
}

