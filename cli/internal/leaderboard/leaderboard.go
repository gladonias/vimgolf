package leaderboard

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"codeberg.org/gadonias/vimgolf/cli/internal/config"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type Entry struct {
	Rank     int
	Name     string
	Username string
	Score    int
}

type Stats struct {
	Golfers int
	Entries int
	Title   string
}

type Result struct {
	Stats      Stats
	Top        []Entry
	UserEntry  *Entry
}

var (
	entryRe = regexp.MustCompile(
		`<b>#(\d+)</b></a>\s*-\s*(.+?)\s*/\s*<a href="[^"]*">@(\w+)`,
	)
	scoreRe = regexp.MustCompile(
		`font-size:2\.4em"><b><a[^>]*>(\d+)</a>`,
	)
	statsRe = regexp.MustCompile(
		`<b class="stat">(\d+)</b>`,
	)
	titleRe = regexp.MustCompile(
		`<h3><b>(.+?)</b></h3>`,
	)
)

func Fetch(challengeID string, username string) (*Result, error) {
	url := fmt.Sprintf("%s/challenges/%s", config.Host(), challengeID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "text/html")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d for challenge %s", resp.StatusCode, challengeID)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	html := string(body)

	result := &Result{}

	if m := titleRe.FindStringSubmatch(html); len(m) > 1 {
		result.Stats.Title = m[1]
	}

	statMatches := statsRe.FindAllStringSubmatch(html, 2)
	if len(statMatches) >= 2 {
		result.Stats.Golfers, _ = strconv.Atoi(statMatches[0][1])
		result.Stats.Entries, _ = strconv.Atoi(statMatches[1][1])
	}

	entryMatches := entryRe.FindAllStringSubmatch(html, -1)
	scoreMatches := scoreRe.FindAllStringSubmatch(html, -1)

	for i, m := range entryMatches {
		rank, _ := strconv.Atoi(m[1])
		score := 0
		if i < len(scoreMatches) {
			score, _ = strconv.Atoi(scoreMatches[i][1])
		}

		entry := Entry{
			Rank:     rank,
			Name:     strings.TrimSpace(m[2]),
			Username: m[3],
			Score:    score,
		}
		result.Top = append(result.Top, entry)

		if username != "" && strings.EqualFold(entry.Username, username) {
			e := entry
			result.UserEntry = &e
		}
	}

	if username != "" && result.UserEntry == nil {
		userEntry, err := fetchUserEntry(challengeID, username)
		if err == nil && userEntry != nil {
			result.UserEntry = userEntry
		}
	}

	return result, nil
}

func fetchUserEntry(challengeID, username string) (*Entry, error) {
	url := fmt.Sprintf("%s/challenges/%s/user/%s", config.Host(), challengeID, username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/html")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	html := string(body)

	bestScoreRe := regexp.MustCompile(
		`<b>Best score:</b>\s*(\d+)`,
	)
	rankRe := regexp.MustCompile(
		`<b>Rank:</b>\s*#(\d+)`,
	)

	entry := &Entry{Username: username}

	if m := bestScoreRe.FindStringSubmatch(html); len(m) > 1 {
		entry.Score, _ = strconv.Atoi(m[1])
	} else {
		return nil, nil
	}

	if m := rankRe.FindStringSubmatch(html); len(m) > 1 {
		entry.Rank, _ = strconv.Atoi(m[1])
	}

	return entry, nil
}
