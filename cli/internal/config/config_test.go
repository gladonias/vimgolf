package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"abcdef1234567890abcdef1234567890", true},
		{"ABCDEF1234567890abcdef1234567890", true},
		{"abcdef_234567890abcdef1234567890", true},  // underscores allowed
		{"abc", false},                               // too short
		{"abcdef1234567890abcdef1234567890a", false},  // too long
		{"abcdef1234567890abcdef123456789!", false},   // special char
		{"", false},
	}
	for _, tt := range tests {
		got := ValidateKey(tt.key)
		if got != tt.want {
			t.Errorf("ValidateKey(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestHostDefault(t *testing.T) {
	os.Unsetenv("GOLFHOST")
	if got := Host(); got != DefaultHost {
		t.Errorf("Host() = %q, want %q", got, DefaultHost)
	}
}

func TestHostEnvOverride(t *testing.T) {
	os.Setenv("GOLFHOST", "http://localhost:3000")
	defer os.Unsetenv("GOLFHOST")
	if got := Host(); got != "http://localhost:3000" {
		t.Errorf("Host() = %q, want %q", got, "http://localhost:3000")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{Key: "abcdef1234567890abcdef1234567890", Username: "testuser"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, ".vimgolf", "config.yaml")); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Key != cfg.Key {
		t.Errorf("loaded Key = %q, want %q", loaded.Key, cfg.Key)
	}
	if loaded.Username != cfg.Username {
		t.Errorf("loaded Username = %q, want %q", loaded.Username, cfg.Username)
	}
}

func TestLoadMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	_, err := Load()
	if err == nil {
		t.Error("Load() should fail when config file is missing")
	}
}

func TestLoadEmptyKey(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	if err := Save(&Config{Key: ""}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() should fail when key is empty")
	}
}
