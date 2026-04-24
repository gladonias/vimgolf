package challenge

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadAndValidate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := challengeJSON{
			In:     fileData{Data: "hello\n", Type: "txt"},
			Out:    fileData{Data: "HELLO\n", Type: "txt"},
			Client: ClientVersion,
		}
		json.NewEncoder(w).Encode(data)
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	putDir := t.TempDir()
	ch := New("test123", putDir)

	if err := ch.Download(); err != nil {
		t.Fatalf("Download() error: %v", err)
	}

	if !ch.Remote {
		t.Error("expected Remote to be true")
	}

	input, err := os.ReadFile(ch.InputPath)
	if err != nil {
		t.Fatalf("failed to read input: %v", err)
	}
	if string(input) != "hello\n\n" {
		t.Errorf("input = %q, want %q", input, "hello\n\n")
	}
}

func TestDownloadVersionMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := challengeJSON{
			In:     fileData{Data: "hello", Type: "txt"},
			Out:    fileData{Data: "world", Type: "txt"},
			Client: "9.9.9",
		}
		json.NewEncoder(w).Encode(data)
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	ch := New("test123", t.TempDir())
	err := ch.Download()
	if err == nil {
		t.Fatal("Download() should fail on version mismatch")
	}
	if got := err.Error(); got == "" {
		t.Error("error message should not be empty")
	}
}

func TestDownload404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	ch := New("nonexistent", t.TempDir())
	err := ch.Download()
	if err == nil {
		t.Fatal("Download() should fail on 404")
	}
}

func TestSetupLocalAndCorrect(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.txt")
	outPath := filepath.Join(dir, "out.txt")
	os.WriteFile(inPath, []byte("hello"), 0o644)
	os.WriteFile(outPath, []byte("hello"), 0o644)

	ch := New("local-test", "")
	if err := ch.SetupLocal(inPath, outPath); err != nil {
		t.Fatalf("SetupLocal() error: %v", err)
	}
	defer ch.Cleanup()

	if ch.Remote {
		t.Error("expected Remote to be false")
	}

	if err := ch.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	correct, err := ch.Correct()
	if err != nil {
		t.Fatalf("Correct() error: %v", err)
	}
	if !correct {
		t.Error("expected Correct() to be true when files match")
	}
}

func TestCorrectMismatch(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.txt")
	outPath := filepath.Join(dir, "out.txt")
	os.WriteFile(inPath, []byte("hello"), 0o644)
	os.WriteFile(outPath, []byte("world"), 0o644)

	ch := New("mismatch-test", "")
	if err := ch.SetupLocal(inPath, outPath); err != nil {
		t.Fatalf("SetupLocal() error: %v", err)
	}
	defer ch.Cleanup()

	if err := ch.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	correct, err := ch.Correct()
	if err != nil {
		t.Fatalf("Correct() error: %v", err)
	}
	if correct {
		t.Error("expected Correct() to be false when files differ")
	}
}

func TestCleanupRemovesFiles(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.txt")
	outPath := filepath.Join(dir, "out.txt")
	os.WriteFile(inPath, []byte("test"), 0o644)
	os.WriteFile(outPath, []byte("test"), 0o644)

	ch := New("cleanup-test", "")
	if err := ch.SetupLocal(inPath, outPath); err != nil {
		t.Fatalf("SetupLocal() error: %v", err)
	}

	workPath := ch.WorkPath
	logPath := ch.LogPath
	vimrcPath := ch.VimrcPath

	for _, p := range []string{workPath, logPath, vimrcPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("temp file %s should exist before cleanup", p)
		}
	}

	ch.Cleanup()

	for _, p := range []string{workPath, logPath, vimrcPath} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("temp file %s should be removed after cleanup", p)
		}
	}
}
