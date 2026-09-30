package app

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Existing local demos used TXT stand-ins. Replace only untouched synthetic
// examples, preserving comments, decisions and user-uploaded documents.
func (a *App) upgradeDemoExamples() (bool, error) {
	examples := map[string]string{
		"Договор на разработку сайта":     "contract.docx",
		"Счёт на оплату № 148":            "invoice.pdf",
		"Регламент работы команды":        "rules.docx",
		"Акт выполненных работ":           "act.pdf",
		"Бюджет проекта · IV квартал":     "budget.xlsx",
		"Соглашение о конфиденциальности": "nda.pdf",
	}
	mimes := map[string]string{
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".pdf":  "application/pdf",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}
	changed := false
	for i := range a.state.Documents {
		d := &a.state.Documents[i]
		asset, ok := examples[d.Title]
		if !ok || len(d.Versions) != 1 || !strings.HasSuffix(d.Versions[0].Name, ".txt") {
			continue
		}
		file := filepath.Join(a.dir, "files", d.Versions[0].ID)
		old, err := os.ReadFile(file)
		if err != nil {
			return false, err
		}
		if !strings.HasPrefix(string(old), "ДЕМОНСТРАЦИОННЫЙ ДОКУМЕНТ") {
			continue
		}
		next, err := demoAssets.ReadFile("demo_assets/" + asset)
		if err != nil {
			return false, err
		}
		if err := os.WriteFile(file+".upgrade", next, 0600); err != nil {
			return false, err
		}
		if err := os.Rename(file+".upgrade", file); err != nil {
			return false, err
		}
		ext := filepath.Ext(asset)
		d.Versions[0].Name = d.Title + ext
		d.Versions[0].MIME = mimes[ext]
		d.Versions[0].Size = int64(len(next))
		changed = true
	}
	// Add one real report to older demos; fresh demos already contain it.
	hasReport, hasFinance := false, false
	for _, d := range a.state.Documents {
		if d.Title == "Отчёт по продажам" {
			hasReport = true
		}
	}
	for _, c := range a.state.Channels {
		if c.ID == "finance" {
			hasFinance = true
		}
	}
	if !hasReport && hasFinance {
		body, err := demoAssets.ReadFile("demo_assets/sales.csv")
		if err != nil {
			return false, err
		}
		v := Version{ID: id(), Name: "Отчёт по продажам.csv", MIME: "text/csv; charset=utf-8", Size: int64(len(body)), Created: now(), Author: demoUsers["2"].Name}
		if err := os.WriteFile(filepath.Join(a.dir, "files", v.ID), body, 0600); err != nil {
			return false, err
		}
		d := Document{ID: id(), Title: "Отчёт по продажам", ChannelID: "finance", Owner: demoUsers["2"], Status: "review", Due: time.Now().AddDate(0, 0, 7).Format("2006-01-02"), Description: "Проверьте выручку по проектам.", Created: now(), Updated: now(), Versions: []Version{v}, Comments: []Comment{}, Revision: 1}
		a.state.Documents = append(a.state.Documents, d)
		changed = true
	}
	return changed, nil
}
