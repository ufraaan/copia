package copia

import (
	"bytes"
	"errors"
	"os"
	"sort"
	"strings"
	"sync"
)

var ErrNotOpen = errors.New("copia: store is not open")

type KV struct {
	mu     sync.RWMutex
	mem    map[string][]byte
	wal    *os.File
	walErr error
}

func (kv *KV) Open() error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if err := kv.closeWAL(); err != nil {
		return err
	}
	kv.mem = map[string][]byte{}
	kv.walErr = nil
	return nil
}

func (kv *KV) OpenFile(path string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if err := kv.closeWAL(); err != nil {
		return err
	}
	file, mem, err := openWAL(path)
	if err != nil {
		kv.mem = nil
		return err
	}
	kv.wal = file
	kv.mem = mem
	kv.walErr = nil
	return nil
}

func (kv *KV) Close() error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.mem = nil
	return kv.closeWAL()
}

func (kv *KV) closeWAL() error {
	if kv.wal == nil {
		return nil
	}
	syncErr := kv.wal.Sync()
	closeErr := kv.wal.Close()
	kv.wal = nil
	kv.walErr = nil
	return errors.Join(syncErr, closeErr)
}

func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {
	// a read lock is enough because returned values are copied.
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	if kv.mem == nil {
		return nil, false, ErrNotOpen
	}
	val, ok = kv.mem[string(key)]
	return clone(val), ok, nil
}

func (kv *KV) Set(key, val []byte) (updated bool, err error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if kv.mem == nil {
		return false, ErrNotOpen
	}
	if err := kv.appendWAL(walSet, key, val); err != nil {
		return false, err
	}
	kv.mem[string(key)] = clone(val)
	return true, nil
}

func (kv *KV) Del(key []byte) (deleted bool, err error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if kv.mem == nil {
		return false, ErrNotOpen
	}
	_, exists := kv.mem[string(key)]
	if !exists {
		return false, nil
	}
	if err := kv.appendWAL(walDel, key, nil); err != nil {
		return false, err
	}
	delete(kv.mem, string(key))
	return exists, nil
}

func (kv *KV) Rename(oldKey, newKey []byte) (renamed bool, err error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if kv.mem == nil {
		return false, ErrNotOpen
	}
	value, ok := kv.mem[string(oldKey)]
	if !ok {
		return false, nil
	}
	if string(oldKey) == string(newKey) {
		return true, nil
	}
	if err := kv.appendWAL(walRename, oldKey, newKey); err != nil {
		return false, err
	}
	kv.mem[string(newKey)] = value
	delete(kv.mem, string(oldKey))
	return true, nil
}

func (kv *KV) Copy(sourceKey, targetKey []byte) (copied bool, err error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if kv.mem == nil {
		return false, ErrNotOpen
	}
	value, ok := kv.mem[string(sourceKey)]
	if !ok {
		return false, nil
	}
	if string(sourceKey) == string(targetKey) {
		return true, nil
	}
	if err := kv.appendWAL(walCopy, sourceKey, targetKey); err != nil {
		return false, err
	}
	kv.mem[string(targetKey)] = clone(value)
	return true, nil
}

func (kv *KV) Exists(key []byte) (bool, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	if kv.mem == nil {
		return false, ErrNotOpen
	}
	_, ok := kv.mem[string(key)]
	return ok, nil
}

func (kv *KV) Keys(prefix []byte) ([][]byte, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	if kv.mem == nil {
		return nil, ErrNotOpen
	}
	keys := make([][]byte, 0)
	for key := range kv.mem {
		if strings.HasPrefix(key, string(prefix)) {
			keys = append(keys, []byte(key))
		}
	}
	// map order is random, so sort before returning a stable list.
	sort.Slice(keys, func(i, j int) bool { return bytes.Compare(keys[i], keys[j]) < 0 })
	return keys, nil
}

func (kv *KV) Len() (int, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	if kv.mem == nil {
		return 0, ErrNotOpen
	}
	return len(kv.mem), nil
}

func (kv *KV) Clear() error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if kv.mem == nil {
		return ErrNotOpen
	}
	if len(kv.mem) == 0 {
		return nil
	}
	if err := kv.appendWAL(walClear, nil, nil); err != nil {
		return err
	}
	kv.mem = map[string][]byte{}
	return nil
}

func clone(b []byte) []byte {
	if b == nil {
		return nil
	}
	return append([]byte{}, b...)
}
