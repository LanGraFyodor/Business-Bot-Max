package app

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type downloadTicket struct {
	DocumentID string
	VersionID  string
	UserID     string
	Expires    time.Time
}

func (a *App) issueDownload(w http.ResponseWriter, u User, d Document, v Version) {
	if a.downloads == nil {
		a.downloads = map[string]downloadTicket{}
	}
	for key, ticket := range a.downloads {
		if time.Now().After(ticket.Expires) {
			delete(a.downloads, key)
		}
	}
	key := id()
	a.downloads[key] = downloadTicket{d.ID, v.ID, u.ID, time.Now().Add(3 * time.Minute)}
	respond(w, http.StatusCreated, map[string]string{"url": "/api/download/" + key})
}

func (a *App) download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, "/api/download/")
	a.mu.Lock()
	ticket, found := a.downloads[key]
	if !found || time.Now().After(ticket.Expires) {
		delete(a.downloads, key)
		a.mu.Unlock()
		problem(w, http.StatusNotFound, "Ссылка недоступна")
		return
	}
	for _, d := range a.state.Documents {
		if d.ID != ticket.DocumentID || !a.canRead(d.ChannelID, ticket.UserID) {
			continue
		}
		for _, v := range d.Versions {
			if v.ID != ticket.VersionID {
				continue
			}
			a.mu.Unlock()
			w.Header().Set("Content-Type", v.MIME)
			w.Header().Set("Content-Security-Policy", "sandbox")
			w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(v.Name))
			http.ServeFile(w, r, filepath.Join(a.dir, "files", v.ID))
			return
		}
	}
	a.mu.Unlock()
	problem(w, http.StatusNotFound, "Файл недоступен")
}
