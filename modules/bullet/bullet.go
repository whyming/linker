package bullet

import (
	"bytes"
	"encoding/binary"
)

type Buttle struct {
	guid uint64
	data []byte
}

func NewBullet(guid uint64, data []byte) *Buttle {
	return &Buttle{
		guid: guid,
		data: data,
	}
}

func (b *Buttle) GetGuid() uint64 {
	return b.guid
}

func (b *Buttle) GetData() []byte {
	return b.data
}

func (b *Buttle) Bytes() []byte {
	buff := bytes.NewBuffer([]byte{})
	binary.Write(buff, binary.BigEndian, uint32(8+len(b.data)))
	binary.Write(buff, binary.BigEndian, b.guid)
	buff.Write(b.data)
	return buff.Bytes()
}

func (b *Buttle) Read([]byte) (int, error) {
	return 0, nil
}

type Writer interface {
	Write(*Buttle) error
}
type Reader interface {
	Read() (*Buttle, error)
}

func Copy(dst Writer, src Reader) error {
	for {
		b, err := src.Read()
		if err != nil {
			return err
		}
		err = dst.Write(b)
		if err != nil {
			return err
		}
	}
}
