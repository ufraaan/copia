package copia

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestKVOperations(t *testing.T) {
	var db KV
	if _, _, err := db.Get([]byte("a")); !errors.Is(err, ErrNotOpen) {
		t.Fatalf("get before open: got %v, want errnotopen", err)
	}
	if err := db.Open(); err != nil {
		t.Fatal(err)
	}

	key := []byte("user:1")
	value := []byte("ada")
	if _, err := db.Set(key, value); err != nil {
		t.Fatal(err)
	}
	key[0] = 'x'
	value[0] = 'x'

	got, found, err := db.Get([]byte("user:1"))
	if err != nil || !found || !bytes.Equal(got, []byte("ada")) {
		t.Fatalf("get = %q, %t, %v", got, found, err)
	}
	got[0] = 'x'
	got, _, _ = db.Get([]byte("user:1"))
	if !bytes.Equal(got, []byte("ada")) {
		t.Fatalf("get returned a shared value: %q", got)
	}

	if exists, err := db.Exists([]byte("user:1")); err != nil || !exists {
		t.Fatalf("exists = %t, %v", exists, err)
	}
	if _, err := db.Set([]byte("user:2"), []byte("grace")); err != nil {
		t.Fatal(err)
	}
	keys, err := db.Keys([]byte("user:"))
	if err != nil || !reflect.DeepEqual(keys, [][]byte{[]byte("user:1"), []byte("user:2")}) {
		t.Fatalf("keys = %q, %v", keys, err)
	}
	if count, err := db.Len(); err != nil || count != 2 {
		t.Fatalf("len = %d, %v", count, err)
	}
	if deleted, err := db.Del([]byte("user:2")); err != nil || !deleted {
		t.Fatalf("del = %t, %v", deleted, err)
	}
	if err := db.Clear(); err != nil {
		t.Fatal(err)
	}
	if count, err := db.Len(); err != nil || count != 0 {
		t.Fatalf("len after clear = %d, %v", count, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.Get([]byte("user:1")); !errors.Is(err, ErrNotOpen) {
		t.Fatalf("get after close: got %v, want errnotopen", err)
	}
}

func TestRenameAndCopy(t *testing.T) {
	var db KV
	if err := db.Open(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Set([]byte("source"), []byte("value")); err != nil {
		t.Fatal(err)
	}
	if copied, err := db.Copy([]byte("source"), []byte("copy")); err != nil || !copied {
		t.Fatalf("copy = %t, %v", copied, err)
	}
	if renamed, err := db.Rename([]byte("source"), []byte("renamed")); err != nil || !renamed {
		t.Fatalf("rename = %t, %v", renamed, err)
	}
	if exists, _ := db.Exists([]byte("source")); exists {
		t.Fatal("source still exists after rename")
	}
	for _, key := range []string{"copy", "renamed"} {
		value, found, err := db.Get([]byte(key))
		if err != nil || !found || string(value) != "value" {
			t.Fatalf("get(%q) = %q, %t, %v", key, value, found, err)
		}
	}
	if copied, err := db.Copy([]byte("missing"), []byte("target")); err != nil || copied {
		t.Fatalf("copy missing key = %t, %v", copied, err)
	}
}
