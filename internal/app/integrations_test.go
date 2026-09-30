package app

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationImportAndRevocation(t *testing.T) {
	dir := t.TempDir()
	a, err := newApp(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := request(a, "POST", "/api/teams/orbita/integrations", "2", `{"name":"1C","provider":"1c"}`); got.Code != 403 {
		t.Fatalf("employee created key: %d", got.Code)
	}
	created := request(a, "POST", "/api/teams/orbita/integrations", "1", `{"name":"1C Accounting","provider":"1c"}`)
	if created.Code != 201 {
		t.Fatal(created.Code, created.Body.String())
	}
	var key struct {
		ID    string `json:"id"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &key); err != nil {
		t.Fatal(err)
	}
	if key.ID == "" || key.Token == "" {
		t.Fatal("missing key")
	}
	stored, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(stored), key.Token) {
		t.Fatal("plaintext key stored")
	}
	state := request(a, "GET", "/api/state", "1", "")
	if strings.Contains(state.Body.String(), key.Token) || strings.Contains(state.Body.String(), a.state.Integrations[0].Hash) {
		t.Fatal("key exposed to UI state")
	}
	initialDocuments := len(a.state.Documents)
	createURL := "/api/integrations/documents?channelId=finance&title=Invoice&filename=invoice.pdf&externalId=invoice-9&review=1"
	send := func(token, target string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", target, strings.NewReader("%PDF-1.7 sample"))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		a.routes().ServeHTTP(w, r)
		return w
	}
	if got := send(key.Token, "/api/integrations/documents?channelId=studio-general&title=Wrong&filename=x.pdf"); got.Code != 404 {
		t.Fatalf("cross-team import: %d", got.Code)
	}
	if got := send("dsi_invalid", createURL); got.Code != 401 {
		t.Fatalf("invalid token: %d", got.Code)
	}
	first := send(key.Token, createURL)
	if first.Code != 201 {
		t.Fatal(first.Code, first.Body.String())
	}
	var doc Document
	if err := json.Unmarshal(first.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Status != "review" || doc.ChannelID != "finance" || doc.ExternalID != "invoice-9" || len(doc.Versions) != 1 {
		t.Fatalf("bad import: %+v", doc)
	}

	if got := send(key.Token, createURL); got.Code != 200 {
		t.Fatalf("duplicate: %d", got.Code)
	}
	if len(a.state.Documents) != initialDocuments+1 {
		t.Fatalf("duplicate created document: %d", len(a.state.Documents))
	}
	file, err := os.ReadFile(filepath.Join(dir, "files", doc.Versions[0].ID))
	if err != nil || string(file) != "%PDF-1.7 sample" {
		t.Fatal("imported file not stored", err)
	}
	restored, err := newApp(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.state.Integrations) != 1 {
		t.Fatal("key not persisted")
	}
	if len(restored.state.Documents) != len(a.state.Documents) {
		t.Fatal("document not persisted")
	}
	path := "/api/teams/orbita/integrations/" + url.PathEscape(key.ID)
	if got := request(a, "DELETE", path, "1", ""); got.Code != 200 {
		t.Fatal(got.Code, got.Body.String())
	}
	if got := send(key.Token, createURL); got.Code != 401 {
		t.Fatalf("revoked key accepted: %d", got.Code)
	}
}
