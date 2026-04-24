package leaderboard

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const sampleHTML = `
<html>
<h3><b>Test Challenge</b></h3>
<h2><b class="stat">500</b> active golfers, <b class="stat">2000</b> entries</h2>

<div style="float:right; text-align:center;font-size:2.4em"><b><a href="/challenges/abc/user/alice">10</a></b></div>
<h6><a class="anchor" href="#"><b>#1</b></a> - Alice Smith / <a href="/1/alice">@alice</a></h6>

<div style="float:right; text-align:center;font-size:2.4em"><b><a href="/challenges/abc/user/bob">12</a></b></div>
<h6><a class="anchor" href="#"><b>#2</b></a> - Bob Jones / <a href="/2/bob">@bob</a></h6>

<div style="float:right; text-align:center;font-size:2.4em"><b><a href="/challenges/abc/user/charlie">15</a></b></div>
<h6><a class="anchor" href="#"><b>#3</b></a> - Charlie Brown / <a href="/3/charlie">@charlie</a></h6>
</html>
`

func TestParseLeaderboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(sampleHTML))
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	result, err := Fetch("abc", "")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if result.Stats.Title != "Test Challenge" {
		t.Errorf("title = %q, want %q", result.Stats.Title, "Test Challenge")
	}
	if result.Stats.Golfers != 500 {
		t.Errorf("golfers = %d, want 500", result.Stats.Golfers)
	}
	if result.Stats.Entries != 2000 {
		t.Errorf("entries = %d, want 2000", result.Stats.Entries)
	}

	if len(result.Top) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Top))
	}

	tests := []struct {
		rank     int
		score    int
		name     string
		username string
	}{
		{1, 10, "Alice Smith", "alice"},
		{2, 12, "Bob Jones", "bob"},
		{3, 15, "Charlie Brown", "charlie"},
	}

	for i, tt := range tests {
		e := result.Top[i]
		if e.Rank != tt.rank {
			t.Errorf("entry %d rank = %d, want %d", i, e.Rank, tt.rank)
		}
		if e.Score != tt.score {
			t.Errorf("entry %d score = %d, want %d", i, e.Score, tt.score)
		}
		if e.Name != tt.name {
			t.Errorf("entry %d name = %q, want %q", i, e.Name, tt.name)
		}
		if e.Username != tt.username {
			t.Errorf("entry %d username = %q, want %q", i, e.Username, tt.username)
		}
	}
}

func TestFetchWithUsername(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(sampleHTML))
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	result, err := Fetch("abc", "bob")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if result.UserEntry == nil {
		t.Fatal("expected UserEntry to be set for bob")
	}
	if result.UserEntry.Rank != 2 {
		t.Errorf("user rank = %d, want 2", result.UserEntry.Rank)
	}
	if result.UserEntry.Score != 12 {
		t.Errorf("user score = %d, want 12", result.UserEntry.Score)
	}
}

func TestFetch404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	_, err := Fetch("nonexistent", "")
	if err == nil {
		t.Error("Fetch() should fail on 404")
	}
}

func TestFetchEmptyLeaderboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><h3><b>Empty Challenge</b></h3></html>"))
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	result, err := Fetch("empty", "")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if len(result.Top) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Top))
	}
	if result.Stats.Title != "Empty Challenge" {
		t.Errorf("title = %q, want %q", result.Stats.Title, "Empty Challenge")
	}
}
