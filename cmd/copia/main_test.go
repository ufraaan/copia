package main

import (
	"testing"

	"github.com/ufraaan/copia/internal/copia"
)

func TestRunStoreCommands(t *testing.T) {
	var db copia.KV
	if err := db.Open(); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{
		"set old value",
		"copy old duplicate",
		"rename old new",
	} {
		if err := run(&db, command); err != nil {
			t.Fatalf("run(%q): %v", command, err)
		}
	}
	for _, key := range []string{"new", "duplicate"} {
		value, found, err := db.Get([]byte(key))
		if err != nil || !found || string(value) != "value" {
			t.Fatalf("get(%q) = %q, %t, %v", key, value, found, err)
		}
	}
	if err := run(&db, "copy only-one"); err == nil {
		t.Fatal("copy without a target should fail")
	}
}
