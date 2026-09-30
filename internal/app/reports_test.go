package app

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReportImportStatusIsolation(t *testing.T) {
	a, err := newApp(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	newKey := func() (string, string) {
		t.Helper()
		response := request(a, "POST", "/api/teams/demo-fin/integrations", "4", `{"name":"Accounting export","provider":"1c"}`)
		if response.Code != 201 {
			t.Fatal(response.Code, response.Body.String())
		}
		var key map[string]string
		if err := json.Unmarshal(response.Body.Bytes(), &key); err != nil {
			t.Fatal(err)
		}
		return key["id"], key["token"]
	}
	keyID, token := newKey()
	_, other := newKey()
	call := func(method, path, key, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+key)
		w := httptest.NewRecorder()
		a.routes().ServeHTTP(w, r)
		return w
	}
	channels := call("GET", "/api/integrations/channels", token, "")
	if channels.Code != 200 {
		t.Fatal(channels.Code)
	}
	var available []Channel
	_ = json.Unmarshal(channels.Body.Bytes(), &available)
	if len(available) == 0 {
		t.Fatal("no channels")
	}
	for _, c := range available {
		if c.TeamID != "demo-fin" {
			t.Fatal("channel access leaked")
		}
	}
	importPath := "/api/integrations/documents?channelId=demo-fin-cashflow&title=Cashflow&filename=report.csv&externalId=cf-september-v1&review=1"
	imported := call("POST", importPath, token, "Department,Plan,Actual\nSales,100,95\n")
	if imported.Code != 201 {
		t.Fatal(imported.Code, imported.Body.String())
	}
	var doc Document
	_ = json.Unmarshal(imported.Body.Bytes(), &doc)
	if doc.SourceSystem != "1c" || doc.ImportMethod != "api" || doc.Status != "review" {
		t.Fatal("wrong source metadata")
	}
	statusPath := "/api/integrations/documents?externalId=cf-september-v1"
	if r := call("GET", statusPath, other, ""); r.Code != 404 {
		t.Fatal("another key read report", r.Code)
	}
	decision := request(a, "PATCH", fmt.Sprintf("/api/documents/%s/status", doc.ID), "4", `{"status":"changes","reason":"Correct the figures","revision":1}`)
	if decision.Code != 200 {
		t.Fatal(decision.Body.String())
	}
	status := call("GET", statusPath, token, "")
	var saved Document
	_ = json.Unmarshal(status.Body.Bytes(), &saved)
	if status.Code != 200 || saved.Status != "changes" || saved.Revision != 2 || !strings.Contains(status.Body.String(), "Correct the figures") {
		t.Fatal("decision not available to sender")
	}
	if got := call("POST", importPath, token, "duplicate"); got.Code != 200 {
		t.Fatal("retry not idempotent")
	}
	if r := request(a, "DELETE", "/api/teams/demo-fin/integrations/"+keyID, "4", ""); r.Code != 200 {
		t.Fatal("revoke failed")
	}
	if r := call("GET", statusPath, token, ""); r.Code != 401 {
		t.Fatal("revoked key read report")
	}
}

func TestManualReportKeepsSourceAndRequiresMembership(t *testing.T) {
	a, err := newApp(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]string{"title": "Manual CRM report", "channelId": "demo-sales-pipeline", "sourceSystem": "bitrix24", "review": "1"}
	if response := uploadTest(a, "/api/documents", "2", fields); response.Code != 404 {
		t.Fatal("outsider uploaded report")
	}
	response := uploadTest(a, "/api/documents", "21", fields)
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var doc Document
	_ = json.Unmarshal(response.Body.Bytes(), &doc)
	if doc.SourceSystem != "bitrix24" || doc.ImportMethod != "file" || doc.Status != "review" || doc.Owner.ID != "21" {
		t.Fatal("manual report lost provenance")
	}
	fields["sourceSystem"] = "invalid"
	if response := uploadTest(a, "/api/documents", "21", fields); response.Code != 400 {
		t.Fatal("invalid source accepted")
	}
}
