package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOldDemoTXTExamplesUpgradeWithoutChangingUserWork(t *testing.T) {
	dir := t.TempDir()
	a, err := newApp(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	d := &a.state.Documents[0]
	originalID := d.ID
	originalStatus := d.Status
	originalComments := len(d.Comments)
	v := &d.Versions[0]
	v.Name = d.Title + ".txt"
	v.MIME = "text/plain; charset=utf-8"
	old := []byte("ДЕМОНСТРАЦИОННЫЙ ДОКУМЕНТ\noriginal placeholder")
	v.Size = int64(len(old))
	if err := os.WriteFile(filepath.Join(dir, "files", v.ID), old, 0600); err != nil {
		t.Fatal(err)
	}
	if err := a.save(); err != nil {
		t.Fatal(err)
	}
	restored, err := newApp(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	got := restored.state.Documents[0]
	if got.ID != originalID || got.Status != originalStatus || len(got.Comments) != originalComments || !strings.HasSuffix(got.Versions[0].Name, ".docx") {
		t.Fatalf("unsafe migration: %+v", got)
	}
	b, err := os.ReadFile(filepath.Join(dir, "files", got.Versions[0].ID))
	if err != nil || !strings.HasPrefix(string(b), "PK") {
		t.Fatalf("Word example missing: %v", err)
	}
}
