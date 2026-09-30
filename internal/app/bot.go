package app

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Update struct {
	Type string `json:"update_type"`
	User struct {
		ID        int64  `json:"user_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"user"`
	Message struct {
		Sender struct {
			ID        int64  `json:"user_id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"sender"`
		Body struct {
			Text string `json:"text"`
		} `json:"body"`
	} `json:"message"`
}

func (a *App) maxRequest(ctx context.Context, method, path string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		b, e := json.Marshal(payload)
		if e != nil {
			return e
		}
		body = bytes.NewReader(b)
	}
	req, e := http.NewRequestWithContext(ctx, method, env("MAX_API_URL", "https://platform-api2.max.ru")+path, body)
	if e != nil {
		return e
	}
	req.Header.Set("Authorization", a.token)
	req.Header.Set("Content-Type", "application/json")
	res, e := a.client.Do(req)
	if e != nil {
		return fmt.Errorf("MAX API: connection failed (check network and trusted CA)")
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("MAX API returned HTTP %d", res.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(out)
	}
	return nil
}
func (a *App) send(ctx context.Context, userID, text string) error {
	payload := map[string]any{"text": text}
	if os.Getenv("PUBLIC_APP_URL") != "" {
		payload["attachments"] = []any{map[string]any{"type": "inline_keyboard", "payload": map[string]any{"buttons": [][]any{
			{map[string]any{"type": "open_app", "text": "Документы", "web_app": "t68_hakaton_max_bot", "payload": "docspace"}},
			{map[string]any{"type": "open_app", "text": "Бизнес-навигатор", "web_app": "t68_hakaton_max_bot", "payload": "guide"}},
		}}}}
	}
	return a.maxRequest(ctx, "POST", "/messages?user_id="+url.QueryEscape(userID), payload, nil)
}
func (a *App) handleUpdate(ctx context.Context, u Update) error {
	userID := u.User.ID
	if u.Type == "message_created" {
		userID = u.Message.Sender.ID
	}
	if userID == 0 || (u.Type != "bot_started" && u.Type != "message_created") {
		return nil
	}
	key := strconv.FormatInt(userID, 10)
	command := strings.TrimSpace(u.Message.Body.Text)
	if command == "/id" {
		return a.send(ctx, key, "MAX ID: "+key)
	}
	if reply, handled := a.botTeamCommand(u, key, command); handled {
		return a.send(ctx, key, reply)
	}
	a.mu.Lock()
	membership := false
	for _, t := range a.state.Teams {
		if a.teamRole(t.ID, key) != "" {
			membership = true
			break
		}
	}
	a.mu.Unlock()
	if !membership && command != "/teams" {
		return a.send(ctx, key, "Добро пожаловать в Docspace! Ваш MAX ID: "+key+"\nСоздать команду: /create Название\nВступить по приглашению: /join КОД\nПосле подключения мини-приложения можно работать с документами с телефона.")
	}
	text := "Docspace — документы и команда в одном месте.\n\nЗагрузите документ в канал → обсудите с коллегами → отправьте руководителю на согласование.\n\n/id — мой MAX ID\n/create Название — создать команду\n/invite ID_команды employee|manager [MAX_ID] — пригласить\n/join КОД — вступить\n/role ID_команды MAX_ID employee|manager — сменить роль\n/remove ID_команды MAX_ID — исключить\n/documents — последние документы\n/review — ожидают согласования\n/teams — мои команды\n/help — помощь\n\nОткройте мини-приложение кнопкой после его подключения."
	if command == "/teams" {
		a.mu.Lock()
		text = "Ваши команды:"
		for _, team := range a.state.Teams {
			if a.teamRole(team.ID, key) != "" {
				role := a.teamRole(team.ID, key)
				label := "участник"
				if role == "manager" {
					label = "руководитель"
				}
				text += "\n• " + team.Name + " · " + label + " · ID: " + team.ID
			}
		}
		a.mu.Unlock()
		if text == "Ваши команды:" {
			text = "Вы пока не состоите в командах. Создайте команду в мини-приложении или попросите код приглашения."
		}
	}
	if command == "/documents" || command == "/review" {
		a.mu.Lock()
		text = "Документы рабочего пространства:\n"
		count := 0
		labels := map[string]string{"draft": "Черновик", "review": "На согласовании", "approved": "Согласован", "changes": "На доработке"}
		for _, d := range a.state.Documents {
			if !a.canRead(d.ChannelID, key) {
				continue
			}
			if command == "/review" && d.Status != "review" {
				continue
			}
			text += "\n• " + d.Title + " — " + labels[d.Status]
			count++
			if count == 10 {
				break
			}
		}
		a.mu.Unlock()
		if count == 0 {
			text = "Документов пока нет. Загрузите первый в мини-приложении."
		}
	}
	return a.send(ctx, key, text)
}
func (a *App) webhook(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("MAX_WEBHOOK_SECRET")
	if env("BOT_MODE", "off") != "webhook" || secret == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Max-Bot-Api-Secret")), []byte(secret)) != 1 {
		problem(w, 403, "Forbidden")
		return
	}
	var u Update
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if json.NewDecoder(r.Body).Decode(&u) != nil {
		problem(w, 400, "Invalid update")
		return
	}
	if e := a.handleUpdate(r.Context(), u); e != nil {
		problem(w, 502, "MAX unavailable")
		return
	}
	respond(w, 200, map[string]bool{"ok": true})
}
func (a *App) poll(ctx context.Context) {
	if a.token == "" {
		log.Print("Bot disabled: MAX_BOT_TOKEN is empty")
		return
	}
	a.mu.Lock()
	marker := a.state.PollMarker
	a.mu.Unlock()
	for ctx.Err() == nil {
		path := "/updates?timeout=25&types=bot_started,message_created"
		if marker != "" {
			path += "&marker=" + url.QueryEscape(marker)
		}
		var batch struct {
			Updates []Update    `json:"updates"`
			Marker  json.Number `json:"marker"`
		}
		if e := a.maxRequest(ctx, "GET", path, nil, &batch); e != nil {
			log.Print(e)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}
		for _, u := range batch.Updates {
			if e := a.handleUpdate(ctx, u); e != nil {
				log.Print(e)
			}
		}
		if batch.Marker != "" && marker != string(batch.Marker) {
			a.mu.Lock()
			before := a.snapshot()
			a.state.PollMarker = string(batch.Marker)
			if err := a.commit(before); err != nil {
				log.Print("Could not save MAX polling marker: ", err)
			} else {
				marker = string(batch.Marker)
			}
			a.mu.Unlock()
		}
	}
}

type Notification struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"`
	Text     string `json:"text"`
	Attempts int    `json:"attempts"`
	Next     int64  `json:"next"`
}

func (a *App) queueNotification(d Document) {
	if d.Status != "review" && strings.HasPrefix(d.Owner.ID, "integration:") {
		return
	}
	if a.demo || a.token == "" {
		return
	}
	recipients := []string{d.Owner.ID}
	text := "Docspace: «" + d.Title + "» — " + map[string]string{"review": "на согласовании", "approved": "согласован", "changes": "на доработке"}[d.Status] + "."
	if d.Status == "review" {
		recipients = nil
		teamID := a.channelTeam(d.ChannelID)
		for _, t := range a.state.Teams {
			if t.ID == teamID {
				for _, m := range t.Members {
					if a.teamRole(t.ID, m.ID) == "manager" {
						recipients = append(recipients, m.ID)
					}
				}
			}
		}
	}
	for _, uid := range recipients {
		numeric, err := strconv.ParseInt(uid, 10, 64)
		if err != nil || numeric <= 0 || !a.canRead(d.ChannelID, uid) {
			continue
		}
		a.state.Outbox = append(a.state.Outbox, Notification{ID: id(), UserID: uid, Text: text})
	}
}
func (a *App) deliverNotifications(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		a.mu.Lock()
		var next *Notification
		for _, n := range a.state.Outbox {
			if n.Next <= time.Now().Unix() && n.Attempts < 8 {
				copy := n
				next = &copy
				break
			}
		}
		a.mu.Unlock()
		if next == nil {
			continue
		}
		sendCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		err := a.send(sendCtx, next.UserID, next.Text)
		cancel()
		a.mu.Lock()
		before := a.snapshot()
		for i, n := range a.state.Outbox {
			if n.ID == next.ID {
				if err == nil {
					a.state.Outbox = append(a.state.Outbox[:i], a.state.Outbox[i+1:]...)
				} else {
					a.state.Outbox[i].Attempts++
					a.state.Outbox[i].Next = time.Now().Add(time.Duration(1<<min(a.state.Outbox[i].Attempts, 8)) * time.Second).Unix()
				}
				break
			}
		}
		if e := a.commit(before); e != nil {
			log.Print("Cannot persist notification delivery")
		}
		a.mu.Unlock()
		if err != nil {
			log.Print(err)
		}
	}
}
