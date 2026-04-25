package listing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func makePage(challenges []Challenge, totalPages int, currentPage int) string {
	html := "<html>\n"
	for _, c := range challenges {
		html += fmt.Sprintf(`<h5 class="challenge">
  <a href="/challenges/%s">%s</a> - %d entries
</h5>
`, c.ID, c.Title, c.Entries)
	}
	if currentPage < totalPages {
		html += fmt.Sprintf(`<a href="/?page=%d">Last &raquo;</a>`, totalPages)
	}
	html += "</html>"
	return html
}

func TestFetchPage(t *testing.T) {
	challenges := []Challenge{
		{ID: "abc123", Title: "Challenge One", Entries: 500},
		{ID: "def456", Title: "Challenge Two", Entries: 50},
		{ID: "ghi789", Title: "Challenge Three", Entries: 200},
	}
	html := makePage(challenges, 3, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	result, totalPages, err := FetchPage(1)
	if err != nil {
		t.Fatalf("FetchPage() error: %v", err)
	}

	if totalPages != 3 {
		t.Errorf("totalPages = %d, want 3", totalPages)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 challenges, got %d", len(result))
	}

	for i, want := range challenges {
		got := result[i]
		if got.ID != want.ID {
			t.Errorf("challenge %d ID = %q, want %q", i, got.ID, want.ID)
		}
		if got.Title != want.Title {
			t.Errorf("challenge %d Title = %q, want %q", i, got.Title, want.Title)
		}
		if got.Entries != want.Entries {
			t.Errorf("challenge %d Entries = %d, want %d", i, got.Entries, want.Entries)
		}
	}
}

func TestFetchPageLastPage(t *testing.T) {
	html := makePage([]Challenge{{ID: "x", Title: "X", Entries: 10}}, 3, 3)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	_, totalPages, err := FetchPage(3)
	if err != nil {
		t.Fatalf("FetchPage() error: %v", err)
	}

	if totalPages != 1 {
		t.Errorf("totalPages = %d, want 1 (no Last link on final page)", totalPages)
	}
}

func TestFetchAll(t *testing.T) {
	page1 := []Challenge{
		{ID: "a", Title: "Big", Entries: 500},
		{ID: "b", Title: "Small", Entries: 30},
	}
	page2 := []Challenge{
		{ID: "c", Title: "Medium", Entries: 150},
		{ID: "d", Title: "Tiny", Entries: 5},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			w.Write([]byte(makePage(page1, 2, 1)))
		case "2":
			w.Write([]byte(makePage(page2, 2, 2)))
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	result, err := FetchAll(100)
	if err != nil {
		t.Fatalf("FetchAll() error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 challenges with >= 100 entries, got %d", len(result))
	}

	if result[0].ID != "a" {
		t.Errorf("first challenge ID = %q, want %q", result[0].ID, "a")
	}
	if result[1].ID != "c" {
		t.Errorf("second challenge ID = %q, want %q", result[1].ID, "c")
	}
}

func TestFetchAllNoResults(t *testing.T) {
	html := makePage([]Challenge{{ID: "a", Title: "Low", Entries: 10}}, 1, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	result, err := FetchAll(100)
	if err != nil {
		t.Fatalf("FetchAll() error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 challenges, got %d", len(result))
	}
}

func TestFetchPageHTMLEntities(t *testing.T) {
	raw := `<html><h5 class="challenge">
  <a href="/challenges/ent1">It&#39;s a &amp; test &lt;challenge&gt;</a> - 42 entries
</h5></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(raw))
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	result, _, err := FetchPage(1)
	if err != nil {
		t.Fatalf("FetchPage() error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 challenge, got %d", len(result))
	}
	want := "It's a & test <challenge>"
	if result[0].Title != want {
		t.Errorf("Title = %q, want %q", result[0].Title, want)
	}
}

func TestFetchPage404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	os.Setenv("GOLFHOST", server.URL)
	defer os.Unsetenv("GOLFHOST")

	_, _, err := FetchPage(1)
	if err == nil {
		t.Error("FetchPage() should fail on 404")
	}
}
