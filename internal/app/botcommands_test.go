package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestMAXBotCommandsShareOneAccountBasedWorkspace(t *testing.T) {
	var sent string
	var buttonPayloads []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !strings.HasPrefix(r.URL.Path, "/messages") {
			t.Errorf("unexpected MAX request: %s", r.URL)
		}
		var body struct {
			Text        string `json:"text"`
			Attachments []struct {
				Payload struct {
					Buttons [][]struct {
						Type    string `json:"type"`
						WebApp  string `json:"web_app"`
						Payload string `json:"payload"`
					} `json:"buttons"`
				} `json:"payload"`
			} `json:"attachments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		sent = body.Text
		buttonPayloads = nil
		if len(body.Attachments) > 0 {
			for _, row := range body.Attachments[0].Payload.Buttons {
				for _, button := range row {
					if button.Type == "open_app" && button.WebApp == "t68_hakaton_max_bot" {
						buttonPayloads = append(buttonPayloads, button.Payload)
					}
				}
			}
		}
		w.WriteHeader(200)
	}))
	defer server.Close()
	t.Setenv("MAX_API_URL", server.URL)
	t.Setenv("MAX_BOT_TOKEN", "one-bot-token")
	t.Setenv("PUBLIC_APP_URL", "https://example.test")
	a, err := newApp(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	talk := func(id int64, name, text string) string {
		t.Helper()
		var update Update
		update.Type = "message_created"
		update.Message.Sender.ID = id
		update.Message.Sender.FirstName = name
		update.Message.Body.Text = text
		if err := a.handleUpdate(context.Background(), update); err != nil {
			t.Fatal(err)
		}
		return sent
	}
	created := talk(101, "Boss", "/create One team")
	if !strings.Contains(created, "создана") || len(a.state.Teams) != 1 {
		t.Fatal(created)
	}
	if len(buttonPayloads) != 2 || buttonPayloads[0] != "docspace" || buttonPayloads[1] != "guide" {
		t.Fatalf("MAX app buttons: %v", buttonPayloads)
	}
	team := a.state.Teams[0]
	if a.teamRole(team.ID, "101") != "manager" {
		t.Fatal("bot creator lacks manager role")
	}
	if text := talk(202, "Member", "/teams"); !strings.Contains(text, "пока не состоите") {
		t.Fatal(text)
	}
	inviteText := talk(101, "Boss", "/invite "+team.ID+" employee 202")
	code := regexp.MustCompile(`Код: ([a-f0-9]{24})`).FindStringSubmatch(inviteText)
	if len(code) != 2 {
		t.Fatal(inviteText)
	}
	if text := talk(303, "Wrong", "/join "+code[1]); !strings.Contains(text, "другому аккаунту") {
		t.Fatal(text)
	}
	if text := talk(202, "Member", "/join "+code[1]); !strings.Contains(text, "как участник") {
		t.Fatal(text)
	}
	if a.teamRole(team.ID, "202") != "employee" || a.teamRole(team.ID, "303") != "" {
		t.Fatal("bot role boundaries failed")
	}
	if w := liveRequest(a, "GET", "/api/state", "202", "Member", ""); w.Code != 200 || !strings.Contains(w.Body.String(), team.ID) {
		t.Fatal("bot membership not visible to signed MAX mini-app", w.Body.String())
	}
	if text := talk(303, "Wrong", "/join "+code[1]); !strings.Contains(text, "уже использовано") {
		t.Fatal(text)
	}
	if text := talk(202, "Member", "/teams"); !strings.Contains(text, "участник") {
		t.Fatal(text)
	}
	if text := talk(202, "Member", "/remove "+team.ID+" 101"); !strings.Contains(text, "только руководитель") {
		t.Fatal(text)
	}
	if text := talk(101, "Boss", "/role "+team.ID+" 202 manager"); !strings.Contains(text, "обновлена") {
		t.Fatal(text)
	}
	if a.teamRole(team.ID, "202") != "manager" {
		t.Fatal("bot role change failed")
	}
	if text := talk(101, "Boss", "/remove "+team.ID+" 202"); !strings.Contains(text, "исключён") {
		t.Fatal(text)
	}
	if a.teamRole(team.ID, "202") != "" {
		t.Fatal("bot access revocation failed")
	}
	if w := liveRequest(a, "GET", "/api/state", "202", "Member", ""); strings.Contains(w.Body.String(), team.ID) {
		t.Fatal("revoked bot member still sees team in mini-app")
	}
}
