package db0101

import "errors"

type KV struct {
	mem map[string][]byte
}

func (kv *KV) Open() error {
	kv.mem = map[string][]byte{}
	return nil
}

func (kv *KV) Close() error {
	return nil
}

func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {

	if kv.mem == nil {
		return nil, false, errors.New("kv not opened")
	}

	k := string(key)
	val, ok = kv.mem[k]
	return val, ok, nil
}

func (kv *KV) Set(key []byte, val []byte) (updated bool, err error) {

	if kv.mem == nil {
		return false, errors.New("kv not opened")
	}

	k := string(key)
	kv.mem[k] = val
	return true, nil

}

func (kv *KV) Del(key []byte) (deleted bool, err error) {

	if kv.mem == nil {
		return false, errors.New("kv not opened")
	}

	k := string(key)
	_, exists := kv.mem[k]
	delete(kv.mem, k)
	return exists, nil
}
