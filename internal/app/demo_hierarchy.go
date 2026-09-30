package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// The second additive demo migration makes the reporting chain visible without
// replacing documents, decisions, or memberships created by a tester.
func (a *App) expandDemoHierarchy() (bool, error) {
	if !a.demo || a.state.DemoSeedVersion >= 2 {
		return false, nil
	}
	parents := map[string]string{"studio": "demo-sales", "demo-fin": "orbita", "demo-sales": "orbita", "demo-law": "orbita", "demo-ops": "orbita", "demo-hr": "orbita"}
	members := map[string][]string{
		"orbita":     {"4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
		"demo-fin":   {"13", "14", "15", "16", "17", "18", "19", "20"},
		"demo-sales": {"3", "13", "16", "21", "22", "23", "24", "25"},
		"studio":     {"8", "21", "22", "26", "27", "28", "29", "30"},
		"demo-law":   {"14", "18", "23", "25", "26", "27", "29", "30"},
		"demo-ops":   {"13", "15", "17", "19", "21", "24", "28", "30"},
		"demo-hr":    {"16", "17", "18", "20", "22", "24", "26", "29"},
	}
	for i := range a.state.Teams {
		t := &a.state.Teams[i]
		if t.ParentTeamID == "" {
			t.ParentTeamID = parents[t.ID]
		}
		existing := map[string]bool{}
		for i, member := range t.Members {
			if t.ID == "studio" && member.ID == "3" {
				t.Members[i].Role = "manager"
			}
			existing[member.ID] = true
		}
		for _, uid := range members[t.ID] {
			if existing[uid] {
				continue
			}
			person := demoUsers[uid]
			person.Role = "employee"
			t.Members = append(t.Members, person)
			existing[uid] = true
		}
	}
	samples := []struct{ team, channel, source, title, asset string }{
		{"demo-fin", "demo-fin-invoices", "1c", "1С · счёт поставщика № 314", "invoice.pdf"},
		{"demo-fin", "demo-fin-acts", "1c", "1С · акт сверки за сентябрь", "act.pdf"},
		{"demo-sales", "demo-sales-pipeline", "bitrix24", "Битрикс24 · отчёт по сделкам", "sales.csv"},
		{"demo-sales", "demo-sales-offers", "amocrm", "amoCRM · предложение клиенту", "contract.docx"},
	}
	for _, sample := range samples {
		found := false
		for _, d := range a.state.Documents {
			if d.Title == sample.title && d.ChannelID == sample.channel {
				found = true
				break
			}
		}
		if found {
			continue
		}
		body, err := demoAssets.ReadFile("demo_assets/" + sample.asset)
		if err != nil {
			return false, err
		}
		if sample.asset == "sales.csv" {
			body, err = demoCSV(sample.title, 3)
			if err != nil {
				return false, err
			}
		}
		version := Version{ID: id(), Name: sample.asset, MIME: mustDemoMIME(sample.asset), Size: int64(len(body)), Created: now(), Author: sample.source}
		if err := os.WriteFile(filepath.Join(a.dir, "files", version.ID), body, 0600); err != nil {
			return false, err
		}
		owner := demoUsers["13"]
		if sample.team == "demo-sales" {
			owner = demoUsers["21"]
		}
		owner.Role = "employee"
		d := Document{ID: id(), Title: sample.title, ChannelID: sample.channel, Owner: owner, Status: "review", Due: time.Now().AddDate(0, 0, 5).Format("2006-01-02"), Description: "Демонстрационная выгрузка. Показывает, как документ из внешней системы поступает в канал на согласование.", Created: now(), Updated: now(), Versions: []Version{version}, Comments: []Comment{{ID: id(), Author: owner, Text: "Пример импорта: проверьте документ и примите решение.", Created: now(), System: true}}, Revision: 1, SourceSystem: sample.source, ImportMethod: "demo", ExternalID: fmt.Sprintf("demo-%s-%s", sample.source, sample.channel)}
		a.state.Documents = append(a.state.Documents, d)
	}
	a.state.DemoSeedVersion = 2
	return true, nil
}

func mustDemoMIME(filename string) string { mime, _ := fileMIME(filename); return mime }

func init() {
	for uid, name := range map[string]string{
		"13": "Артём Зайцев", "14": "Вера Павлова", "15": "Никита Морозов", "16": "Дарья Фролова", "17": "Илья Миронов", "18": "Полина Тихонова", "19": "Максим Егоров", "20": "Алина Сергеева", "21": "Тимур Лебедев", "22": "Ксения Власова", "23": "Виктор Беляев", "24": "Екатерина Козлова", "25": "Андрей Сидоров", "26": "Мария Николаева", "27": "Лев Комаров", "28": "Наталья Цветкова", "29": "Глеб Соловьёв", "30": "Юлия Васильева",
	} {
		demoUsers[uid] = User{ID: uid, Name: name, Role: "employee"}
	}
}
