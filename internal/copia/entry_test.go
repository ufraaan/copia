package copia

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"testing"
)

func TestEntryEncodeAndDecode(t *testing.T) {
	entry := Entry{Key: []byte("cat"), Value: []byte("nap")}
	encoded := entry.Encode()
	if got, want := hex.EncodeToString(encoded), "03000000030000006361746e6170"; got != want {
		t.Fatalf("encode = %s, want %s", got, want)
	}
	var decoded Entry
	if err := decoded.Decode(bytes.NewReader(encoded)); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded.Key, entry.Key) || !bytes.Equal(decoded.Value, entry.Value) {
		t.Fatalf("decode = (%q, %q)", decoded.Key, decoded.Value)
	}
}

func TestEntryDecodeRejectsTruncatedAndOversizedRecords(t *testing.T) {
	var entry Entry
	if err := entry.Decode(strings.NewReader("short")); err == nil {
		t.Fatal("decode accepted a truncated header")
	}
	var header [8]byte
	binary.LittleEndian.PutUint32(header[:4], 64<<20)
	binary.LittleEndian.PutUint32(header[4:], 1)
	if err := entry.Decode(bytes.NewReader(header[:])); err == nil {
		t.Fatal("decode accepted an oversized entry")
	}
}
