package copia

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

const (
	walHeaderSize = 12
	maxRecordSize = 64 << 20
)

var walMagic = [4]byte{'C', 'P', 'W', 'L'}

const (
	walSet byte = iota + 1
	walDel
	walRename
	walCopy
	walClear
	walSetBatch
)

var ErrRecordTooLarge = errors.New("copia: entry exceeds maximum size")

func openWAL(path string) (*os.File, map[string][]byte, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, nil, err
	}
	mem, goodEnd, err := replayWAL(file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	if err := file.Truncate(goodEnd); err != nil {
		file.Close()
		return nil, nil, err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return nil, nil, err
	}
	if _, err := file.Seek(goodEnd, io.SeekStart); err != nil {
		file.Close()
		return nil, nil, err
	}
	return file, mem, nil
}

func replayWAL(file *os.File) (map[string][]byte, int64, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}
	mem := make(map[string][]byte)
	var offset int64
	for offset < stat.Size() {
		remaining := stat.Size() - offset
		if remaining < walHeaderSize {
			break
		}
		var header [walHeaderSize]byte
		if _, err := file.ReadAt(header[:], offset); err != nil {
			return nil, offset, err
		}
		if !bytes.Equal(header[:4], walMagic[:]) {
			return nil, offset, fmt.Errorf("copia: invalid wal header at byte %d", offset)
		}
		payloadLen := binary.LittleEndian.Uint32(header[4:8])
		if payloadLen == 0 || uint64(payloadLen) > maxRecordSize+9 {
			return nil, offset, fmt.Errorf("copia: invalid wal record size at byte %d", offset)
		}
		frameLen := int64(walHeaderSize) + int64(payloadLen)
		if remaining < frameLen {
			break
		}
		payload := make([]byte, int(payloadLen))
		if _, err := file.ReadAt(payload, offset+walHeaderSize); err != nil {
			return nil, offset, err
		}
		wantCRC := binary.LittleEndian.Uint32(header[8:12])
		if crc32.ChecksumIEEE(payload) != wantCRC {
			return nil, offset, fmt.Errorf("copia: wal checksum mismatch at byte %d", offset)
		}
		if err := applyWAL(mem, payload); err != nil {
			return nil, offset, fmt.Errorf("copia: invalid wal record at byte %d: %w", offset, err)
		}
		offset += frameLen
	}
	return mem, offset, nil
}

//todo- log compaction when replaying the full history starts to drag (need to refer https://riak.com/assets/bitcask-intro.pdf)

func applyWAL(mem map[string][]byte, payload []byte) error {
	op := payload[0]
	if op == walClear {
		if len(payload) != 1 {
			return errors.New("clear record has data")
		}
		clear(mem)
		return nil
	}
	if op == walSetBatch {
		return applySetBatch(mem, payload[1:])
	}
	var entry Entry
	r := bytes.NewReader(payload[1:])
	if err := entry.Decode(r); err != nil {
		return err
	}
	if r.Len() != 0 {
		return errors.New("record has trailing data")
	}
	switch op {
	case walSet:
		mem[string(entry.Key)] = clone(entry.Value)
	case walDel:
		delete(mem, string(entry.Key))
	case walRename:
		value, ok := mem[string(entry.Key)]
		if !ok {
			return errors.New("rename source is missing")
		}
		mem[string(entry.Value)] = value
		if !bytes.Equal(entry.Key, entry.Value) {
			delete(mem, string(entry.Key))
		}
	case walCopy:
		value, ok := mem[string(entry.Key)]
		if !ok {
			return errors.New("copy source is missing")
		}
		mem[string(entry.Value)] = clone(value)
	default:
		return fmt.Errorf("unknown operation %d", op)
	}
	return nil
}

func (kv *KV) appendWAL(op byte, key, value []byte) error {
	payload := []byte{op}
	if op != walClear {
		if uint64(len(key))+uint64(len(value)) > maxRecordSize {
			return ErrRecordTooLarge
		}
		payload = append(payload, (Entry{Key: key, Value: value}).Encode()...)
	}
	return kv.appendWALPayload(payload)
}

func (kv *KV) appendWALPayload(payload []byte) error {
	if kv.wal == nil {
		return nil
	}
	if kv.walErr != nil {
		return kv.walErr
	}
	var header [walHeaderSize]byte
	copy(header[:4], walMagic[:])
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(payload)))
	binary.LittleEndian.PutUint32(header[8:12], crc32.ChecksumIEEE(payload))
	start, err := kv.wal.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if err := writeAll(kv.wal, header[:]); err != nil {
		return kv.rollbackWAL(start, err)
	}
	if err := writeAll(kv.wal, payload); err != nil {
		return kv.rollbackWAL(start, err)
	}
	if err := kv.wal.Sync(); err != nil {
		kv.walErr = err
		return err
	}
	return nil
}

func encodeSetBatch(entries []Entry) ([]byte, error) {
	if uint64(len(entries)) > uint64(^uint32(0)) {
		return nil, ErrRecordTooLarge
	}
	// the batch header takes a few bytes so room for dat
	const batchHeaderSize = 5
	var size uint64 = batchHeaderSize
	for _, entry := range entries {
		if uint64(len(entry.Key))+uint64(len(entry.Value)) > maxRecordSize {
			return nil, ErrRecordTooLarge
		}
		size += 8 + uint64(len(entry.Key)) + uint64(len(entry.Value))
		if size > maxRecordSize+9 {
			return nil, ErrRecordTooLarge
		}
	}
	payload := make([]byte, int(size))
	payload[0] = walSetBatch
	binary.LittleEndian.PutUint32(payload[1:5], uint32(len(entries)))
	offset := batchHeaderSize
	for _, entry := range entries {
		encoded := entry.Encode()
		copy(payload[offset:], encoded)
		offset += len(encoded)
	}
	return payload, nil
}

func applySetBatch(mem map[string][]byte, data []byte) error {
	if len(data) < 4 {
		return errors.New("batch record is missing entry count")
	}
	count := binary.LittleEndian.Uint32(data[:4])
	data = data[4:]
	if uint64(count)*8 > uint64(len(data)) {
		return errors.New("batch record has invalid entry count")
	}
	entries := make([]Entry, 0, int(count))
	r := bytes.NewReader(data)
	for i := uint32(0); i < count; i++ {
		var entry Entry
		if err := entry.Decode(r); err != nil {
			return err
		}
		entries = append(entries, entry)
	}
	if r.Len() != 0 {
		return errors.New("batch record has trailing data")
	}
	for _, entry := range entries {
		mem[string(entry.Key)] = clone(entry.Value)
	}
	return nil
}

func (kv *KV) rollbackWAL(start int64, cause error) error {
	if err := kv.wal.Truncate(start); err != nil {
		kv.walErr = errors.Join(cause, err)
		return kv.walErr
	}
	if _, err := kv.wal.Seek(start, io.SeekStart); err != nil {
		kv.walErr = errors.Join(cause, err)
		return kv.walErr
	}
	if err := kv.wal.Sync(); err != nil {
		kv.walErr = errors.Join(cause, err)
		return kv.walErr
	}
	return cause
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}
