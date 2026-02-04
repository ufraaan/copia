package db0102

import (
	"encoding/binary"
	"io"
)

type Entry struct {
	key []byte
	val []byte
}

func (ent *Entry) Encode() []byte {
	data := make([]byte, 4+4+len(ent.key)+len(ent.val))
	binary.LittleEndian.PutUint32(data[0:4], uint32(len(ent.key)))
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(ent.val)))
	// copy(dst, src)
	copy(data[8:], ent.key)
	copy(data[8+len(ent.key):], ent.val)
	return data
}

func (ent *Entry) Decode(r io.Reader) error {
	// allocate header buffer. here we use 'header' to only store the metadata.
	// metadata meaning the sizes of key and val.
	var header [8]byte

	// reads from r (the argument passed in function call), and stores in header.
	// _, err := io.ReadFull(r, header[:])
	// if err != nil {
	// 	return err
	// }

	//shorthand of above
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}

	// first 4, length of key
	klen := int(binary.LittleEndian.Uint32(header[0:4]))
	// next 4, length of val
	vlen := int(binary.LittleEndian.Uint32(header[4:8]))

	// making 'data' variable to store byte slice, of length -> klen+vlen
	data := make([]byte, klen+vlen)

	// read from r, and now store in data.
	// _, err = io.ReadFull(r, data)
	// if err != nil {
	// 	return err
	// }

	//shorthand of above
	if _, err := io.ReadFull(r, data); err != nil {
		return err
	}

	// key value is in the data byte slice, from byte 0, till length of key (excluding klen)
	ent.key = data[:klen]
	// val value is in the data byte slice, from the klen, up until the end.
	ent.val = data[klen:]

	return nil

}
