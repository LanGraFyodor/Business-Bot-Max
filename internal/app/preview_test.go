package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBusinessFilePreviews(t *testing.T) {
	cases := []struct{ name, mime, want string }{
		{"contract.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "Договор на разработку сайта"},
		{"budget.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "Бюджет"},
		{"sales.csv", "text/csv; charset=utf-8", "Меридиан"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := demoAssets.ReadFile("demo_assets/" + tc.name)
			if err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(t.TempDir(), tc.name)
			if err := os.WriteFile(file, b, 0600); err != nil {
				t.Fatal(err)
			}
			p, err := previewFile(file, tc.mime)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, block := range p.Blocks {
				if strings.Contains(block.Text, tc.want) {
					found = true
				}
			}
			for _, sheet := range p.Sheets {
				if strings.Contains(sheet.Name, tc.want) {
					found = true
				}
				for _, row := range sheet.Rows {
					for _, cell := range row {
						if strings.Contains(cell, tc.want) {
							found = true
						}
					}
				}
			}
			if !found {
				t.Fatalf("missing %q in preview: %+v", tc.want, p)
			}
		})
	}
}
func TestPreviewRequiresTeamAccess(t *testing.T) {
	a, err := newApp(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	uploaded := uploadTest(a, "/api/documents", "1", map[string]string{"title": "Private", "channelId": "studio-projects"})
	if uploaded.Code != 201 {
		t.Fatal(uploaded.Body.String())
	}
	var d Document
	if err := json.Unmarshal(uploaded.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	base := "/api/documents/" + d.ID + "/files/" + d.Versions[0].ID + "/preview"
	if w := request(a, "GET", base, "2", ""); w.Code != 404 {
		t.Fatalf("outsider preview: %d", w.Code)
	}
}
