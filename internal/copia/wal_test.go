package copia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWALReplaysMutations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.wal")
	var db KV
	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Set([]byte("user:1"), []byte("ada")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Set([]byte("user:2"), []byte("grace")); err != nil {
		t.Fatal(err)
	}
	if copied, err := db.Copy([]byte("user:1"), []byte("backup")); err != nil || !copied {
		t.Fatalf("copy = %t, %v", copied, err)
	}
	if renamed, err := db.Rename([]byte("user:2"), []byte("user:two")); err != nil || !renamed {
		t.Fatalf("rename = %t, %v", renamed, err)
	}
	if deleted, err := db.Del([]byte("backup")); err != nil || !deleted {
		t.Fatalf("del = %t, %v", deleted, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"user:1": "ada", "user:two": "grace"} {
		got, found, err := db.Get([]byte(key))
		if err != nil || !found || string(got) != want {
			t.Fatalf("get(%q) = %q, %t, %v", key, got, found, err)
		}
	}
	if found, _ := db.Exists([]byte("user:2")); found {
		t.Fatal("renamed key was restored")
	}
	if count, err := db.Len(); err != nil || count != 2 {
		t.Fatalf("len = %d, %v", count, err)
	}
	if err := db.Clear(); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	if count, err := db.Len(); err != nil || count != 0 {
		t.Fatalf("len after replaying clear = %d, %v", count, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWALTruncatesIncompleteTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.wal")
	var db KV
	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Set([]byte("key"), []byte("value")); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("CPW")); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	if got, found, err := db.Get([]byte("key")); err != nil || !found || string(got) != "value" {
		t.Fatalf("get after recovery = %q, %t, %v", got, found, err)
	}
	if _, err := db.Set([]byte("next"), []byte("record")); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	if got, found, err := db.Get([]byte("next")); err != nil || !found || string(got) != "record" {
		t.Fatalf("get after appending to recovered log = %q, %t, %v", got, found, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWALRejectsChecksumCorruption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.wal")
	var db KV
	if err := db.OpenFile(path); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Set([]byte("key"), []byte("value")); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	stat, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	last := make([]byte, 1)
	if _, err := file.ReadAt(last, stat.Size()-1); err != nil {
		t.Fatal(err)
	}
	last[0] ^= 0xff
	if _, err := file.WriteAt(last, stat.Size()-1); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.OpenFile(path); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("open corrupted wal = %v, want checksum error", err)
	}
}
