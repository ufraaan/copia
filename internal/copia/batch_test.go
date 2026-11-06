//making sure batches stick together, even if WAL gets cut off
package copia

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSetManyInMemory(t *testing.T) {
	var db KV
	if err := db.SetMany(nil); !errors.Is(err, ErrNotOpen) {
		t.Fatalf("SetMany before open = %v, want %v", err, ErrNotOpen)
	}
	if err := db.Open(); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	key := []byte("shared")
	first := []byte("first")
	last := []byte("last")
	entries := []Entry{
		{Key: key, Value: first},
		{Key: []byte("other"), Value: []byte("value")},
		{Key: key, Value: last},
	}
	if err := db.SetMany(entries); err != nil {
		t.Fatal(err)
	}
	key[0] = 'x'
	last[0] = 'x'

	got, found, err := db.Get([]byte("shared"))
	if err != nil || !found || !bytes.Equal(got, []byte("last")) {
		t.Fatalf("Get(shared) = %q, %t, %v; want last, true, nil", got, found, err)
	}
	got, found, err = db.Get([]byte("other"))
	if err != nil || !found || !bytes.Equal(got, []byte("value")) {
		t.Fatalf("Get(other) = %q, %t, %v; want value, true, nil", got, found, err)
	}
}

func TestSetManyReplays(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.wal")
	var db KV
	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Set([]byte("before"), []byte("kept")); err != nil {
		t.Fatal(err)
	}
	if err := db.SetMany([]Entry{
		{Key: []byte("one"), Value: []byte("1")},
		{Key: []byte("two"), Value: []byte("2")},
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for key, want := range map[string]string{"before": "kept", "one": "1", "two": "2"} {
		got, found, err := db.Get([]byte(key))
		if err != nil || !found || string(got) != want {
			t.Fatalf("Get(%q) = %q, %t, %v; want %q, true, nil", key, got, found, err, want)
		}
	}
}

func TestSetManyTruncatedWALRecordIsNotApplied(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.wal")
	var db KV
	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Set([]byte("before"), []byte("kept")); err != nil {
		t.Fatal(err)
	}
	if err := db.SetMany([]Entry{
		{Key: []byte("one"), Value: []byte("1")},
		{Key: []byte("two"), Value: []byte("2")},
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data[:len(data)-1], 0o600); err != nil {
		t.Fatal(err)
	}
	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if got, found, err := db.Get([]byte("before")); err != nil || !found || string(got) != "kept" {
		t.Fatalf("Get(before) = %q, %t, %v; want kept, true, nil", got, found, err)
	}
	for _, key := range []string{"one", "two"} {
		if got, found, err := db.Get([]byte(key)); err != nil || found {
			t.Fatalf("Get(%q) = %q, %t, %v; want missing key", key, got, found, err)
		}
	}
}
