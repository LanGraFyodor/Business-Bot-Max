package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestPollingCursorSurvivesRestart(t *testing.T) {
	var count atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/updates" {
			t.Errorf("unexpected request: %s", r.URL)
			w.WriteHeader(404)
			return
		}
		n := count.Add(1)
		expected := "123"
		if n > 1 {
			expected = "456"
		}
		if r.URL.Query().Get("marker") != expected {
			t.Errorf("poll marker = %q, want %q", r.URL.Query().Get("marker"), expected)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"updates":[],"marker":456}`)
		if n > 1 {
			cancel()
		}
	}))
	defer server.Close()
	t.Setenv("MAX_API_URL", server.URL)
	t.Setenv("MAX_BOT_TOKEN", "test-token")
	dir := t.TempDir()
	a, err := newApp(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	a.state.PollMarker = "123"
	if err := a.save(); err != nil {
		t.Fatal(err)
	}
	a.poll(ctx)
	restored, err := newApp(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() < 2 || restored.state.PollMarker != "456" {
		t.Fatalf("cursor not persisted: calls=%d marker=%s", count.Load(), restored.state.PollMarker)
	}
}
