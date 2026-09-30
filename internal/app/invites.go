package app

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Invitation struct {
	MaxUses          int      `json:"maxUses,omitempty"`
	UseCount         int      `json:"useCount,omitempty"`
	WorkspaceTeamIDs []string `json:"workspaceTeamIds,omitempty"`
	ID               string   `json:"id"`
	Code             string   `json:"code"`
	Role             string   `json:"role"`
	TargetUserID     string   `json:"targetUserId,omitempty"`
	CreatedBy        string   `json:"createdBy"`
	Created          string   `json:"created"`
	Expires          string   `json:"expires"`
	UsedBy           string   `json:"usedBy,omitempty"`
	UsedAt           string   `json:"usedAt,omitempty"`
	RevokedAt        string   `json:"revokedAt,omitempty"`
}

func (a *App) inviteAPI(w http.ResponseWriter, r *http.Request, u User, p string) bool {
	seg := strings.Split(p, "/")
	if len(seg) == 2 && seg[0] == "invites" && r.Method == http.MethodGet {
		a.previewInvitation(w, u, seg[1])
		return true
	}
	if len(seg) == 3 && seg[0] == "teams" && seg[2] == "invites" && r.Method == http.MethodPost {
		if a.teamRole(seg[1], u.ID) != "manager" {
			problem(w, http.StatusForbidden, "Приглашать участников может руководитель команды")
			return true
		}
		var in struct {
			Role         string `json:"role"`
			TargetUserID string `json:"targetUserId"`
		}
		if !decode(w, r, &in) {
			return true
		}
		if in.Role != "manager" && in.Role != "employee" {
			problem(w, http.StatusBadRequest, "Выберите роль приглашения")
			return true
		}
		target := strings.TrimSpace(in.TargetUserID)
		if target != "" {
			number, err := strconv.ParseInt(target, 10, 64)
			if err != nil || number <= 0 {
				problem(w, http.StatusBadRequest, "MAX ID должен быть положительным числом")
				return true
			}
		}
		for i := range a.state.Teams {
			t := &a.state.Teams[i]
			if t.ID != seg[1] {
				continue
			}
			if len(t.Invites) >= 100 {
				problem(w, http.StatusBadRequest, "Лимит приглашений команды достигнут")
				return true
			}
			before := a.snapshot()
			invite := Invitation{ID: id(), Code: id(), Role: in.Role, TargetUserID: target, CreatedBy: u.ID, Created: now(), Expires: time.Now().Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339)}
			t.Invites = append(t.Invites, invite)
			if a.commit(before) != nil {
				problem(w, 500, "Не удалось создать приглашение")
				return true
			}
			respond(w, http.StatusCreated, invite)
			return true
		}
		problem(w, http.StatusNotFound, "Команда не найдена")
		return true
	}
	if len(seg) == 4 && seg[0] == "teams" && seg[2] == "invites" && r.Method == http.MethodDelete {
		if a.teamRole(seg[1], u.ID) != "manager" {
			problem(w, http.StatusForbidden, "Недостаточно прав")
			return true
		}
		for i := range a.state.Teams {
			t := &a.state.Teams[i]
			if t.ID != seg[1] {
				continue
			}
			for j := range t.Invites {
				invite := &t.Invites[j]
				if invite.ID != seg[3] {
					continue
				}
				if invite.RevokedAt != "" || (invite.UsedAt != "" && invite.MaxUses == 0) {
					problem(w, 409, "Приглашение уже недействительно")
					return true
				}
				before := a.snapshot()
				invite.RevokedAt = now()
				if a.commit(before) != nil {
					problem(w, 500, "Не удалось отозвать приглашение")
					return true
				}
				respond(w, 200, invite)
				return true
			}
		}
		problem(w, 404, "Приглашение не найдено")
		return true
	}
	return false
}

func (a *App) findInvitation(code string) (*Team, *Invitation) {
	for i := range a.state.Teams {
		for j := range a.state.Teams[i].Invites {
			if a.state.Teams[i].Invites[j].Code == code {
				return &a.state.Teams[i], &a.state.Teams[i].Invites[j]
			}
		}
	}
	return nil, nil
}
func invitationActive(invite Invitation) bool {
	expiry, err := time.Parse(time.RFC3339, invite.Expires)
	return err == nil && time.Now().Before(expiry) && (invite.UsedAt == "" || (invite.MaxUses > 0 && invite.UseCount < invite.MaxUses)) && invite.RevokedAt == ""
}
func (a *App) previewInvitation(w http.ResponseWriter, u User, code string) {
	team, invite := a.findInvitation(code)
	if invite == nil || !invitationActive(*invite) {
		problem(w, 404, "Приглашение недействительно")
		return
	}
	if invite.TargetUserID != "" && invite.TargetUserID != u.ID {
		problem(w, 403, "Приглашение выписано другому аккаунту MAX")
		return
	}
	respond(w, 200, map[string]any{"teamId": team.ID, "teamName": team.Name, "role": invite.Role, "expires": invite.Expires, "alreadyMember": a.teamRole(team.ID, u.ID) != ""})
}
func (a *App) joinInvitation(w http.ResponseWriter, u User, code string) bool {
	team, invite := a.findInvitation(code)
	if invite == nil {
		return false
	}
	if !invitationActive(*invite) {
		problem(w, 404, "Приглашение недействительно")
		return true
	}
	if invite.TargetUserID != "" && invite.TargetUserID != u.ID {
		problem(w, 403, "Приглашение выписано другому аккаунту MAX")
		return true
	}
	if a.teamRole(team.ID, u.ID) != "" {
		problem(w, 409, "Вы уже состоите в этой команде. Роль может изменить руководитель")
		return true
	}
	if err := a.acceptInvitation(u, team, invite); err != nil {
		problem(w, 409, err.Error())
		return true
	}
	respond(w, 200, a.publicTeam(*team, u))
	return true
}

// Caller holds a.mu. Workspace invitations are generated offline for the isolated
// review copy; ordinary HTTP invitation creation never accepts workspaceTeamIds.
func (a *App) acceptInvitation(u User, root *Team, invite *Invitation) error {
	if !invitationActive(*invite) || !teamAccessActive(*root) {
		return errors.New("Приглашение или срок тестирования истёк")
	}
	targets := []*Team{root}
	if len(invite.WorkspaceTeamIDs) > 0 {
		targets = nil
		seen := map[string]bool{}
		for _, teamID := range invite.WorkspaceTeamIDs {
			if seen[teamID] {
				return errors.New("Некорректное приглашение компании")
			}
			seen[teamID] = true
			var found *Team
			for i := range a.state.Teams {
				if a.state.Teams[i].ID == teamID {
					found = &a.state.Teams[i]
					break
				}
			}
			if found == nil || root.ReviewUntil == "" || found.ReviewUntil != root.ReviewUntil || !teamAccessActive(*found) {
				return errors.New("Приглашение не соответствует тестовой компании")
			}
			targets = append(targets, found)
		}
		if !seen[root.ID] {
			return errors.New("В приглашении нет основной команды")
		}
	}
	for _, team := range targets {
		for _, member := range team.Members {
			if member.ID == u.ID {
				return errors.New("Вы уже состоите в команде. Для второй роли используйте другой MAX-аккаунт")
			}
		}
	}
	before := a.snapshot()
	member := u
	member.Role = invite.Role
	for _, team := range targets {
		team.Members = append(team.Members, member)
	}
	invite.UsedBy, invite.UsedAt = u.ID, now()
	invite.UseCount++
	if a.commit(before) != nil {
		return errors.New("Не удалось сохранить приглашение")
	}
	return nil
}
