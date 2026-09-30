package app

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func liveRequest(a *App, method, path, userID, name, body string) *httptest.ResponseRecorder {
	raw := signed("one-bot-token", url.Values{
		"user":      {fmt.Sprintf(`{"id":%s,"first_name":%q}`, userID, name)},
		"auth_date": {strconv.FormatInt(time.Now().Unix(), 10)},
	})
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("X-Max-Init-Data", raw)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	return w
}

func TestOneMAXBotMultipleVerifiedAccountsAndRoleInvitations(t *testing.T) {
	t.Setenv("MAX_BOT_TOKEN", "one-bot-token")
	dir := t.TempDir()
	a, err := newApp(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.state.Teams) != 0 {
		t.Fatal("new live deployment should have no demo workspace")
	}
	create := liveRequest(a, "POST", "/api/teams", "101", "Boss", `{"name":"Work","color":"blue"}`)
	if create.Code != 201 {
		t.Fatal(create.Body.String())
	}
	var team Team
	if err := json.Unmarshal(create.Body.Bytes(), &team); err != nil {
		t.Fatal(err)
	}
	if team.Members[0].ID != "101" || team.Members[0].Role != "manager" {
		t.Fatal("creator is not manager")
	}
	outsider := liveRequest(a, "GET", "/api/state", "202", "Employee", "")
	if outsider.Code != 200 || strings.Contains(outsider.Body.String(), team.ID) {
		t.Fatal("private team exposed")
	}
	targeted := liveRequest(a, "POST", "/api/teams/"+team.ID+"/invites", "101", "Boss", `{"role":"employee","targetUserId":"202"}`)
	if targeted.Code != 201 {
		t.Fatal(targeted.Body.String())
	}
	var invite Invitation
	if err := json.Unmarshal(targeted.Body.Bytes(), &invite); err != nil {
		t.Fatal(err)
	}
	if invite.Role != "employee" || invite.TargetUserID != "202" {
		t.Fatal("invite attributes wrong")
	}
	path := "/api/teams/join"
	body := fmt.Sprintf(`{"code":%q}`, invite.Code)
	if w := liveRequest(a, "POST", path, "303", "Other", body); w.Code != 403 {
		t.Fatalf("other account accepted: %d", w.Code)
	}
	if w := liveRequest(a, "GET", "/api/invites/"+invite.Code, "303", "Other", ""); w.Code != 403 {
		t.Fatal("targeted invite preview leaked")
	}
	if w := liveRequest(a, "POST", path, "202", "Employee", body); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if a.teamRole(team.ID, "202") != "employee" {
		t.Fatal("wrong employee role")
	}
	if w := liveRequest(a, "POST", path, "303", "Other", body); w.Code != 404 {
		t.Fatal("one-use invite was reused")
	}
	if w := liveRequest(a, "POST", "/api/channels", "202", "Employee", fmt.Sprintf(`{"teamId":%q,"name":"Secret"}`, team.ID)); w.Code != 403 {
		t.Fatal("employee created channel")
	}
	elevated := liveRequest(a, "POST", "/api/teams/"+team.ID+"/invites", "101", "Boss", `{"role":"manager","targetUserId":"303"}`)
	if elevated.Code != 201 {
		t.Fatal(elevated.Body.String())
	}
	if err := json.Unmarshal(elevated.Body.Bytes(), &invite); err != nil {
		t.Fatal(err)
	}
	if w := liveRequest(a, "POST", path, "303", "Other", fmt.Sprintf(`{"code":%q}`, invite.Code)); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := liveRequest(a, "POST", "/api/channels", "303", "Other", fmt.Sprintf(`{"teamId":%q,"name":"Manager channel"}`, team.ID)); w.Code != 201 {
		t.Fatal("invited manager lacked permission", w.Body.String())
	}
	state := liveRequest(a, "GET", "/api/state", "202", "Renamed", "")
	if state.Code != 200 || strings.Contains(state.Body.String(), invite.Code) {
		t.Fatal("employee saw manager invite")
	}
	if a.state.Accounts[1].Name != "Renamed" {
		t.Fatal("MAX profile not updated")
	}
	restored, err := newApp(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if restored.teamRole(team.ID, "202") != "employee" || restored.teamRole(team.ID, "303") != "manager" || len(restored.state.Accounts) != 3 {
		t.Fatal("MAX identities and roles not persisted")
	}
	removePath := "/api/teams/" + team.ID + "/members/202"
	if w := liveRequest(restored, "DELETE", removePath, "202", "Renamed", ""); w.Code != 403 {
		t.Fatal("employee removed themselves")
	}
	if w := liveRequest(restored, "DELETE", removePath, "101", "Boss", ""); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if restored.teamRole(team.ID, "202") != "" {
		t.Fatal("revoked member retained access")
	}
	if w := liveRequest(restored, "GET", "/api/state", "202", "Renamed", ""); strings.Contains(w.Body.String(), team.ID) {
		t.Fatal("revoked team remains visible")
	}
	pending := liveRequest(restored, "POST", "/api/teams/"+team.ID+"/invites", "303", "Other", `{"role":"employee","targetUserId":"404"}`)
	if pending.Code != 201 {
		t.Fatal(pending.Body.String())
	}
	var revoked Invitation
	if err := json.Unmarshal(pending.Body.Bytes(), &revoked); err != nil {
		t.Fatal(err)
	}
	if w := liveRequest(restored, "DELETE", "/api/teams/"+team.ID+"/invites/"+revoked.ID, "101", "Boss", ""); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := liveRequest(restored, "POST", "/api/teams/join", "404", "New", fmt.Sprintf(`{"code":%q}`, revoked.Code)); w.Code != 404 {
		t.Fatal("revoked invite was accepted")
	}
}
