package app

import "testing"

func TestEmptyLiveWorkspaceDoesNotMigrateToOrphanTeam(t *testing.T) {
	dir := t.TempDir()
	first, err := newApp(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.state.Teams) != 0 {
		t.Fatal("new live workspace is not empty")
	}
	restored, err := newApp(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.state.Teams) != 0 {
		t.Fatal("restart created orphan team")
	}
	restored.state.Teams = []Team{{ID: "orbita", Name: "Орбита", Members: []User{}}}
	if err := restored.save(); err != nil {
		t.Fatal(err)
	}
	repaired, err := newApp(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(repaired.state.Teams) != 0 {
		t.Fatal("empty legacy orphan was not repaired")
	}
}
