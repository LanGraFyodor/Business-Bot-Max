package app

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) api(w http.ResponseWriter, r *http.Request, u User) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.demo && a.registerAccount(u) != nil {
		problem(w, 500, "Не удалось зарегистрировать профиль MAX")
		return
	}
	p := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
	if p == "state" && r.Method == "GET" {
		respond(w, 200, a.visibleState(u))
		return
	}
	if a.demoImportAPI(w, r, u, p) {
		return
	}
	if a.integrationAPI(w, r, u, p) {
		return
	}
	if a.teamAPI(w, r, u, p) {
		return
	}
	if p == "channels" && r.Method == "POST" {
		var in struct {
			TeamID      string `json:"teamId"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if !decode(w, r, &in) {
			return
		}
		if in.TeamID == "" {
			in.TeamID = "orbita"
		}
		if a.teamRole(in.TeamID, u.ID) != "manager" {
			problem(w, 403, "Создать канал может руководитель команды")
			return
		}
		in.Name = strings.TrimSpace(in.Name)
		if in.Name == "" || len([]rune(in.Name)) > 50 {
			problem(w, 400, "Название: от 1 до 50 символов")
			return
		}
		before := a.snapshot()
		c := Channel{ID: id(), Name: in.Name, Description: in.Description, Color: "blue", TeamID: in.TeamID}
		a.state.Channels = append(a.state.Channels, c)
		if a.commit(before) != nil {
			problem(w, 500, "Не удалось сохранить канал")
			return
		}
		respond(w, 201, c)
		return
	}
	if p == "documents" && r.Method == "POST" {
		a.upload(w, r, u, nil)
		return
	}
	seg := strings.Split(p, "/")
	if len(seg) >= 2 && seg[0] == "documents" {
		for i := range a.state.Documents {
			d := &a.state.Documents[i]
			if d.ID != seg[1] {
				continue
			}
			if !a.canRead(d.ChannelID, u.ID) {
				problem(w, 404, "Документ не найден")
				return
			}
			u.Role = a.teamRole(a.channelTeam(d.ChannelID), u.ID)
			if len(seg) == 3 && seg[2] == "escalate" && r.Method == "POST" {
				a.escalate(w, r, u, *d)
				return
			}
			if len(seg) == 3 && seg[2] == "versions" && r.Method == "POST" {
				if d.Owner.ID != u.ID && u.Role != "manager" {
					problem(w, 403, "Новую версию может загрузить автор или руководитель")
					return
				}
				a.upload(w, r, u, d)
				return
			}
			if len(seg) == 5 && seg[2] == "files" && seg[4] == "download-link" && r.Method == "POST" {
				for _, v := range d.Versions {
					if v.ID == seg[3] {
						a.issueDownload(w, u, *d, v)
						return
					}
				}
				problem(w, 404, "File not found")
				return
			}
			if len(seg) == 5 && seg[2] == "files" && seg[4] == "preview" && r.Method == "GET" {
				for _, v := range d.Versions {
					if v.ID == seg[3] {
						a.servePreview(w, v)
						return
					}
				}
				problem(w, 404, "Файл не найден")
				return
			}
			if len(seg) == 4 && seg[2] == "files" && r.Method == "GET" {
				for _, v := range d.Versions {
					if v.ID == seg[3] {
						w.Header().Set("Content-Type", v.MIME)
						w.Header().Set("Content-Security-Policy", "sandbox")
						disposition := "attachment"
						if v.MIME == "application/pdf" || strings.HasPrefix(v.MIME, "image/") || strings.HasPrefix(v.MIME, "text/plain") {
							disposition = "inline"
						}
						w.Header().Set("Content-Disposition", disposition+"; filename*=UTF-8''"+url.PathEscape(v.Name))
						http.ServeFile(w, r, filepath.Join(a.dir, "files", v.ID))
						return
					}
				}
			}
			if len(seg) == 3 && seg[2] == "comments" && r.Method == "POST" {
				var in struct {
					Text string `json:"text"`
				}
				if !decode(w, r, &in) {
					return
				}
				in.Text = strings.TrimSpace(in.Text)
				if in.Text == "" || len([]rune(in.Text)) > 4000 {
					problem(w, 400, "Комментарий: от 1 до 4000 символов")
					return
				}
				before := a.snapshot()
				d.Comments = append(d.Comments, Comment{id(), u, in.Text, now(), false})
				d.Updated = now()
				if a.commit(before) != nil {
					problem(w, 500, "Не удалось сохранить комментарий")
					return
				}
				respond(w, 201, d)
				return
			}
			if len(seg) == 3 && seg[2] == "status" && r.Method == "PATCH" {
				var in struct {
					Status   string `json:"status"`
					Reason   string `json:"reason"`
					Revision int    `json:"revision"`
				}
				if !decode(w, r, &in) {
					return
				}
				if in.Revision != d.Revision {
					problem(w, 409, "Документ изменился. Обновите карточку и повторите действие")
					return
				}
				if e := transition(*d, u, in.Status, in.Reason); e != nil {
					problem(w, 403, e.Error())
					return
				}
				before := a.snapshot()
				d.Status = in.Status
				d.Revision++
				d.Updated = now()
				labels := map[string]string{"review": "Отправлено на согласование", "approved": "Документ согласован", "changes": "Возвращено на доработку"}
				d.Comments = append(d.Comments, Comment{id(), u, strings.TrimSpace(labels[in.Status] + ". " + in.Reason), now(), true})
				a.queueNotification(*d)
				if a.commit(before) != nil {
					problem(w, 500, "Не удалось сохранить решение")
					return
				}
				respond(w, 200, d)
				return
			}
			problem(w, 404, "Действие не найдено")
			return
		}
	}
	problem(w, 404, "Не найдено")
}
func transition(d Document, u User, status, reason string) error {
	if status == "review" && (d.Status == "draft" || d.Status == "changes") && (u.ID == d.Owner.ID || u.Role == "manager") {
		return nil
	}
	if (status == "approved" || status == "changes") && d.Status == "review" && u.Role == "manager" {
		if status == "changes" && strings.TrimSpace(reason) == "" {
			return errors.New("Укажите причину возврата")
		}
		if len([]rune(reason)) > 4000 {
			return errors.New("Слишком длинный комментарий")
		}
		return nil
	}
	return errors.New("Недопустимое действие для вашей роли или текущего статуса")
}
func (a *App) upload(w http.ResponseWriter, r *http.Request, u User, existing *Document) {
	r.Body = http.MaxBytesReader(w, r.Body, 26<<20)
	if e := r.ParseMultipartForm(26 << 20); e != nil {
		problem(w, 400, "Не удалось загрузить файл. Лимит — 25 МБ")
		return
	}
	defer r.MultipartForm.RemoveAll()
	title := strings.TrimSpace(r.FormValue("title"))
	channel := r.FormValue("channelId")
	due := r.FormValue("due")
	source := r.FormValue("sourceSystem")
	if existing == nil && source != "" && source != "1c" && source != "bitrix24" && source != "amocrm" && source != "other" {
		problem(w, 400, "Неизвестный источник отчёта")
		return
	}
	if due != "" {
		if _, e := time.Parse("2006-01-02", due); e != nil {
			problem(w, 400, "Некорректный срок")
			return
		}
	}
	if existing == nil {
		if !a.canRead(channel, u.ID) {
			problem(w, 404, "Канал не найден")
			return
		}
		found := false
		for _, c := range a.state.Channels {
			if c.ID == channel {
				found = true
			}
		}
		if !found || title == "" || len([]rune(title)) > 180 {
			problem(w, 400, "Укажите канал и название до 180 символов")
			return
		}
	}
	f, h, e := r.FormFile("file")
	if e != nil {
		problem(w, 400, "Выберите файл")
		return
	}
	defer f.Close()
	if h.Size > 25<<20 || h.Size == 0 {
		problem(w, 400, "Размер файла: от 1 байта до 25 МБ")
		return
	}
	mime, ok := fileMIME(h.Filename)
	if !ok {
		problem(w, 400, "Поддерживаются PDF, DOCX, XLSX, TXT, CSV, PNG и JPG")
		return
	}
	v := Version{id(), filepath.Base(h.Filename), mime, h.Size, now(), u.Name}
	path := filepath.Join(a.dir, "files", v.ID)
	out, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		problem(w, 500, "Ошибка хранилища")
		return
	}
	_, e = io.Copy(out, f)
	closeErr := out.Close()
	if e != nil || closeErr != nil {
		_ = os.Remove(path)
		problem(w, 500, "Файл не сохранён")
		return
	}
	before := a.snapshot()
	var d *Document
	if existing != nil {
		d = existing
		d.Versions = append(d.Versions, v)
		// A manager can attach the signed copy to a document awaiting approval.
		keepReview := d.Status == "review" && u.Role == "manager"
		if !keepReview {
			d.Status = "draft"
		}
		d.Revision++
		d.Updated = now()
		message := "Загружена новая версия. Требуется повторное согласование."
		if keepReview {
			message = "Руководитель загрузил новую версию на согласование."
		}
		d.Comments = append(d.Comments, Comment{id(), u, message, now(), true})
	} else {
		u.Role = a.teamRole(a.channelTeam(channel), u.ID)
		a.state.Documents = append([]Document{{ID: id(), Title: title, ChannelID: channel, Owner: u, Status: "draft", Due: due, Description: r.FormValue("description"), Created: now(), Updated: now(), Versions: []Version{v}, Comments: []Comment{}, Revision: 1}}, a.state.Documents...)
		d = &a.state.Documents[0]
		if source != "" {
			d.SourceSystem = source
			d.ImportMethod = "file"
		}
		if r.FormValue("review") == "1" {
			d.Status = "review"
			d.Comments = append(d.Comments, Comment{ID: id(), Author: u, Text: "Отчёт загружен сотрудником и отправлен на согласование.", Created: now(), System: true})
			a.queueNotification(*d)
		}
	}
	if a.commit(before) != nil {
		_ = os.Remove(path)
		problem(w, 500, "Не удалось сохранить документ")
		return
	}
	respond(w, 201, d)
}
