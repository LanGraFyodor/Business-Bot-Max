package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestHierarchyRolesAndEscalation(t *testing.T) {
	a, err := newApp(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	var root, sales, studio Team
	for _, team := range a.state.Teams {
		switch team.ID {
		case "orbita":
			root = team
		case "demo-sales":
			sales = team
		case "studio":
			studio = team
		}
	}
	if sales.ParentTeamID != root.ID || studio.ParentTeamID != sales.ID {
		t.Fatal("demo reporting chain is missing")
	}
	if a.teamRole(root.ID, "5") != "employee" || a.teamRole(sales.ID, "5") != "manager" || a.teamRole(sales.ID, "3") != "employee" || a.teamRole(studio.ID, "3") != "manager" {
		t.Fatalf("roles: root5=%s sales5=%s sales3=%s studio3=%s", a.teamRole(root.ID, "5"), a.teamRole(sales.ID, "5"), a.teamRole(sales.ID, "3"), a.teamRole(studio.ID, "3"))
	}
	managerState := a.visibleState(demoUsers["5"])
	chart := managerState["orgChart"].([]OrgTeam)
	visibleStudio := false
	for _, node := range chart {
		if node.ID == "studio" {
			visibleStudio = true
		}
	}
	if !visibleStudio || a.teamRole("studio", "5") != "" {
		t.Fatal("manager should see child metadata without document access")
	}
	employeeState := a.visibleState(demoUsers["2"])
	for _, node := range employeeState["orgChart"].([]OrgTeam) {
		if node.ID == "studio" {
			t.Fatal("employee can see private child metadata")
		}
	}
	if len(sales.Members) < 10 || len(studio.Members) < 10 {
		t.Fatal("teams remain too small")
	}
	if got := request(a, "POST", "/api/teams", "2", `{"name":"Private child","parentTeamId":"demo-sales"}`); got.Code != http.StatusForbidden {
		t.Fatalf("outsider created child: %d", got.Code)
	}
	child := request(a, "POST", "/api/teams", "3", `{"name":"New unit","parentTeamId":"demo-sales"}`)
	if child.Code != http.StatusCreated {
		t.Fatalf("subordinate could not lead a child team: %d %s", child.Code, child.Body.String())
	}
	var created Team
	if err := json.Unmarshal(child.Body.Bytes(), &created); err != nil || created.ParentTeamID != "demo-sales" {
		t.Fatal("child hierarchy not stored")
	}
	var source Document
	for _, d := range a.state.Documents {
		if a.channelTeam(d.ChannelID) == studio.ID && d.Status == "approved" {
			source = d
			break
		}
	}
	if source.ID == "" {
		t.Fatal("no approved child report")
	}
	path := fmt.Sprintf("/api/documents/%s/escalate", source.ID)
	payload := `{"channelId":"demo-sales-pipeline"}`
	if got := request(a, "POST", path, "8", payload); got.Code != http.StatusForbidden {
		t.Fatalf("employee escalated: %d", got.Code)
	}
	if got := request(a, "POST", path, "3", `{"channelId":"demo-fin-invoices"}`); got.Code != http.StatusBadRequest {
		t.Fatalf("wrong destination allowed: %d", got.Code)
	}
	forwarded := request(a, "POST", path, "3", payload)
	if forwarded.Code != http.StatusCreated {
		t.Fatalf("manager escalation failed: %d %s", forwarded.Code, forwarded.Body.String())
	}
	var childDoc Document
	if err := json.Unmarshal(forwarded.Body.Bytes(), &childDoc); err != nil {
		t.Fatal(err)
	}
	if childDoc.EscalatedFrom != source.ID || childDoc.Status != "review" || a.channelTeam(childDoc.ChannelID) != sales.ID {
		t.Fatal("bad forwarded document")
	}
	if got := request(a, "GET", fmt.Sprintf("/api/documents/%s/files/%s", childDoc.ID, childDoc.Versions[0].ID), "2", ""); got.Code != http.StatusNotFound {
		t.Fatalf("outsider read forwarded file: %d", got.Code)
	}
	if got := request(a, "POST", path, "3", payload); got.Code != http.StatusOK {
		t.Fatalf("retry duplicated document: %d", got.Code)
	}
	approve := request(a, "PATCH", fmt.Sprintf("/api/documents/%s/status", childDoc.ID), "5", `{"status":"approved","revision":1}`)
	if approve.Code != http.StatusOK {
		t.Fatalf("parent manager could not approve: %d %s", approve.Code, approve.Body.String())
	}
}

func TestDemoIntegrationProducesReviewDocument(t *testing.T) {
	a, err := newApp(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	got := request(a, "POST", "/api/demo/integrations/sample", "5", `{"provider":"1c","channelId":"demo-fin-invoices"}`)
	if got.Code != http.StatusForbidden {
		t.Fatalf("other team manager imported: %d", got.Code)
	}
	got = request(a, "POST", "/api/demo/integrations/sample", "4", `{"provider":"1c","channelId":"demo-fin-invoices"}`)
	if got.Code != http.StatusCreated {
		t.Fatalf("demo import failed: %d %s", got.Code, got.Body.String())
	}
	var doc Document
	if err := json.Unmarshal(got.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.SourceSystem != "1c" || doc.Status != "review" || len(doc.Versions) != 1 {
		t.Fatalf("bad demo import: %+v", doc)
	}
}
