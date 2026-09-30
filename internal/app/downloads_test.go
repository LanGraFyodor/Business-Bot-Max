package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDownloadTicketRespectsTeamAndExpiry(t *testing.T) {
	a, err := newApp(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	uploaded := uploadTest(a, "/api/documents", "3", map[string]string{"title": "Studio private", "channelId": "studio-projects"})
	if uploaded.Code != 201 {
		t.Fatal(uploaded.Body.String())
	}
	var d Document
	if err := json.Unmarshal(uploaded.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	linkPath := "/api/documents/" + d.ID + "/files/" + d.Versions[0].ID + "/download-link"
	if w := request(a, "POST", linkPath, "2", ""); w.Code != 404 {
		t.Fatalf("outsider received ticket: %d", w.Code)
	}
	w := request(a, "POST", linkPath, "3", "")
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result.URL, "/api/download/") {
		t.Fatal(result.URL)
	}
	if w = request(a, "GET", result.URL, "", ""); w.Code != 200 || w.Body.Len() == 0 {
		t.Fatalf("valid ticket failed: %d", w.Code)
	}
	key := strings.TrimPrefix(result.URL, "/api/download/")
	ticket := a.downloads[key]
	ticket.Expires = time.Now().Add(-time.Second)
	a.downloads[key] = ticket
	if w = request(a, "GET", result.URL, "", ""); w.Code != 404 {
		t.Fatalf("expired ticket worked: %d", w.Code)
	}
}
