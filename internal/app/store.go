package app

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"

	"os"
	"path/filepath"
	"strings"
	"time"
)

func id() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func now() string { return time.Now().UTC().Format(time.RFC3339) }
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func ids(s string) map[string]bool {
	m := map[string]bool{}
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			m[v] = true
		}
	}
	return m
}
func newApp(dir string, demo bool) (*App, error) {
	a := &App{dir: dir, demo: demo, token: os.Getenv("MAX_BOT_TOKEN"), members: ids(os.Getenv("MAX_MEMBER_IDS")), managers: ids(os.Getenv("MAX_MANAGER_IDS")), client: nil}
	client, clientErr := newMAXClient()
	if clientErr != nil {
		return nil, clientErr
	}
	a.client = client
	if err := os.MkdirAll(filepath.Join(dir, "files"), 0700); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err == nil {
		err = json.Unmarshal(b, &a.state)
		if err != nil {
			return nil, err
		}
		if !demo && len(a.state.Teams) == 1 && a.state.Teams[0].ID == "orbita" && len(a.state.Teams[0].Members) == 0 && len(a.state.Channels) == 0 && len(a.state.Documents) == 0 && len(a.state.Messages) == 0 {
			a.state.Teams = []Team{}
			return a, a.save()
		}
		migrated := a.migrateTeams()
		upgraded := false
		expanded := false
		hierarchy := false
		reports := false
		if demo {
			upgraded, err = a.upgradeDemoExamples()
			if err != nil {
				return nil, err
			}
			expanded, err = a.expandDemoCompany()
			if err != nil {
				return nil, err
			}
			hierarchy, err = a.expandDemoHierarchy()
			if err != nil {
				return nil, err
			}
			reports, err = a.expandDemoReports()
			if err != nil {
				return nil, err
			}
		}
		if migrated || upgraded || expanded || hierarchy || reports {
			return a, a.save()
		}
		return a, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	if !demo {
		a.state = State{Teams: []Team{}, Channels: []Channel{}, Documents: []Document{}, Messages: []ChannelMessage{}, Accounts: []Account{}}
		return a, a.save()
	}
	a.state = State{Channels: []Channel{{ID: "general", Name: "общее", Description: "Рабочие вопросы и документы", Color: "blue"}, {ID: "finance", Name: "финансы", Description: "Счета, акты и отчёты", Color: "green"}, {ID: "legal", Name: "договоры", Description: "Договоры и согласования", Color: "violet"}, {ID: "people", Name: "команда", Description: "Регламенты и кадровые документы", Color: "coral"}}, Documents: []Document{}}
	if demo {
		if err = a.seed(); err != nil {
			return nil, err
		}
	}
	a.migrateTeams()
	if demo {
		if _, err := a.expandDemoCompany(); err != nil {
			return nil, err
		}
		if _, err := a.expandDemoHierarchy(); err != nil {
			return nil, err
		}
		if _, err := a.expandDemoReports(); err != nil {
			return nil, err
		}
	}
	return a, a.save()
}
func (a *App) save() error {
	b, e := json.MarshalIndent(a.state, "", "  ")
	if e != nil {
		return e
	}
	p := filepath.Join(a.dir, "state.json")
	if e = os.WriteFile(p+".tmp", b, 0600); e != nil {
		return e
	}
	return os.Rename(p+".tmp", p)
}
func (a *App) commit(before State) error {
	if e := a.save(); e != nil {
		a.state = before
		return e
	}
	return nil
}
func (a *App) snapshot() State {
	b, _ := json.Marshal(a.state)
	var s State
	_ = json.Unmarshal(b, &s)
	return s
}

//go:embed demo_assets/*
var demoAssets embed.FS

var demoUsers = map[string]User{"1": {"1", "Александр Морозов", "manager"}, "2": {"2", "Анна Смирнова", "employee"}, "3": {"3", "Михаил Волков", "employee"}}

func (a *App) seed() error {
	titles := []string{"Договор на разработку сайта", "Счёт на оплату № 148", "Регламент работы команды", "Акт выполненных работ", "Бюджет проекта · IV квартал", "Соглашение о конфиденциальности", "Отчёт по продажам"}
	assets := []string{"contract.docx", "invoice.pdf", "rules.docx", "act.pdf", "budget.xlsx", "nda.pdf", "sales.csv"}
	channels := []string{"legal", "finance", "people", "general", "finance", "legal", "finance"}
	statuses := []string{"review", "review", "approved", "draft", "changes", "approved", "review"}
	mimes := map[string]string{".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document", ".pdf": "application/pdf", ".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", ".csv": "text/csv; charset=utf-8"}
	for i, t := range titles {
		stamp := time.Now().Add(-time.Duration(i+1) * 3 * time.Hour).UTC().Format(time.RFC3339)
		asset := assets[i]
		body, err := demoAssets.ReadFile("demo_assets/" + asset)
		if err != nil {
			return err
		}
		v := Version{ID: id(), Name: t + filepath.Ext(asset), MIME: mimes[filepath.Ext(asset)], Size: int64(len(body)), Created: stamp, Author: "Анна Смирнова"}
		if err := os.WriteFile(filepath.Join(a.dir, "files", v.ID), body, 0600); err != nil {
			return err
		}
		a.state.Documents = append(a.state.Documents, Document{ID: id(), Title: t, ChannelID: channels[i], Owner: demoUsers["2"], Status: statuses[i], Due: time.Now().AddDate(0, 0, i+1).Format("2006-01-02"), Description: "Проверьте условия и реквизиты перед согласованием.", Created: stamp, Updated: stamp, Versions: []Version{v}, Comments: []Comment{{ID: id(), Author: demoUsers["2"], Text: "Коллеги, документ готов к проверке. Буду рада обратной связи!", Created: stamp}}, Revision: 1})
	}
	return nil
}
