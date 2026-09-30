package app

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Add examples of actual tabular report shapes, without touching user work.
func (a *App) expandDemoReports() (bool, error) {
	if !a.demo || a.state.DemoSeedVersion >= 3 {
		return false, nil
	}
	specs := []struct{ source, channel, title, owner string }{
		{"1c", "demo-fin-cashflow", "Движение денежных средств · сентябрь", "13"},
		{"1c", "demo-fin-budgets", "Исполнение бюджета подразделений", "13"},
		{"1c", "demo-fin-reconcile", "Оборотно-сальдовая ведомость · пример", "13"},
		{"1c", "demo-fin-invoices", "Реестр платежей поставщикам", "13"},
		{"bitrix24", "demo-sales-pipeline", "Воронка продаж · итоги недели", "21"},
		{"bitrix24", "demo-sales-offers", "План и факт по менеджерам", "21"},
		{"bitrix24", "demo-sales-contracts", "Реестр сделок на согласовании", "21"},
		{"amocrm", "demo-sales-pipeline", "Сделки и бюджет по этапам", "21"},
		{"amocrm", "demo-sales-offers", "Результаты менеджеров · сентябрь", "21"},
		{"amocrm", "demo-sales-contracts", "Клиентский портфель · активные сделки", "21"},
		{"other", "demo-ops-quality", "ERP · исполнение заказов", "13"},
		{"other", "demo-ops-resources", "BI · ресурсы подразделений", "13"},
	}
	for i, spec := range specs {
		exists := false
		for _, d := range a.state.Documents {
			if d.ExternalID == fmt.Sprintf("demo-report-v3-%d", i) {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		var buffer bytes.Buffer
		buffer.Write([]byte{0xef, 0xbb, 0xbf})
		writer := csv.NewWriter(&buffer)
		rows := [][]string{{"Демонстрационный отчёт", spec.title}, {"Источник", spec.source}, {"Период", "Сентябрь 2026"}}
		if spec.source == "1c" {
			if i == 2 {
				rows = append(rows, []string{"Счёт", "Наименование", "Начальное сальдо", "Оборот дебет", "Оборот кредит", "Конечное сальдо"})
				for j, name := range []string{"Материалы", "Расчётный счёт", "Расчёты с поставщиками", "Расчёты с покупателями", "Расчёты с персоналом", "Продажи"} {
					opening := 100000 + j*23000
					debit := 80000 + j*5000
					credit := 65000 + j*3000
					rows = append(rows, []string{[]string{"10", "51", "60", "62", "70", "90"}[j], name, fmt.Sprint(opening), fmt.Sprint(debit), fmt.Sprint(credit), fmt.Sprint(opening + debit - credit)})
				}
			} else {
				rows = append(rows, []string{"Подразделение", "Статья", "План, руб.", "Факт, руб.", "Отклонение, руб."})
				for j := 0; j < 16; j++ {
					plan := 125000 + j*13000
					fact := plan - 5000 + (j%4)*4200
					rows = append(rows, []string{[]string{"Продажи", "Операции", "Дизайн-студия", "Управление"}[j%4], []string{"Оплата поставщикам", "Фонд оплаты труда", "Лицензии", "Аренда"}[j/4], fmt.Sprint(plan), fmt.Sprint(fact), fmt.Sprint(fact - plan)})
				}
			}
		} else {
			rows = append(rows, []string{"ID", "Компания", "Ответственный", "Этап", "Бюджет, руб.", "Плановая дата"})
			for j := 0; j < 18; j++ {
				rows = append(rows, []string{fmt.Sprintf("DEMO-%d-%03d", i, j+1), []string{"Альфа", "Вектор", "Горизонт", "Нева", "Север", "Спектр"}[j%6], []string{"Денис Кузнецов", "Марина Белова", "Тимур Лебедев"}[j%3], []string{"Предложение", "Переговоры", "Согласование", "Успешно"}[j%4], fmt.Sprint(80000 + j*17000), time.Now().AddDate(0, 0, j-4).Format("2006-01-02")})
			}
		}
		if err := writer.WriteAll(rows); err != nil {
			return false, err
		}
		body := buffer.Bytes()
		owner := demoUsers[spec.owner]
		owner.Role = "employee"
		stamp := time.Now().Add(-time.Duration(i*4) * time.Hour).UTC().Format(time.RFC3339)
		version := Version{ID: id(), Name: spec.title + ".csv", MIME: "text/csv; charset=utf-8", Size: int64(len(body)), Created: stamp, Author: owner.Name}
		if err := os.WriteFile(filepath.Join(a.dir, "files", version.ID), body, 0600); err != nil {
			return false, err
		}
		d := Document{ID: id(), Title: spec.title, ChannelID: spec.channel, Owner: owner, Status: []string{"review", "approved", "changes", "draft"}[i%4], Description: "Демонстрационная выгрузка: данные вымышлены. Откройте таблицу, проверьте цифры и оставьте замечания.", Created: stamp, Updated: stamp, Due: time.Now().AddDate(0, 0, 3+i).Format("2006-01-02"), Versions: []Version{version}, Comments: []Comment{}, Revision: 1, SourceSystem: spec.source, ImportMethod: "demo", ExternalID: fmt.Sprintf("demo-report-v3-%d", i)}
		a.state.Documents = append(a.state.Documents, d)
	}
	a.state.DemoSeedVersion = 3
	return true, nil
}
