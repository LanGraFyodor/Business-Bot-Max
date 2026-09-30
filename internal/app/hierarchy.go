package app

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Escalation creates a separate review item in the parent team. Membership in
// a child team never grants access to the parent, or vice versa.
func (a *App) escalate(w http.ResponseWriter, r *http.Request, u User, source Document) {
	childID := a.channelTeam(source.ChannelID)
	if a.teamRole(childID, u.ID) != "manager" || source.Status != "approved" {
		problem(w, 403, "Передать наверх может руководитель после согласования")
		return
	}
	parentID := ""
	childName := ""
	for _, t := range a.state.Teams {
		if t.ID == childID {
			parentID = t.ParentTeamID
			childName = t.Name
			break
		}
	}
	if parentID == "" || a.teamRole(parentID, u.ID) == "" {
		problem(w, 403, "Нужен доступ к родительской команде")
		return
	}
	var in struct {
		ChannelID string `json:"channelId"`
	}
	if !decode(w, r, &in) {
		return
	}
	if a.channelTeam(in.ChannelID) != parentID {
		problem(w, 400, "Выберите канал родительской команды")
		return
	}
	for _, d := range a.state.Documents {
		if d.EscalatedFrom == source.ID && d.EscalatedRevision == source.Revision {
			respond(w, 200, d)
			return
		}
	}
	if len(source.Versions) == 0 {
		problem(w, 400, "У документа нет файла")
		return
	}
	latest := source.Versions[len(source.Versions)-1]
	body, err := os.ReadFile(filepath.Join(a.dir, "files", latest.ID))
	if err != nil {
		problem(w, 500, "Исходный файл недоступен")
		return
	}
	version := latest
	version.ID = id()
	version.Created = now()
	version.Author = u.Name
	targetPath := filepath.Join(a.dir, "files", version.ID)
	if err = os.WriteFile(targetPath, body, 0600); err != nil {
		problem(w, 500, "Не удалось скопировать файл")
		return
	}
	owner := u
	owner.Role = a.teamRole(parentID, u.ID)
	stamp := now()
	d := Document{ID: id(), Title: source.Title, ChannelID: in.ChannelID, Owner: owner, Status: "review", Due: source.Due, Description: "Передано из команды «" + childName + "» на следующий уровень согласования. " + source.Description, Created: stamp, Updated: stamp, Versions: []Version{version}, Comments: []Comment{{ID: id(), Author: owner, Text: "Документ согласован в команде «" + childName + "» и передан руководству.", Created: stamp, System: true}}, Revision: 1, SourceSystem: source.SourceSystem, ImportMethod: source.ImportMethod, EscalatedFrom: source.ID, EscalatedRevision: source.Revision}
	before := a.snapshot()
	a.state.Documents = append([]Document{d}, a.state.Documents...)
	a.queueNotification(d)
	if a.commit(before) != nil {
		_ = os.Remove(targetPath)
		problem(w, 500, "Не удалось передать документ")
		return
	}
	respond(w, 201, d)
}

func (a *App) demoImportAPI(w http.ResponseWriter, r *http.Request, u User, p string) bool {
	if p != "demo/integrations/sample" {
		return false
	}
	if !a.demo {
		problem(w, 404, "Доступно только в демо")
		return true
	}
	if r.Method != "POST" {
		problem(w, 405, "Метод не поддерживается")
		return true
	}
	var in struct {
		ChannelID string `json:"channelId"`
		Provider  string `json:"provider"`
	}
	if !decode(w, r, &in) {
		return true
	}
	teamID := a.channelTeam(in.ChannelID)
	if teamID == "" || a.teamRole(teamID, u.ID) != "manager" {
		problem(w, 403, "Только руководитель этой команды может запускать пример")
		return true
	}
	type sample struct{ asset, title, source string }
	samples := map[string]sample{"1c": {"invoice.pdf", "1С · новый счёт поставщика", "1c"}, "bitrix24": {"sales.csv", "Битрикс24 · новый отчёт по сделкам", "bitrix24"}, "amocrm": {"contract.docx", "amoCRM · новое предложение", "amocrm"}}
	spec, ok := samples[in.Provider]
	if !ok {
		problem(w, 400, "Выберите 1С, Битрикс24 или amoCRM")
		return true
	}
	body, err := demoAssets.ReadFile("demo_assets/" + spec.asset)
	if err != nil {
		problem(w, 500, "Пример файла не найден")
		return true
	}
	if spec.asset == "sales.csv" {
		body, err = demoCSV(spec.title, 4)
		if err != nil {
			problem(w, 500, "Не удалось подготовить пример")
			return true
		}
	}
	version := Version{ID: id(), Name: spec.asset, MIME: mustDemoMIME(spec.asset), Size: int64(len(body)), Created: now(), Author: spec.source}
	if err = os.WriteFile(filepath.Join(a.dir, "files", version.ID), body, 0600); err != nil {
		problem(w, 500, "Не удалось сохранить пример")
		return true
	}
	owner := u
	owner.Role = a.teamRole(teamID, u.ID)
	stamp := now()
	d := Document{ID: id(), Title: spec.title + " · " + time.Now().Format("15:04:05"), ChannelID: in.ChannelID, Owner: owner, Status: "review", Due: time.Now().AddDate(0, 0, 3).Format("2006-01-02"), Description: "Демонстрационная выгрузка из " + strings.ToUpper(spec.source) + ". Это пример входящего импорта, без соединения с внешней системой.", Created: stamp, Updated: stamp, Versions: []Version{version}, Comments: []Comment{{ID: id(), Author: owner, Text: "Демо: файл поступил в команду на согласование.", Created: stamp, System: true}}, Revision: 1, SourceSystem: spec.source, ImportMethod: "demo"}
	before := a.snapshot()
	a.state.Documents = append([]Document{d}, a.state.Documents...)
	a.queueNotification(d)
	if a.commit(before) != nil {
		_ = os.Remove(filepath.Join(a.dir, "files", version.ID))
		problem(w, 500, "Не удалось сохранить документ")
		return true
	}
	respond(w, 201, d)
	return true
}
