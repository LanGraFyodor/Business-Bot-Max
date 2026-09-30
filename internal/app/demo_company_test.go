package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestCompanyDemoHasUsableDocumentsAndRoles(t *testing.T) {
	dir := t.TempDir()
	a, err := newApp(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if a.state.DemoSeedVersion != 3 || len(a.state.Teams) < 7 || len(a.state.Channels) < 40 || len(a.state.Documents) < 170 || len(a.state.Messages) < 70 {
		t.Fatalf("small company demo: teams=%d channels=%d documents=%d messages=%d", len(a.state.Teams), len(a.state.Channels), len(a.state.Documents), len(a.state.Messages))
	}
	statuses := map[string]int{}
	formats := map[string]int{}
	for _, d := range a.state.Documents {
		if !a.canRead(d.ChannelID, d.Owner.ID) {
			t.Fatalf("owner cannot access %s", d.Title)
		}
		statuses[d.Status]++
		for _, v := range d.Versions {
			info, err := os.Stat(filepath.Join(dir, "files", v.ID))
			if err != nil || info.Size() != v.Size {
				t.Fatalf("broken file for %s: %v", d.Title, err)
			}
			formats[filepath.Ext(v.Name)]++
		}
	}
	for _, status := range []string{"draft", "review", "changes", "approved"} {
		if statuses[status] < 10 {
			t.Fatalf("too few %s documents: %d", status, statuses[status])
		}
	}
	for _, ext := range []string{".pdf", ".docx", ".xlsx", ".csv"} {
		if formats[ext] < 5 {
			t.Fatalf("too few %s files: %d", ext, formats[ext])
		}
	}
	state := a.visibleState(demoUsers["1"])
	profiles, ok := state["demoProfiles"].([]User)
	if !ok || len(profiles) < 30 {
		t.Fatal("demo employees not switchable")
	}
	anna := a.visibleState(demoUsers["2"])
	if len(anna["teams"].([]Team)) != 1 {
		t.Fatal("Anna gained access to another team")
	}
	var finance Document
	for _, d := range a.state.Documents {
		if a.channelTeam(d.ChannelID) == "demo-fin" && d.Status == "review" {
			finance = d
			break
		}
	}
	if finance.ID == "" {
		t.Fatal("no finance approval example")
	}
	path := fmt.Sprintf("/api/documents/%s/files/%s", finance.ID, finance.Versions[0].ID)
	if got := request(a, "GET", path, "2", ""); got.Code != http.StatusNotFound {
		t.Fatalf("Anna read finance file: %d", got.Code)
	}
	if got := request(a, "GET", path, "4", ""); got.Code != http.StatusOK {
		t.Fatalf("finance manager could not download: %d", got.Code)
	}
	decision := fmt.Sprintf("/api/documents/%s/status", finance.ID)
	payload := fmt.Sprintf(`{"status":"approved","revision":%d}`, finance.Revision)
	if got := request(a, "PATCH", decision, "8", payload); got.Code != http.StatusForbidden {
		t.Fatalf("employee approved: %d", got.Code)
	}
	if got := request(a, "PATCH", decision, "4", payload); got.Code != http.StatusOK {
		t.Fatalf("manager could not approve: %d %s", got.Code, got.Body.String())
	}
	uploaded := uploadTest(a, "/api/documents", "2", map[string]string{"title": "User work stays", "channelId": "legal"})
	if uploaded.Code != http.StatusCreated {
		t.Fatalf("user upload failed: %d", uploaded.Code)
	}
	before := len(a.state.Documents)
	reopened, err := newApp(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reopened.state.Documents) != before {
		t.Fatalf("demo duplicated on restart: %d vs %d", len(reopened.state.Documents), before)
	}
	var saved Document
	for _, d := range reopened.state.Documents {
		if d.ID == finance.ID {
			saved = d
		}
	}
	if saved.Status != "approved" {
		t.Fatal("user decision lost on restart")
	}
	var userDoc Document
	if err := json.Unmarshal(uploaded.Body.Bytes(), &userDoc); err != nil {
		t.Fatal(err)
	}
	if !reopened.canRead(userDoc.ChannelID, "2") {
		t.Fatal("user upload lost")
	}
}
