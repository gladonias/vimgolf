package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/gadonias/vimgolf/cli/internal/challenge"
	"codeberg.org/gadonias/vimgolf/cli/internal/config"
)

func setupTestHome(t *testing.T) (string, func()) {
	t.Helper()
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	return tmpHome, func() { os.Setenv("HOME", origHome) }
}

func TestUploadSuccess(t *testing.T) {
	tmpHome, cleanup := setupTestHome(t)
	defer cleanup()

	cfg := &config.Config{Key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Username: "testuser"}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	logFile := filepath.Join(tmpHome, "keylog")
	os.WriteFile(logFile, []byte("test-keystrokes"), 0o644)

	var receivedID, receivedKey, receivedEntry string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)
		receivedID = r.FormValue("challenge_id")
		receivedKey = r.FormValue("apikey")
		receivedEntry = r.FormValue("entry")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	ch := &challenge.Challenge{ID: "test123", Remote: true, LogPath: logFile}
	if !upload(ch) {
		t.Error("upload() returned false, expected true")
	}
	if receivedID != "test123" {
		t.Errorf("challenge_id = %q, want %q", receivedID, "test123")
	}
	if receivedKey != cfg.Key {
		t.Errorf("apikey = %q, want %q", receivedKey, cfg.Key)
	}
	if receivedEntry != "test-keystrokes" {
		t.Errorf("entry = %q, want %q", receivedEntry, "test-keystrokes")
	}
}

func TestUploadServerReject(t *testing.T) {
	tmpHome, cleanup := setupTestHome(t)
	defer cleanup()

	cfg := &config.Config{Key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	config.Save(cfg)

	logFile := filepath.Join(tmpHome, "keylog")
	os.WriteFile(logFile, []byte("data"), 0o644)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "failed"})
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	ch := &challenge.Challenge{ID: "test123", Remote: true, LogPath: logFile}
	if upload(ch) {
		t.Error("upload() returned true on server rejection, expected false")
	}
}

func TestUploadMissingLogFile(t *testing.T) {
	tmpHome, cleanup := setupTestHome(t)
	defer cleanup()

	cfg := &config.Config{Key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	config.Save(cfg)

	ch := &challenge.Challenge{
		ID:      "test123",
		Remote:  true,
		LogPath: filepath.Join(tmpHome, "nonexistent"),
	}
	if upload(ch) {
		t.Error("upload() returned true with missing log file, expected false")
	}
}

func TestUploadNoConfig(t *testing.T) {
	_, cleanup := setupTestHome(t)
	defer cleanup()

	ch := &challenge.Challenge{ID: "test123", Remote: true, LogPath: "/dev/null"}
	if upload(ch) {
		t.Error("upload() returned true without config, expected false")
	}
}
