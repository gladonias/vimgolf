package listing

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"codeberg.org/gadonias/vimgolf/cli/internal/config"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type Challenge struct {
	ID      string
	Title   string
	Entries int
}

var (
	challengeRe = regexp.MustCompile(
		`<a href="/challenges/([^"]+)">([^<]+)</a>\s*-\s*(\d+)\s*entr`,
	)
	lastPageRe = regexp.MustCompile(
		`\?page=(\d+)">Last`,
	)
)

func FetchPage(page int) (challenges []Challenge, totalPages int, err error) {
	url := fmt.Sprintf("%s/?page=%d", config.Host(), page)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "text/html")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch challenges page %d: %w", page, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("server returned %d for challenges page %d", resp.StatusCode, page)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response: %w", err)
	}
	pageHTML := string(body)

	matches := challengeRe.FindAllStringSubmatch(pageHTML, -1)
	for _, m := range matches {
		entries, _ := strconv.Atoi(m[3])
		challenges = append(challenges, Challenge{
			ID:      m[1],
			Title:   html.UnescapeString(m[2]),
			Entries: entries,
		})
	}

	totalPages = 1
	if m := lastPageRe.FindStringSubmatch(pageHTML); len(m) > 1 {
		totalPages, _ = strconv.Atoi(m[1])
	}

	return challenges, totalPages, nil
}

func FetchAll(minEntries int) ([]Challenge, error) {
	first, totalPages, err := FetchPage(1)
	if err != nil {
		return nil, err
	}

	var all []Challenge
	for _, c := range first {
		if c.Entries >= minEntries {
			all = append(all, c)
		}
	}

	for page := 2; page <= totalPages; page++ {
		challenges, _, err := FetchPage(page)
		if err != nil {
			return nil, err
		}
		for _, c := range challenges {
			if c.Entries >= minEntries {
				all = append(all, c)
			}
		}
	}

	return all, nil
}
