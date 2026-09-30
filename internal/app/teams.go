package app

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

type Team struct {
	ReviewUntil   string       `json:"reviewUntil,omitempty"`
	ParentTeamID  string       `json:"parentTeamId,omitempty"`
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Color         string       `json:"color"`
	Description   string       `json:"description"`
	Members       []User       `json:"members"`
	InviteCode    string       `json:"inviteCode,omitempty"`
	InviteExpires string       `json:"inviteExpires,omitempty"`
	Invites       []Invitation `json:"invites,omitempty"`
}
type OrgTeam struct {
	ID           string `json:"id"`
	ParentTeamID string `json:"parentTeamId,omitempty"`
	Name         string `json:"name"`
	MemberCount  int    `json:"memberCount"`
}

func (a *App) canSeeOrgTeam(t Team, userID string) bool {
	if a.teamRole(t.ID, userID) != "" {
		return true
	}
	parentID := t.ParentTeamID
	seen := map[string]bool{}
	for parentID != "" && !seen[parentID] {
		seen[parentID] = true
		if a.teamRole(parentID, userID) == "manager" {
			return true
		}
		next := ""
		for _, candidate := range a.state.Teams {
			if candidate.ID == parentID {
				next = candidate.ParentTeamID
				break
			}
		}
		parentID = next
	}
	return false
}

type ChannelMessage struct {
	ID        string `json:"id"`
	ChannelID string `json:"channelId"`
	Author    User   `json:"author"`
	Text      string `json:"text"`
	Created   string `json:"created"`
}

func teamAccessActive(t Team) bool {
	if t.ReviewUntil == "" {
		return true
	}
	until, err := time.Parse(time.RFC3339, t.ReviewUntil)
	return err == nil && time.Now().Before(until)
}

func (a *App) teamRole(teamID, userID string) string {
	for _, t := range a.state.Teams {
		if t.ID == teamID {
			if !teamAccessActive(t) {
				return ""
			}
			for _, m := range t.Members {
				if m.ID == userID {
					return m.Role
				}
			}
		}
	}
	return ""
}
func (a *App) channelTeam(channelID string) string {
	for _, c := range a.state.Channels {
		if c.ID == channelID {
			return c.TeamID
		}
	}
	return ""
}
func (a *App) canRead(channelID, userID string) bool {
	return a.teamRole(a.channelTeam(channelID), userID) != ""
}
func (a *App) publicTeam(t Team, u User) Team {
	if a.teamRole(t.ID, u.ID) != "manager" {
		t.InviteCode = ""
		t.InviteExpires = ""
		t.Invites = nil
	}
	return t
}
func (a *App) visibleState(u User) map[string]any {
	teams := []Team{}
	channels := []Channel{}
	docs := []Document{}
	messages := []ChannelMessage{}
	orgChart := []OrgTeam{}
	for _, t := range a.state.Teams {
		if a.canSeeOrgTeam(t, u.ID) {
			orgChart = append(orgChart, OrgTeam{ID: t.ID, ParentTeamID: t.ParentTeamID, Name: t.Name, MemberCount: len(t.Members)})
		}
		if a.teamRole(t.ID, u.ID) != "" {
			teams = append(teams, a.publicTeam(t, u))
		}
	}
	for _, c := range a.state.Channels {
		if a.canRead(c.ID, u.ID) {
			channels = append(channels, c)
		}
	}
	for _, d := range a.state.Documents {
		if a.canRead(d.ChannelID, u.ID) {
			docs = append(docs, d)
		}
	}
	for _, m := range a.state.Messages {
		if a.canRead(m.ChannelID, u.ID) {
			messages = append(messages, m)
		}
	}
	result := map[string]any{"user": u, "demo": a.demo, "teams": teams, "orgChart": orgChart, "channels": channels, "documents": docs, "messages": messages, "botConfigured": a.token != ""}
	if a.demo {
		profiles := []User{}
		for _, profile := range demoUsers {
			profiles = append(profiles, profile)
		}
		sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
		result["demoProfiles"] = profiles
	}
	return result
}

// Runs once for the original single-workspace data. Existing document IDs and files are preserved.
func (a *App) migrateTeams() bool {
	if len(a.state.Teams) > 0 || (len(a.state.Channels) == 0 && len(a.state.Documents) == 0) {
		return false
	}
	members := []User{}
	if a.demo {
		for _, uid := range []string{"1", "2", "3"} {
			members = append(members, demoUsers[uid])
		}
	} else {
		for uid := range a.managers {
			members = append(members, User{uid, "Участник " + uid, "manager"})
		}
		for uid := range a.members {
			if !a.managers[uid] {
				members = append(members, User{uid, "Участник " + uid, "employee"})
			}
		}
	}
	a.state.Teams = []Team{{ID: "orbita", Name: "Орбита", Color: "violet", Description: "Основная команда · документы и проекты", Members: members, InviteCode: id(), InviteExpires: time.Now().AddDate(0, 0, 7).UTC().Format(time.RFC3339)}}
	for i := range a.state.Channels {
		a.state.Channels[i].TeamID = "orbita"
	}
	if a.state.Messages == nil {
		a.state.Messages = []ChannelMessage{}
	}
	if a.demo {
		a.state.Teams = append(a.state.Teams, Team{ID: "studio", Name: "Дизайн-студия", Color: "coral", Description: "Айдентика, дизайн и запуски", Members: []User{demoUsers["1"], demoUsers["3"]}, InviteCode: id(), InviteExpires: time.Now().AddDate(0, 0, 7).UTC().Format(time.RFC3339)})
		a.state.Channels = append(a.state.Channels, Channel{ID: "studio-general", Name: "общее", Description: "Рабочие вопросы студии", Color: "coral", TeamID: "studio"}, Channel{ID: "studio-projects", Name: "проекты", Description: "Брифы и материалы клиентов", Color: "coral", TeamID: "studio"})
		a.state.Messages = append(a.state.Messages, ChannelMessage{id(), "general", demoUsers["2"], "Коллеги, здесь собираем документы команды. Замечания к конкретному файлу оставляйте в его обсуждении — так они не потеряются.", now()}, ChannelMessage{id(), "studio-general", demoUsers["3"], "Привет! Это отдельная команда студии. Её документы и переписка доступны только участникам.", now()})
	}
	return true
}
func (a *App) teamAPI(w http.ResponseWriter, r *http.Request, u User, p string) bool {
	if a.inviteAPI(w, r, u, p) {
		return true
	}
	if p == "teams" && r.Method == "POST" {
		var in struct {
			Name         string `json:"name"`
			Description  string `json:"description"`
			Color        string `json:"color"`
			ParentTeamID string `json:"parentTeamId"`
		}
		if !decode(w, r, &in) {
			return true
		}
		in.Name = strings.TrimSpace(in.Name)
		if in.ParentTeamID != "" && a.teamRole(in.ParentTeamID, u.ID) == "" {
			problem(w, 403, "Parent team requires membership")
			return true
		}
		if in.Name == "" || len([]rune(in.Name)) > 40 || len([]rune(in.Description)) > 200 {
			problem(w, 400, "Название до 40 символов, описание до 200")
			return true
		}
		count := 0
		for _, t := range a.state.Teams {
			if a.teamRole(t.ID, u.ID) == "manager" {
				count++
			}
		}
		if count >= 20 {
			problem(w, 400, "Можно управлять не более чем 20 командами")
			return true
		}
		colors := map[string]bool{"violet": true, "blue": true, "green": true, "coral": true}
		if !colors[in.Color] {
			in.Color = "violet"
		}
		before := a.snapshot()
		owner := u
		owner.Role = "manager"
		t := Team{ID: id(), ParentTeamID: in.ParentTeamID, Name: in.Name, Color: in.Color, Description: in.Description, Members: []User{owner}, Invites: []Invitation{}}
		a.state.Teams = append(a.state.Teams, t)
		a.state.Channels = append(a.state.Channels, Channel{ID: id(), Name: "общее", Description: "Главный канал команды", Color: in.Color, TeamID: t.ID})
		if a.commit(before) != nil {
			problem(w, 500, "Не удалось создать команду")
			return true
		}
		respond(w, 201, t)
		return true
	}
	if p == "teams/join" && r.Method == "POST" {
		var in struct {
			Code string `json:"code"`
		}
		if !decode(w, r, &in) {
			return true
		}
		code := strings.TrimSpace(in.Code)
		if code == "" {
			problem(w, 400, "Введите код приглашения")
			return true
		}
		if a.joinInvitation(w, u, code) {
			return true
		}
		for i := range a.state.Teams {
			t := &a.state.Teams[i]
			if t.InviteCode == "" || t.InviteCode != code {
				continue
			}
			expiry, err := time.Parse(time.RFC3339, t.InviteExpires)
			if err != nil || time.Now().After(expiry) {
				problem(w, 400, "Приглашение истекло. Попросите новый код")
				return true
			}
			if a.teamRole(t.ID, u.ID) != "" {
				respond(w, 200, a.publicTeam(*t, u))
				return true
			}
			before := a.snapshot()
			member := u
			member.Role = "employee"
			t.Members = append(t.Members, member)
			if a.commit(before) != nil {
				problem(w, 500, "Не удалось вступить в команду")
				return true
			}
			respond(w, 200, a.publicTeam(*t, u))
			return true
		}
		problem(w, 404, "Приглашение не найдено")
		return true
	}
	seg := strings.Split(p, "/")
	if len(seg) == 3 && seg[0] == "teams" && seg[2] == "invite" && r.Method == "POST" {
		if a.teamRole(seg[1], u.ID) != "manager" {
			problem(w, 403, "Приглашать участников может руководитель команды")
			return true
		}
		for i := range a.state.Teams {
			t := &a.state.Teams[i]
			if t.ID == seg[1] {
				before := a.snapshot()
				t.InviteCode = id()
				t.InviteExpires = time.Now().AddDate(0, 0, 7).UTC().Format(time.RFC3339)
				if a.commit(before) != nil {
					problem(w, 500, "Не удалось обновить приглашение")
					return true
				}
				respond(w, 200, t)
				return true
			}
		}
	}
	if len(seg) == 4 && seg[0] == "teams" && seg[2] == "members" && r.Method == "DELETE" {
		if a.teamRole(seg[1], u.ID) != "manager" {
			problem(w, 403, "Недостаточно прав")
			return true
		}
		for i := range a.state.Teams {
			t := &a.state.Teams[i]
			if t.ID != seg[1] {
				continue
			}
			managers := 0
			for _, m := range t.Members {
				if m.Role == "manager" {
					managers++
				}
			}
			for j, m := range t.Members {
				if m.ID != seg[3] {
					continue
				}
				if m.Role == "manager" && managers == 1 {
					problem(w, 400, "В команде должен остаться руководитель")
					return true
				}
				before := a.snapshot()
				t.Members = append(t.Members[:j], t.Members[j+1:]...)
				if a.commit(before) != nil {
					problem(w, 500, "Не удалось исключить участника")
					return true
				}
				respond(w, 200, a.publicTeam(*t, u))
				return true
			}
		}
		problem(w, 404, "Участник не найден")
		return true
	}
	if len(seg) == 4 && seg[0] == "teams" && seg[2] == "members" && r.Method == "PATCH" {
		if a.teamRole(seg[1], u.ID) != "manager" {
			problem(w, 403, "Недостаточно прав")
			return true
		}
		var in struct {
			Role string `json:"role"`
		}
		if !decode(w, r, &in) {
			return true
		}
		if in.Role != "employee" && in.Role != "manager" {
			problem(w, 400, "Неизвестная роль")
			return true
		}
		for i := range a.state.Teams {
			t := &a.state.Teams[i]
			if t.ID != seg[1] {
				continue
			}
			managers := 0
			for _, m := range t.Members {
				if m.Role == "manager" {
					managers++
				}
			}
			for j, m := range t.Members {
				if m.ID != seg[3] {
					continue
				}
				if m.Role == "manager" && in.Role != "manager" && managers == 1 {
					problem(w, 400, "В команде должен остаться руководитель")
					return true
				}
				before := a.snapshot()
				t.Members[j].Role = in.Role
				if a.commit(before) != nil {
					problem(w, 500, "Не удалось изменить роль")
					return true
				}
				respond(w, 200, a.publicTeam(*t, u))
				return true
			}
		}
		problem(w, 404, "Участник не найден")
		return true
	}
	if len(seg) == 3 && seg[0] == "channels" && seg[2] == "messages" && r.Method == "POST" {
		if !a.canRead(seg[1], u.ID) {
			problem(w, 404, "Канал не найден")
			return true
		}
		var in struct {
			Text string `json:"text"`
		}
		if !decode(w, r, &in) {
			return true
		}
		in.Text = strings.TrimSpace(in.Text)
		if in.Text == "" || len([]rune(in.Text)) > 4000 {
			problem(w, 400, "Сообщение: от 1 до 4000 символов")
			return true
		}
		before := a.snapshot()
		m := ChannelMessage{id(), seg[1], u, in.Text, now()}
		a.state.Messages = append(a.state.Messages, m)
		if a.commit(before) != nil {
			problem(w, 500, "Сообщение не сохранено")
			return true
		}
		respond(w, 201, m)
		return true
	}
	return false
}
