package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestReviewCompanySharedRoleCodesAndExpiry(t *testing.T) {
	t.Setenv("MAX_BOT_TOKEN", "one-bot-token")
	a, err := newApp(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	until := time.Now().Add(31 * 24 * time.Hour).UTC().Format(time.RFC3339)
	a.state.Teams = []Team{
		{ID: "review-root", Name: "Review", ReviewUntil: until, Members: []User{}, Invites: []Invitation{
			{ID: "manager-pass", Code: "manager-code", Role: "manager", Expires: until, MaxUses: 100, WorkspaceTeamIDs: []string{"review-root", "review-child"}},
			{ID: "employee-pass", Code: "employee-code", Role: "employee", Expires: until, MaxUses: 100, WorkspaceTeamIDs: []string{"review-root", "review-child"}},
		}},
		{ID: "review-child", Name: "Child", ParentTeamID: "review-root", ReviewUntil: until, Members: []User{}},
		{ID: "unrelated", Name: "Private", Members: []User{{ID: "999", Name: "Owner", Role: "manager"}}},
	}
	a.state.Channels = []Channel{{ID: "review-channel", TeamID: "review-root", Name: "reports"}}
	for _, uid := range []string{"101", "202"} {
		w := liveRequest(a, "POST", "/api/teams/join", uid, "Manager", `{"code":"manager-code"}`)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		for _, id := range []string{"review-root", "review-child"} {
			if a.teamRole(id, uid) != "manager" {
				t.Fatal("group role missing")
			}
		}
	}
	var update Update
	update.Type = "message_created"
	update.Message.Sender.ID = 303
	update.Message.Sender.FirstName = "Employee"
	if reply, handled := a.botTeamCommand(update, "303", "/join employee-code"); !handled || !strings.Contains(reply, "2") {
		t.Fatal(reply)
	}
	if a.teamRole("review-child", "303") != "employee" || a.teamRole("unrelated", "101") != "" {
		t.Fatal("scope or role broken")
	}
	if w := liveRequest(a, "POST", "/api/channels", "303", "Employee", `{"teamId":"review-root","name":"No"}`); w.Code != 403 {
		t.Fatal("employee gained manager rights")
	}
	if w := liveRequest(a, "POST", "/api/teams/join", "101", "Manager", `{"code":"manager-code"}`); w.Code != 409 {
		t.Fatal("duplicate join accepted")
	}
	if a.state.Teams[0].Invites[0].UseCount != 2 {
		t.Fatal("duplicate consumed capacity")
	}
	if w := request(a, "GET", "/api/state", "1", ""); w.Code != http.StatusUnauthorized {
		t.Fatal("demo impersonation enabled")
	}
	keyResponse := liveRequest(a, "POST", "/api/teams/review-root/integrations", "101", "Manager", `{"name":"Review import","provider":"1c"}`)
	if keyResponse.Code != 201 {
		t.Fatal(keyResponse.Code, keyResponse.Body.String())
	}
	var key struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(keyResponse.Body.Bytes(), &key); err != nil {
		t.Fatal(err)
	}
	if a.integrationKey("Bearer "+key.Token) == nil {
		t.Fatal("active key rejected")
	}
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	a.state.Teams[0].ReviewUntil = past
	a.state.Teams[1].ReviewUntil = past
	if a.teamRole("review-root", "101") != "" || a.canRead("review-channel", "303") {
		t.Fatal("expired access still works")
	}
	if a.integrationKey("Bearer "+key.Token) != nil {
		t.Fatal("expired review integration still works")
	}
	if w := liveRequest(a, "POST", "/api/teams/join", "404", "Later", `{"code":"employee-code"}`); w.Code == 200 {
		t.Fatal("expired company joined")
	}
	w := liveRequest(a, "GET", "/api/state", "101", "Manager", "")
	if w.Code != 200 || strings.Contains(w.Body.String(), "\"id\":\"review-root\"") {
		t.Fatal("expired company exposed")
	}
}

func TestSharedReviewInviteCanBeRevokedAfterUse(t *testing.T) {
	t.Setenv("MAX_BOT_TOKEN", "one-bot-token")
	a, err := newApp(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	until := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	a.state.Teams = []Team{{ID: "t", Name: "Review", ReviewUntil: until, Members: []User{{ID: "101", Name: "Boss", Role: "manager"}}, Invites: []Invitation{{ID: "i", Code: "shared", Role: "employee", Expires: until, MaxUses: 100}}}}
	if w := liveRequest(a, "POST", "/api/teams/join", "202", "First", `{"code":"shared"}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := liveRequest(a, "DELETE", "/api/teams/t/invites/i", "101", "Boss", ""); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := liveRequest(a, "POST", "/api/teams/join", "303", "Next", `{"code":"shared"}`); w.Code != 404 {
		t.Fatal(fmt.Sprint(w.Code), w.Body.String())
	}
}
