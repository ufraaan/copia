package copia

import (
	"encoding/binary"
	"io"
)

type Entry struct {
	Key   []byte
	Value []byte
}

func (ent Entry) Encode() []byte {
	data := make([]byte, 8+len(ent.Key)+len(ent.Value))
	binary.LittleEndian.PutUint32(data[:4], uint32(len(ent.Key)))
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(ent.Value)))
	copy(data[8:], ent.Key)
	copy(data[8+len(ent.Key):], ent.Value)
	return data
}

func (ent *Entry) Decode(r io.Reader) error {
	var header [8]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}
	klen := uint64(binary.LittleEndian.Uint32(header[:4]))
	vlen := uint64(binary.LittleEndian.Uint32(header[4:8]))
	// lengths come from the input, so cap them before allocating
	if klen+vlen > maxRecordSize {
		return ErrRecordTooLarge
	}
	data := make([]byte, int(klen+vlen))
	if _, err := io.ReadFull(r, data); err != nil {
		return err
	}
	ent.Key = clone(data[:int(klen)])
	ent.Value = clone(data[int(klen):])
	return nil
}
