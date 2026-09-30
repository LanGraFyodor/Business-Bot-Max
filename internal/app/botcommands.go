package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (a *App) botIdentity(update Update, userID string) User {
	first, last := update.Message.Sender.FirstName, update.Message.Sender.LastName
	if update.Type == "bot_started" {
		first, last = update.User.FirstName, update.User.LastName
	}
	name := strings.TrimSpace(first + " " + last)
	if name == "" {
		name = "MAX " + userID
	}
	return User{ID: userID, Name: name, Role: "employee"}
}

// Commands are authenticated by the MAX update sender, not by user-supplied IDs.
func (a *App) botTeamCommand(update Update, userID, command string) (string, bool) {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "", false
	}
	verb := fields[0]
	if verb != "/create" && verb != "/invite" && verb != "/join" && verb != "/role" && verb != "/remove" {
		return "", false
	}
	if a.demo {
		return "Команды реальных аккаунтов доступны при DEMO_MODE=false. Локальное демо откройте в браузере.", true
	}
	u := a.botIdentity(update, userID)
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.registerAccount(u); err != nil {
		return "Профиль MAX не сохранён. Повторите попытку.", true
	}
	switch verb {
	case "/role", "/remove":
		if (verb == "/role" && len(fields) != 4) || (verb == "/remove" && len(fields) != 3) {
			return "Формат: /role ID_команды MAX_ID employee|manager или /remove ID_команды MAX_ID", true
		}
		teamID, target := fields[1], fields[2]
		if a.teamRole(teamID, u.ID) != "manager" {
			return "Изменять состав может только руководитель команды.", true
		}
		newRole := ""
		if verb == "/role" {
			newRole = fields[3]
			if newRole != "employee" && newRole != "manager" {
				return "Роль: employee или manager.", true
			}
		}
		for ti := range a.state.Teams {
			team := &a.state.Teams[ti]
			if team.ID != teamID {
				continue
			}
			managers := 0
			for _, member := range team.Members {
				if member.Role == "manager" {
					managers++
				}
			}
			for mi, member := range team.Members {
				if member.ID != target {
					continue
				}
				if member.Role == "manager" && newRole != "manager" && managers == 1 {
					return "В команде должен остаться руководитель.", true
				}
				before := a.snapshot()
				if verb == "/remove" {
					team.Members = append(team.Members[:mi], team.Members[mi+1:]...)
				} else {
					team.Members[mi].Role = newRole
				}
				if a.commit(before) != nil {
					return "Не удалось сохранить изменение.", true
				}
				if verb == "/remove" {
					return "Участник исключён. Доступ к документам команды закрыт.", true
				}
				return "Роль участника обновлена: " + newRole, true
			}
		}
		return "Участник не найден в этой команде.", true
	case "/create":
		name := strings.TrimSpace(strings.TrimPrefix(command, verb))
		if name == "" || len([]rune(name)) > 40 {
			return "Формат: /create Название команды (до 40 символов)", true
		}
		count := 0
		for _, t := range a.state.Teams {
			if a.teamRole(t.ID, u.ID) == "manager" {
				count++
			}
		}
		if count >= 20 {
			return "Можно управлять не более чем 20 командами.", true
		}
		before := a.snapshot()
		owner := u
		owner.Role = "manager"
		team := Team{ID: id(), Name: name, Color: "violet", Members: []User{owner}, Invites: []Invitation{}}
		a.state.Teams = append(a.state.Teams, team)
		a.state.Channels = append(a.state.Channels, Channel{ID: id(), Name: "общее", Description: "Главный канал команды", Color: "violet", TeamID: team.ID})
		if a.commit(before) != nil {
			return "Команда не сохранена. Повторите попытку.", true
		}
		return fmt.Sprintf("Команда «%s» создана. Вы — руководитель. ID команды: %s\nСоздать приглашение: /invite %s employee [MAX_ID]", name, team.ID, team.ID), true
	case "/invite":
		if len(fields) < 3 || len(fields) > 4 {
			return "Формат: /invite ID_команды employee|manager [MAX_ID]", true
		}
		teamID, role := fields[1], fields[2]
		if role != "employee" && role != "manager" {
			return "Роль: employee (участник) или manager (руководитель).", true
		}
		if a.teamRole(teamID, u.ID) != "manager" {
			return "Приглашать может только руководитель этой команды.", true
		}
		target := ""
		if len(fields) == 4 {
			target = fields[3]
			value, err := strconv.ParseInt(target, 10, 64)
			if err != nil || value <= 0 {
				return "MAX ID должен быть положительным числом.", true
			}
		}
		for i := range a.state.Teams {
			team := &a.state.Teams[i]
			if team.ID != teamID {
				continue
			}
			if len(team.Invites) >= 100 {
				return "Лимит приглашений достигнут.", true
			}
			before := a.snapshot()
			invite := Invitation{ID: id(), Code: id(), Role: role, TargetUserID: target, CreatedBy: u.ID, Created: now(), Expires: time.Now().Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339)}
			team.Invites = append(team.Invites, invite)
			if a.commit(before) != nil {
				return "Приглашение не сохранено. Повторите попытку.", true
			}
			roleName := "участник"
			if role == "manager" {
				roleName = "руководитель"
			}
			targetText := ""
			if target != "" {
				targetText = " Только для MAX ID " + target + "."
			}
			return fmt.Sprintf("Приглашение в «%s»: роль — %s.%s\nКод: %s\nЕсли мини-приложение ещё не подключено, отправьте коллеге: /join %s\nСсылка: https://max.ru/t68_hakaton_max_bot?startapp=i_%s", team.Name, roleName, targetText, invite.Code, invite.Code, invite.Code), true
		}
		return "Команда не найдена.", true
	case "/join":
		if len(fields) != 2 {
			return "Формат: /join КОД_ПРИГЛАШЕНИЯ", true
		}
		code := strings.TrimPrefix(fields[1], "i_")
		team, invite := a.findInvitation(code)
		if invite == nil || !invitationActive(*invite) {
			return "Приглашение не найдено или уже использовано.", true
		}
		if invite.TargetUserID != "" && invite.TargetUserID != u.ID {
			return "Приглашение выписано другому аккаунту MAX.", true
		}
		if a.teamRole(team.ID, u.ID) != "" {
			return "Вы уже состоите в этой команде.", true
		}
		if err := a.acceptInvitation(u, team, invite); err != nil {
			return err.Error(), true
		}
		roleName := "участник"
		if invite.Role == "manager" {
			roleName = "руководитель"
		}
		if len(invite.WorkspaceTeamIDs) > 0 {
			return fmt.Sprintf("Вы присоединились к тестовой компании: %d команд, роль — %s. Доступ до %s. Нажмите «Документы».", len(invite.WorkspaceTeamIDs), roleName, team.ReviewUntil), true
		}
		return fmt.Sprintf("Вы вступили в «%s» как %s. Откройте «Документы».", team.Name, roleName), true
	}
	return "", false
}
