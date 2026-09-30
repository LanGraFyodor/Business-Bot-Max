package app

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func problem(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		problem(w, 400, "Некорректные данные запроса")
		return false
	}
	return true
}
func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, map[string]bool{"ok": true}) })
	mux.HandleFunc("POST /api/max/webhook", a.webhook)
	mux.HandleFunc("POST /api/integrations/documents", a.integrationImport)
	mux.HandleFunc("GET /api/integrations/documents", a.integrationRead)
	mux.HandleFunc("GET /api/integrations/channels", a.integrationRead)
	mux.HandleFunc("/api/download/", a.download)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		u, e := a.user(r)
		if e != nil {
			problem(w, 401, e.Error())
			return
		}
		a.api(w, r, u)
	})
	fs := http.FileServer(http.Dir("web/dist"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.WriteHeader(405)
			return
		}
		path := filepath.Join("web/dist", filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))
		if _, e := os.Stat(path); e != nil {
			http.ServeFile(w, r, "web/dist/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	})
}
