package challenge

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"codeberg.org/gadonias/vimgolf/cli/internal/config"
	"codeberg.org/gadonias/vimgolf/cli/internal/ui"
)

//go:embed vimgolf.vimrc
var embeddedVimrc []byte

var typeSanitizer = regexp.MustCompile(`[^\w-]`)

type challengeJSON struct {
	In     fileData `json:"in"`
	Out    fileData `json:"out"`
	Client string   `json:"client"`
}

type fileData struct {
	Data string `json:"data"`
	Type string `json:"type"`
}

type Challenge struct {
	ID         string
	Remote     bool
	InputPath  string
	OutputPath string
	WorkPath   string
	LogPath    string
	VimrcPath  string

	putDir string
}

func New(id string, putDir string) *Challenge {
	return &Challenge{
		ID:     id,
		putDir: putDir,
	}
}

func (c *Challenge) Download() error {
	c.Remote = true

	url := fmt.Sprintf("%s/challenges/%s.json", config.Host(), c.ID)
	ui.Debug(fmt.Sprintf("GET %s", url))

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download challenge: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d for challenge %s", resp.StatusCode, c.ID)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var data challengeJSON
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to parse challenge JSON: %w", err)
	}

	data.In.Data = strings.ReplaceAll(data.In.Data, "\r\n", "\n")
	data.Out.Data = strings.ReplaceAll(data.Out.Data, "\r\n", "\n")

	inType := typeSanitizer.ReplaceAllString(data.In.Type, ".")
	outType := typeSanitizer.ReplaceAllString(data.Out.Type, ".")

	c.InputPath = filepath.Join(c.putDir, c.ID+".input."+inType)
	c.OutputPath = filepath.Join(c.putDir, c.ID+".output."+outType)

	if err := os.WriteFile(c.InputPath, []byte(data.In.Data+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to write input file: %w", err)
	}
	if err := os.WriteFile(c.OutputPath, []byte(data.Out.Data+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	if err := c.setupWorkFiles(inType); err != nil {
		return err
	}

	return nil
}

func (c *Challenge) SetupLocal(infile, outfile string) error {
	c.Remote = false

	absIn, err := filepath.Abs(infile)
	if err != nil {
		return err
	}
	absOut, err := filepath.Abs(outfile)
	if err != nil {
		return err
	}

	c.InputPath = absIn
	c.OutputPath = absOut

	ext := filepath.Ext(infile)
	if ext != "" {
		ext = ext[1:]
	}

	return c.setupWorkFiles(ext)
}

func (c *Challenge) setupWorkFiles(ext string) error {
	suffix := ""
	if ext != "" {
		suffix = "." + ext
	}

	workFile, err := os.CreateTemp("", "vimgolf*"+suffix)
	if err != nil {
		return fmt.Errorf("failed to create work file: %w", err)
	}
	workFile.Close()
	c.WorkPath = workFile.Name()

	logFile, err := os.CreateTemp("", "golflog*")
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	logFile.Close()
	c.LogPath = logFile.Name()

	vimrcFile, err := os.CreateTemp("", "vimgolf-vimrc*")
	if err != nil {
		return fmt.Errorf("failed to create vimrc file: %w", err)
	}
	if _, err := vimrcFile.Write(embeddedVimrc); err != nil {
		vimrcFile.Close()
		return fmt.Errorf("failed to write vimrc: %w", err)
	}
	vimrcFile.Close()
	c.VimrcPath = vimrcFile.Name()

	return nil
}

func (c *Challenge) Start() error {
	data, err := os.ReadFile(c.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	return os.WriteFile(c.WorkPath, data, 0o644)
}

func (c *Challenge) Correct() (bool, error) {
	work, err := os.ReadFile(c.WorkPath)
	if err != nil {
		return false, err
	}
	out, err := os.ReadFile(c.OutputPath)
	if err != nil {
		return false, err
	}
	return string(work) == string(out), nil
}

func (c *Challenge) Cleanup() {
	if c.WorkPath != "" {
		os.Remove(c.WorkPath)
	}
	if c.LogPath != "" {
		os.Remove(c.LogPath)
	}
	if c.VimrcPath != "" {
		os.Remove(c.VimrcPath)
	}
}
