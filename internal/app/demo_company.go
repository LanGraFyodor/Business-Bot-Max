package app

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type demoTeamSpec struct {
	id, name, color, description, manager string
	employees                             []string
}
type demoChannelSpec struct {
	team, id, name, description, title, asset string
}

var companyTeams = []demoTeamSpec{
	{"demo-fin", "Финансовый блок", "green", "Бюджеты, счета, акты и управленческая отчётность", "4", []string{"8", "11", "12"}},
	{"demo-sales", "Продажи", "blue", "Клиенты, сделки, предложения и выполнение плана", "5", []string{"8", "10", "12"}},
	{"demo-law", "Юридический отдел", "violet", "Договоры, NDA, претензии и внутренние регламенты", "6", []string{"9", "11"}},
	{"demo-ops", "Операции", "coral", "Закупки, поставки, проекты и контроль качества", "7", []string{"3", "10", "12"}},
	{"demo-hr", "HR и развитие", "blue", "Онбординг, обучение, графики и кадровые документы", "9", []string{"4", "11", "12"}},
}

var companyChannels = []demoChannelSpec{
	{"orbita", "demo-hq-strategy", "стратегия", "Решения и планы руководства", "Протокол стратегической сессии — %s", "rules.docx"},
	{"orbita", "demo-hq-approvals", "решения", "Документы на уровне управляющей команды", "Решение комитета по проекту «%s»", "act.pdf"},
	{"orbita", "demo-hq-projects", "проекты", "Ключевые инициативы компании", "План проекта «%s»", "contract.docx"},
	{"orbita", "demo-hq-reports", "сводки", "Сводные отчёты подразделений", "Сводка результатов — %s", "sales.csv"},
	{"orbita", "demo-hq-partners", "партнёры", "Контракты и встречи с партнёрами", "Партнёрское соглашение — %s", "contract.docx"},
	{"studio", "demo-studio-brand", "бренд", "Айдентика и стандарты бренда", "Бренд-гайд проекта «%s»", "rules.docx"},
	{"studio", "demo-studio-briefs", "брифы", "Задачи на дизайн и исследования", "Бриф на дизайн — %s", "contract.docx"},
	{"studio", "demo-studio-assets", "макеты", "Сметы и материалы дизайн-команды", "Смета дизайн-проекта «%s»", "budget.xlsx"},
	{"studio", "demo-studio-reviews", "приёмка", "Сдача и приёмка материалов", "Акт приёмки дизайна — %s", "act.pdf"},
	{"demo-fin", "demo-fin-invoices", "счета", "Оплата поставщиков и подрядчиков", "Счёт поставщика — %s", "invoice.pdf"},
	{"demo-fin", "demo-fin-acts", "акты", "Закрывающие документы", "Акт оказанных услуг — %s", "act.pdf"},
	{"demo-fin", "demo-fin-budgets", "бюджеты", "Лимиты и заявки подразделений", "Бюджет подразделения — %s", "budget.xlsx"},
	{"demo-fin", "demo-fin-cashflow", "денежный поток", "Платёжный календарь и движение средств", "ДДС по направлению — %s", "sales.csv"},
	{"demo-fin", "demo-fin-reconcile", "сверки", "Сверка взаиморасчётов", "Акт сверки с контрагентом — %s", "act.pdf"},
	{"demo-sales", "demo-sales-offers", "предложения", "Коммерческие предложения клиентам", "Коммерческое предложение — %s", "contract.docx"},
	{"demo-sales", "demo-sales-contracts", "договоры", "Договоры и приложения к сделкам", "Договор поставки — %s", "contract.docx"},
	{"demo-sales", "demo-sales-pipeline", "воронка", "Аналитика продаж и прогноз", "Отчёт по воронке — %s", "sales.csv"},
	{"demo-sales", "demo-sales-delivery", "отгрузки", "Закрытие сделок и отгрузка", "Акт передачи заказа — %s", "act.pdf"},
	{"demo-sales", "demo-sales-billing", "выставление счетов", "Счета клиентам", "Счёт клиенту — %s", "invoice.pdf"},
	{"demo-law", "demo-law-contracts", "контракты", "Юридическая экспертиза договоров", "Экспертиза договора — %s", "contract.docx"},
	{"demo-law", "demo-law-nda", "NDA", "Соглашения о конфиденциальности", "NDA с партнёром — %s", "nda.pdf"},
	{"demo-law", "demo-law-policies", "политики", "Внутренние положения и регламенты", "Регламент работы с данными — %s", "rules.docx"},
	{"demo-law", "demo-law-claims", "претензии", "Ответы и досудебные документы", "Позиция по претензии — %s", "contract.docx"},
	{"demo-law", "demo-law-risk", "риски", "Оценка правовых рисков проектов", "Правовое заключение — %s", "nda.pdf"},
	{"demo-ops", "demo-ops-projects", "исполнение", "Планы исполнения проектов", "План исполнения — %s", "rules.docx"},
	{"demo-ops", "demo-ops-procurement", "закупки", "Заявки и счета на закупку", "Заявка на закупку — %s", "invoice.pdf"},
	{"demo-ops", "demo-ops-acceptance", "приёмка", "Акты приёмки работ", "Акт приёмки работ — %s", "act.pdf"},
	{"demo-ops", "demo-ops-resources", "ресурсы", "Ресурсные планы и сметы", "Ресурсный план — %s", "budget.xlsx"},
	{"demo-ops", "demo-ops-quality", "качество", "Контроль качества и инциденты", "Отчёт по качеству — %s", "sales.csv"},
	{"demo-hr", "demo-hr-onboarding", "онбординг", "Ввод новых сотрудников", "План адаптации — %s", "rules.docx"},
	{"demo-hr", "demo-hr-policies", "регламенты", "Политики и правила компании", "Кадровый регламент — %s", "rules.docx"},
	{"demo-hr", "demo-hr-learning", "обучение", "Программы развития и обучение", "План обучения — %s", "budget.xlsx"},
	{"demo-hr", "demo-hr-schedule", "графики", "Планирование смен и отпусков", "График команды — %s", "sales.csv"},
	{"demo-hr", "demo-hr-confidential", "персональные данные", "Согласия и обязательства", "Обязательство о конфиденциальности — %s", "nda.pdf"},
}

var companyTopics = []string{"Альфа", "Вектор", "Горизонт", "Нева", "Север"}

func demoCSV(title string, revision int) ([]byte, error) {
	var b bytes.Buffer
	b.Write([]byte{0xef, 0xbb, 0xbf})
	w := csv.NewWriter(&b)
	rows := [][]string{
		{"Отчёт", title},
		{"Показатель", "План", "Факт"},
		{"Выручка", fmt.Sprint(1250000 + revision*11000), fmt.Sprint(1190000 + revision*12000)},
		{"Расходы", fmt.Sprint(420000 + revision*3000), fmt.Sprint(397000 + revision*5000)},
		{"Проектов завершено", fmt.Sprint(12 + revision), fmt.Sprint(11 + revision)},
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return b.Bytes(), w.Error()
}

// Additive, versioned demo migration: existing uploads and decisions stay untouched.
func (a *App) expandDemoCompany() (bool, error) {
	if !a.demo || a.state.DemoSeedVersion >= 1 {
		return false, nil
	}
	teams := map[string]Team{}
	for _, t := range a.state.Teams {
		teams[t.ID] = t
	}
	for _, spec := range companyTeams {
		if _, exists := teams[spec.id]; exists {
			continue
		}
		members := []User{}
		for _, uid := range []string{"1", spec.manager} {
			if uid == "1" && spec.manager == "1" {
				continue
			}
			member := demoUsers[uid]
			member.Role = "manager"
			members = append(members, member)
		}
		for _, uid := range spec.employees {
			member := demoUsers[uid]
			member.Role = "employee"
			members = append(members, member)
		}
		team := Team{ID: spec.id, Name: spec.name, Color: spec.color, Description: spec.description, Members: members, Invites: []Invitation{}}
		a.state.Teams = append(a.state.Teams, team)
		teams[team.ID] = team
	}
	channels := map[string]bool{}
	for _, c := range a.state.Channels {
		channels[c.ID] = true
	}
	existing := map[string]bool{}
	for _, d := range a.state.Documents {
		existing[d.ChannelID+"|"+d.Title] = true
	}
	assets := map[string][]byte{}
	for _, spec := range companyChannels {
		team, ok := teams[spec.team]
		if !ok {
			continue
		}
		if !channels[spec.id] {
			a.state.Channels = append(a.state.Channels, Channel{ID: spec.id, TeamID: spec.team, Name: spec.name, Description: spec.description, Color: team.Color})
			channels[spec.id] = true
		}
		body, ok := assets[spec.asset]
		if !ok {
			var err error
			body, err = demoAssets.ReadFile("demo_assets/" + spec.asset)
			if err != nil {
				return false, err
			}
			assets[spec.asset] = body
		}
		employees := []User{}
		manager := team.Members[0]
		for _, member := range team.Members {
			if member.Role == "manager" {
				manager = member
			} else {
				employees = append(employees, member)
			}
		}
		if len(employees) == 0 {
			continue
		}
		for i, topic := range companyTopics {
			title := fmt.Sprintf(spec.title, topic)
			if existing[spec.id+"|"+title] {
				continue
			}
			owner := employees[i%len(employees)]
			status := []string{"review", "approved", "changes", "draft", "approved"}[i]
			stamp := time.Now().Add(-time.Duration(48+len(a.state.Documents)%450) * time.Hour).UTC().Format(time.RFC3339)
			due := time.Now().AddDate(0, 0, []int{-2, 4, 8, 15, -1}[i]).Format("2006-01-02")
			filename := title + filepath.Ext(spec.asset)
			mime, _ := fileMIME(filename)
			fileBody := body
			if spec.asset == "sales.csv" {
				var err error
				fileBody, err = demoCSV(title, i)
				if err != nil {
					return false, err
				}
			}
			version := Version{ID: id(), Name: filename, MIME: mime, Size: int64(len(fileBody)), Created: stamp, Author: owner.Name}
			if err := os.WriteFile(filepath.Join(a.dir, "files", version.ID), fileBody, 0600); err != nil {
				return false, err
			}
			versions := []Version{version}
			updated := stamp
			if spec.asset == "sales.csv" && i == 2 {
				revised, err := demoCSV(title, i+5)
				if err != nil {
					return false, err
				}
				extra := Version{ID: id(), Name: filename, MIME: mime, Size: int64(len(revised)), Created: time.Now().Add(-25 * time.Hour).UTC().Format(time.RFC3339), Author: owner.Name}
				if err := os.WriteFile(filepath.Join(a.dir, "files", extra.ID), revised, 0600); err != nil {
					return false, err
				}
				versions = append(versions, extra)
				updated = extra.Created
			}
			comments := []Comment{{ID: id(), Author: owner, Text: "Материалы и исходный файл загружены. Прошу проверить данные и оставить замечания здесь.", Created: stamp}}
			if status == "approved" {
				comments = append(comments, Comment{ID: id(), Author: manager, Text: "Согласовано. Итоговая версия принята в работу.", Created: stamp, System: true})
			}
			if status == "changes" {
				comments = append(comments, Comment{ID: id(), Author: manager, Text: "Вернуть на доработку: уточнить реквизиты и сроки.", Created: stamp, System: true})
			}
			d := Document{ID: id(), Title: title, ChannelID: spec.id, Owner: owner, Status: status, Due: due, Description: spec.description + ". Проверьте содержание, реквизиты и срок.", Created: stamp, Updated: updated, Versions: versions, Comments: comments, Revision: 1}
			a.state.Documents = append(a.state.Documents, d)
			existing[spec.id+"|"+title] = true
		}
		a.state.Messages = append(a.state.Messages,
			ChannelMessage{ID: id(), ChannelID: spec.id, Author: employees[0], Text: "Документы по этому направлению собраны в канале. Замечания оставляйте в карточке нужного файла.", Created: time.Now().Add(-72 * time.Hour).UTC().Format(time.RFC3339)},
			ChannelMessage{ID: id(), ChannelID: spec.id, Author: manager, Text: "На этой неделе сначала разбираем документы на согласовании и просроченные задачи.", Created: time.Now().Add(-50 * time.Hour).UTC().Format(time.RFC3339)})
	}
	a.state.DemoSeedVersion = 1
	return true, nil
}

func init() {
	for id, user := range map[string]User{
		"4":  {ID: "4", Name: "Ольга Петрова", Role: "employee"},
		"5":  {ID: "5", Name: "Денис Кузнецов", Role: "employee"},
		"6":  {ID: "6", Name: "Елена Соколова", Role: "employee"},
		"7":  {ID: "7", Name: "Павел Орлов", Role: "employee"},
		"8":  {ID: "8", Name: "Марина Белова", Role: "employee"},
		"9":  {ID: "9", Name: "София Крылова", Role: "employee"},
		"10": {ID: "10", Name: "Кирилл Волков", Role: "employee"},
		"11": {ID: "11", Name: "Ирина Романова", Role: "employee"},
		"12": {ID: "12", Name: "Роман Фёдоров", Role: "employee"},
	} {
		demoUsers[id] = user
	}
}
