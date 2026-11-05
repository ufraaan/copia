package copia

import (
	"bytes"
	"errors"
	"sort"
	"strings"
	"sync"
)

var ErrNotOpen = errors.New("copia: store is not open")

type KV struct {
	mu  sync.RWMutex
	mem map[string][]byte
}

func (kv *KV) Open() error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.mem = map[string][]byte{}
	return nil
}

func (kv *KV) Close() error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.mem = nil
	return nil
}

func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {
	// only need a read lock here,stored bytes never escape directly
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
	kv.mem[string(newKey)] = value
	if string(oldKey) != string(newKey) {
		delete(kv.mem, string(oldKey))
	}
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
	// map order,so sort before returning a stable list
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
	kv.mem = map[string][]byte{}
	return nil
}

func clone(b []byte) []byte {
	if b == nil {
		return nil
	}
	return append([]byte{}, b...)
}
