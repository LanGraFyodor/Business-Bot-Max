package app

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func signed(token string, values url.Values) string {
	keys := []string{}
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := []string{}
	for _, k := range keys {
		parts = append(parts, k+"="+values.Get(k))
	}
	key := hmac.New(sha256.New, []byte("WebAppData"))
	key.Write([]byte(token))
	mac := hmac.New(sha256.New, key.Sum(nil))
	mac.Write([]byte(strings.Join(parts, "\n")))
	values.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return values.Encode()
}
func TestMAXSignature(t *testing.T) {
	at := time.Now()
	token := "test-token"
	raw := signed(token, url.Values{"user": {`{"id":42,"first_name":"????"}`}, "auth_date": {strconv.FormatInt(at.Unix(), 10)}})
	u, e := validateInit(raw, token, at)
	if e != nil || u.ID != "42" {
		t.Fatalf("valid signature rejected: %v", e)
	}
	for _, bad := range []string{raw + "&user=other", strings.Replace(raw, "42", "43", 1), strings.Replace(raw, "hash=", "hash=00", 1)} {
		if _, e := validateInit(bad, token, at); e == nil {
			t.Fatal("forged data accepted")
		}
	}
	if _, e := validateInit(raw, token, at.Add(25*time.Hour)); e == nil {
		t.Fatal("expired data accepted")
	}
	if _, e := validateInit(raw, "wrong-token", at); e == nil {
		t.Fatal("wrong token accepted")
	}
	noName := signed(token, url.Values{"user": {`{"id":77,"username":"account77"}`}, "auth_date": {strconv.FormatInt(at.Unix(), 10)}})
	if user, err := validateInit(noName, token, at); err != nil || user.Name != "@account77" {
		t.Fatalf("MAX username fallback failed: %+v %v", user, err)
	}
}
func request(a *App, method, path, user, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("X-Demo-User", user)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	return w
}
func uploadTest(a *App, path, user string, fields map[string]string) *httptest.ResponseRecorder {
	var b bytes.Buffer
	m := multipart.NewWriter(&b)
	for k, v := range fields {
		_ = m.WriteField(k, v)
	}
	part, _ := m.CreateFormFile("file", "test.txt")
	_, _ = part.Write([]byte("Test document content"))
	_ = m.Close()
	r := httptest.NewRequest("POST", path, &b)
	r.Header.Set("Content-Type", m.FormDataContentType())
	r.Header.Set("X-Demo-User", user)
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	return w
}
func TestDocumentWorkflowAndPersistence(t *testing.T) {
	dir := t.TempDir()
	a, e := newApp(dir, true)
	if e != nil {
		t.Fatal(e)
	}
	w := uploadTest(a, "/api/documents", "2", map[string]string{"title": "???????? ???????", "channelId": "legal", "due": "2026-10-01"})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var d Document
	_ = json.Unmarshal(w.Body.Bytes(), &d)
	base := "/api/documents/" + d.ID
	call := func(user, status, reason string, revision, want int) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"status": status, "reason": reason, "revision": revision})
		w := request(a, "PATCH", base+"/status", user, string(b))
		if w.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", user, status, w.Code, want, w.Body.String())
		}
	}
	call("3", "review", "", 1, 403)
	call("2", "approved", "", 1, 403)
	call("2", "review", "", 1, 200)
	call("2", "approved", "", 2, 403)
	call("1", "changes", "", 2, 403)
	call("1", "changes", "????????? ?????????", 2, 200)
	w = request(a, "POST", base+"/comments", "3", `{"text":"???????? ?????????"}`)
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	w = uploadTest(a, base+"/versions", "3", nil)
	if w.Code != 403 {
		t.Fatal("non-owner replaced file")
	}
	w = uploadTest(a, base+"/versions", "2", nil)
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	call("2", "review", "", 4, 200)
	call("1", "approved", "", 4, 409)
	call("1", "approved", "", 5, 200)
	w = request(a, "GET", base+"/files/"+d.Versions[0].ID, "2", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Test document") {
		t.Fatal("file unavailable")
	}
	restored, e := newApp(dir, true)
	if e != nil {
		t.Fatal(e)
	}
	var found *Document
	for _, doc := range restored.state.Documents {
		if doc.ID == d.ID {
			copy := doc
			found = &copy
		}
	}
	if found == nil || found.Status != "approved" || len(found.Versions) != 2 || len(found.Comments) != 6 {
		t.Fatalf("persistence failed: %+v", found)
	}
	if _, e = os.Stat(filepath.Join(dir, "files", d.Versions[0].ID)); e != nil {
		t.Fatal(e)
	}
}
func TestProductionDeniesDemoAndNonMembers(t *testing.T) {
	t.Setenv("MAX_BOT_TOKEN", "test")
	a, e := newApp(t.TempDir(), false)
	if e != nil {
		t.Fatal(e)
	}
	w := request(a, "GET", "/api/state", "1", "")
	if w.Code != 401 {
		t.Fatal("demo impersonation accepted")
	}
	raw := signed("test", url.Values{"user": {`{"id":42,"first_name":"Test"}`}, "auth_date": {strconv.FormatInt(time.Now().Unix(), 10)}})
	r := httptest.NewRequest("GET", "/api/state", nil)
	r.Header.Set("X-Max-Init-Data", raw)
	w = httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal("signed user rejected")
	}
	var state struct {
		Teams     []Team     `json:"teams"`
		Documents []Document `json:"documents"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if len(state.Teams) != 0 || len(state.Documents) != 0 {
		t.Fatal("non-member received private team data")
	}

}
func TestWebhookSecret(t *testing.T) {
	a, e := newApp(t.TempDir(), true)
	if e != nil {
		t.Fatal(e)
	}
	t.Setenv("BOT_MODE", "webhook")
	t.Setenv("MAX_WEBHOOK_SECRET", "test-secret")
	w := request(a, "POST", "/api/max/webhook", "1", `{"update_type":"other"}`)
	if w.Code != http.StatusForbidden {
		t.Fatal("webhook accepted without secret")
	}
	r := httptest.NewRequest("POST", "/api/max/webhook", strings.NewReader(`{"update_type":"other"}`))
	r.Header.Set("X-Max-Bot-Api-Secret", "test-secret")
	w = httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}
func TestSaveFailureRollsBack(t *testing.T) {
	a, e := newApp(t.TempDir(), true)
	if e != nil {
		t.Fatal(e)
	}
	count := len(a.state.Channels)
	a.dir = filepath.Join(a.dir, "missing", "folder")
	w := request(a, "POST", "/api/channels", "1", `{"name":"New","description":""}`)
	if w.Code != 500 || len(a.state.Channels) != count {
		t.Fatal("failed persistence did not roll back")
	}
}
func TestTeamIsolationInvitesAndRoles(t *testing.T) {
	a, e := newApp(t.TempDir(), true)
	if e != nil {
		t.Fatal(e)
	}
	state := func(uid string) struct {
		Teams     []Team           `json:"teams"`
		Channels  []Channel        `json:"channels"`
		Documents []Document       `json:"documents"`
		Messages  []ChannelMessage `json:"messages"`
	} {
		t.Helper()
		w := request(a, "GET", "/api/state", uid, "")
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		var data struct {
			Teams     []Team           `json:"teams"`
			Channels  []Channel        `json:"channels"`
			Documents []Document       `json:"documents"`
			Messages  []ChannelMessage `json:"messages"`
		}
		if e := json.Unmarshal(w.Body.Bytes(), &data); e != nil {
			t.Fatal(e)
		}
		return data
	}
	anna := state("2")
	if len(anna.Teams) != 1 || anna.Teams[0].ID != "orbita" {
		t.Fatal("team metadata leaked")
	}
	for _, c := range anna.Channels {
		if c.TeamID == "studio" {
			t.Fatal("studio channel leaked")
		}
	}
	for _, m := range anna.Messages {
		if m.ChannelID == "studio-general" {
			t.Fatal("studio discussion leaked")
		}
	}
	w := uploadTest(a, "/api/documents", "3", map[string]string{"title": "Private studio file", "channelId": "studio-projects"})
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var d Document
	_ = json.Unmarshal(w.Body.Bytes(), &d)
	if w = request(a, "GET", "/api/documents/"+d.ID+"/files/"+d.Versions[0].ID, "2", ""); w.Code != 404 {
		t.Fatal("private file leaked", w.Code)
	}
	if w = request(a, "POST", "/api/documents/"+d.ID+"/comments", "2", `{"text":"test"}`); w.Code != 404 {
		t.Fatal("private comment accepted")
	}
	if w = request(a, "POST", "/api/channels/studio-general/messages", "2", `{"text":"test"}`); w.Code != 404 {
		t.Fatal("private channel writable")
	}
	manager := state("1")
	var invite string
	for _, team := range manager.Teams {
		if team.ID == "studio" {
			invite = team.InviteCode
		}
	}
	if len(invite) < 20 {
		t.Fatal("manager cannot invite")
	}
	body, _ := json.Marshal(map[string]string{"code": invite})
	w = request(a, "POST", "/api/teams/join", "2", string(body))
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	anna = state("2")
	if len(anna.Teams) != 2 {
		t.Fatal("joined team missing")
	}
	if w = request(a, "GET", "/api/documents/"+d.ID+"/files/"+d.Versions[0].ID, "2", ""); w.Code != 200 {
		t.Fatal("joined member cannot open file")
	}
	if w = request(a, "POST", "/api/channels", "2", `{"teamId":"studio","name":"finance","description":""}`); w.Code != 403 {
		t.Fatal("employee created channel")
	}
	w = request(a, "PATCH", "/api/teams/studio/members/2", "1", `{"role":"manager"}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w = request(a, "POST", "/api/channels", "2", `{"teamId":"studio","name":"finance","description":""}`); w.Code != 201 {
		t.Fatal("promoted manager rejected", w.Body.String())
	}
	if w = request(a, "POST", "/api/teams/studio/invite", "8", ""); w.Code != 403 {
		t.Fatal("employee rotated invitation")
	}
	if w = request(a, "POST", "/api/teams/studio/invite", "1", ""); w.Code != 200 {
		t.Fatal("manager cannot rotate")
	}
	if w = request(a, "POST", "/api/teams/join", "2", string(body)); w.Code != 404 {
		t.Fatal("old invite remained valid")
	}
}
func TestNewTeamIsPrivateAndPersists(t *testing.T) {
	dir := t.TempDir()
	a, e := newApp(dir, true)
	if e != nil {
		t.Fatal(e)
	}
	w := request(a, "POST", "/api/teams", "2", `{"name":"Project Delta","color":"blue","description":"Launch team"}`)
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var team Team
	_ = json.Unmarshal(w.Body.Bytes(), &team)
	if a.teamRole(team.ID, "2") != "manager" || a.teamRole(team.ID, "1") != "" {
		t.Fatal("new team membership wrong")
	}
	var channel string
	for _, c := range a.state.Channels {
		if c.TeamID == team.ID {
			channel = c.ID
		}
	}
	if channel == "" {
		t.Fatal("new team has no general channel")
	}
	w = uploadTest(a, "/api/documents", "2", map[string]string{"title": "Secret project", "channelId": channel})
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var d Document
	_ = json.Unmarshal(w.Body.Bytes(), &d)
	if w = request(a, "GET", "/api/documents/"+d.ID+"/files/"+d.Versions[0].ID, "1", ""); w.Code != 404 {
		t.Fatal("other team manager read file")
	}
	restored, e := newApp(dir, true)
	if e != nil {
		t.Fatal(e)
	}
	if restored.teamRole(team.ID, "2") != "manager" || !restored.canRead(channel, "2") {
		t.Fatal("team not persisted")
	}
}
