package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Integration tokens belong to one team and are stored only as SHA-256 digests.
func (a *App) integrationAPI(w http.ResponseWriter, r *http.Request, u User, p string) bool {
	seg := strings.Split(p, "/")
	if len(seg) < 3 || seg[0] != "teams" || seg[2] != "integrations" || (len(seg) != 3 && len(seg) != 4) {
		return false
	}
	if a.teamRole(seg[1], u.ID) != "manager" {
		problem(w, 403, "Интеграциями управляет руководитель команды")
		return true
	}
	if len(seg) == 3 && r.Method == "GET" {
		items := []map[string]string{}
		for _, key := range a.state.Integrations {
			if key.TeamID == seg[1] {
				items = append(items, map[string]string{"id": key.ID, "name": key.Name, "provider": key.Provider, "created": key.Created})
			}
		}
		respond(w, 200, items)
		return true
	}
	if len(seg) == 3 && r.Method == "POST" {
		var in struct {
			Name     string `json:"name"`
			Provider string `json:"provider"`
		}
		if !decode(w, r, &in) {
			return true
		}
		in.Name = strings.TrimSpace(in.Name)
		if in.Name == "" || len([]rune(in.Name)) > 50 {
			problem(w, 400, "Название: от 1 до 50 символов")
			return true
		}
		if in.Provider != "1c" && in.Provider != "bitrix24" && in.Provider != "amocrm" && in.Provider != "other" {
			problem(w, 400, "Неизвестный источник")
			return true
		}
		count := 0
		for _, key := range a.state.Integrations {
			if key.TeamID == seg[1] {
				count++
			}
		}
		if count >= 10 {
			problem(w, 400, "Не более 10 ключей на команду")
			return true
		}
		token := "dsi_" + id() + id()
		digest := sha256.Sum256([]byte(token))
		key := IntegrationKey{ID: id(), TeamID: seg[1], Name: in.Name, Provider: in.Provider, Hash: hex.EncodeToString(digest[:]), Created: now()}
		before := a.snapshot()
		a.state.Integrations = append(a.state.Integrations, key)
		if a.commit(before) != nil {
			problem(w, 500, "Не удалось создать ключ")
			return true
		}
		respond(w, 201, map[string]string{"id": key.ID, "name": key.Name, "provider": key.Provider, "created": key.Created, "token": token})
		return true
	}
	if len(seg) == 4 && r.Method == "DELETE" {
		for i, key := range a.state.Integrations {
			if key.TeamID == seg[1] && key.ID == seg[3] {
				before := a.snapshot()
				a.state.Integrations = append(a.state.Integrations[:i], a.state.Integrations[i+1:]...)
				if a.commit(before) != nil {
					problem(w, 500, "Не удалось отозвать ключ")
					return true
				}
				respond(w, 200, map[string]bool{"ok": true})
				return true
			}
		}
		problem(w, 404, "Ключ не найден")
		return true
	}
	problem(w, 405, "Метод не поддерживается")
	return true
}

func fileMIME(name string) (string, bool) {
	types := map[string]string{
		".pdf": "application/pdf", ".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
		".txt": "text/plain; charset=utf-8", ".csv": "text/csv; charset=utf-8",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}
	mime, ok := types[strings.ToLower(filepath.Ext(name))]
	return mime, ok
}

// A 1C HTTP request (or another authorized CRM) sends raw file bytes here.
func (a *App) integrationImport(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := a.integrationKey(r.Header.Get("Authorization"))
	if key == nil {
		problem(w, 401, "Неверный ключ интеграции")
		return
	}
	q := r.URL.Query()
	channelID := q.Get("channelId")
	if a.channelTeam(channelID) != key.TeamID {
		problem(w, 404, "Канал этой команды не найден")
		return
	}
	title := strings.TrimSpace(q.Get("title"))
	filename := filepath.Base(strings.ReplaceAll(q.Get("filename"), "\\", "/"))
	if title == "" || len([]rune(title)) > 180 || filename == "." || filename == "" || len([]rune(filename)) > 180 {
		problem(w, 400, "Укажите название и имя файла до 180 символов")
		return
	}
	mime, ok := fileMIME(filename)
	if !ok {
		problem(w, 400, "Поддерживаются PDF, DOCX, XLSX, TXT, CSV, PNG и JPG")
		return
	}
	due := q.Get("due")
	if due != "" {
		if _, err := time.Parse("2006-01-02", due); err != nil {
			problem(w, 400, "Некорректный срок")
			return
		}
	}
	description := q.Get("description")
	if len([]rune(description)) > 2000 {
		problem(w, 400, "Описание слишком длинное")
		return
	}
	externalID := q.Get("externalId")
	if len([]rune(externalID)) > 180 {
		problem(w, 400, "Идентификатор источника слишком длинный")
		return
	}
	if externalID != "" {
		for _, d := range a.state.Documents {
			if d.IntegrationID == key.ID && d.ExternalID == externalID {
				respond(w, 200, d)
				return
			}
		}
	}
	if r.ContentLength > 25<<20 {
		problem(w, 413, "Файл больше 25 МБ")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 25<<20)
	versionID := id()
	path := filepath.Join(a.dir, "files", versionID)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		problem(w, 500, "Ошибка хранилища")
		return
	}
	size, copyErr := io.Copy(file, r.Body)
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil || size == 0 {
		_ = os.Remove(path)
		problem(w, 400, "Не удалось принять файл размером до 25 МБ")
		return
	}
	owner := User{ID: "integration:" + key.ID, Name: key.Name, Role: "employee"}
	status := "draft"
	comments := []Comment{}
	if q.Get("review") == "1" {
		status = "review"
		comments = append(comments, Comment{ID: id(), Author: owner, Text: "Поступил из внешней системы на согласование.", Created: now(), System: true})
	}
	stamp := now()
	d := Document{ID: id(), Title: title, ChannelID: channelID, Owner: owner, Status: status, Due: due, Description: description, Created: stamp, Updated: stamp, Versions: []Version{{ID: versionID, Name: filename, MIME: mime, Size: size, Created: stamp, Author: key.Name}}, Comments: comments, Revision: 1, IntegrationID: key.ID, SourceSystem: key.Provider, ImportMethod: "api", ExternalID: externalID}
	before := a.snapshot()
	a.state.Documents = append([]Document{d}, a.state.Documents...)
	if status == "review" {
		a.queueNotification(d)
	}
	if a.commit(before) != nil {
		_ = os.Remove(path)
		problem(w, 500, "Не удалось сохранить документ")
		return
	}
	respond(w, 201, d)
}

// Caller must hold a.mu. Keys are scoped to the issuing team and connector.
func (a *App) integrationKey(authorization string) *IntegrationKey {
	if !strings.HasPrefix(authorization, "Bearer ") {
		return nil
	}
	token := strings.TrimPrefix(authorization, "Bearer ")
	if !strings.HasPrefix(token, "dsi_") || len(token) != 52 {
		return nil
	}
	digest := sha256.Sum256([]byte(token))
	for i := range a.state.Integrations {
		saved, err := hex.DecodeString(a.state.Integrations[i].Hash)
		if err == nil && hmac.Equal(saved, digest[:]) {
			for _, team := range a.state.Teams {
				if team.ID == a.state.Integrations[i].TeamID && teamAccessActive(team) {
					return &a.state.Integrations[i]
				}
			}
			return nil
		}
	}
	return nil
}

func (a *App) integrationRead(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := a.integrationKey(r.Header.Get("Authorization"))
	if key == nil {
		problem(w, 401, "Неверный ключ интеграции")
		return
	}
	if r.URL.Path == "/api/integrations/channels" {
		channels := []Channel{}
		for _, channel := range a.state.Channels {
			if channel.TeamID == key.TeamID {
				channels = append(channels, channel)
			}
		}
		respond(w, 200, channels)
		return
	}
	externalID := r.URL.Query().Get("externalId")
	if externalID == "" || len([]rune(externalID)) > 180 {
		problem(w, 400, "Укажите внешний ID отчёта")
		return
	}
	for _, doc := range a.state.Documents {
		if doc.IntegrationID == key.ID && doc.ExternalID == externalID {
			respond(w, 200, doc)
			return
		}
	}
	problem(w, 404, "Отчёт этого подключения не найден")
}
