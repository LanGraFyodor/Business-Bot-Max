package app

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
)

func uploadSignedVersion(a *App, path, user, contents string) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "signed.txt")
	_, _ = part.Write([]byte(contents))
	_ = writer.Close()
	r := httptest.NewRequest("POST", path, &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("X-Demo-User", user)
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	return w
}

func TestManagerSignedCopyCanBeApprovedAndDownloaded(t *testing.T) {
	a, err := newApp(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	uploaded := uploadTest(a, "/api/documents", "2", map[string]string{"title": "Contract", "channelId": "legal"})
	if uploaded.Code != 201 {
		t.Fatal(uploaded.Body.String())
	}
	var doc Document
	if err := json.Unmarshal(uploaded.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	base := "/api/documents/" + doc.ID
	if w := request(a, "PATCH", base+"/status", "2", `{"status":"review","revision":1}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	signed := uploadSignedVersion(a, base+"/versions", "1", "signed and stamped contract")
	if signed.Code != 201 {
		t.Fatal(signed.Body.String())
	}
	if err := json.Unmarshal(signed.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Status != "review" || doc.Revision != 3 || len(doc.Versions) != 2 {
		t.Fatalf("signed copy left approval flow: %+v", doc)
	}
	if w := request(a, "PATCH", base+"/status", "2", `{"status":"approved","revision":3}`); w.Code != 403 {
		t.Fatal("employee approved document")
	}
	if w := request(a, "PATCH", base+"/status", "1", `{"status":"approved","revision":3}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	latest := doc.Versions[1]
	link := request(a, "POST", base+"/files/"+latest.ID+"/download-link", "1", "")
	if link.Code != 201 {
		t.Fatal(link.Body.String())
	}
	var ticket struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(link.Body.Bytes(), &ticket); err != nil {
		t.Fatal(err)
	}
	downloaded := request(a, "GET", ticket.URL, "", "")
	if downloaded.Code != 200 || downloaded.Body.String() != "signed and stamped contract" {
		t.Fatalf("downloaded wrong file: %d %q", downloaded.Code, downloaded.Body.String())
	}
	if !strings.Contains(downloaded.Header().Get("Content-Disposition"), "signed.txt") {
		t.Fatal("filename missing")
	}
}
