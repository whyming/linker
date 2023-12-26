package bullet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type Buttle struct {
	guid uint64
	cmd  uint8
	data []byte
}

const (
	CmdData  uint8 = 1
	CmdClose uint8 = 2
)
const (
	OffsetLength = 4
	OffsetGuid   = 8
	OffsetCmd    = 1
)

var (
	ErrInvalidLength = errors.New("invalid length")
	ErrInvalidGuid   = errors.New("invalid guid")
)

func NewBullet(guid uint64, cmd uint8, data []byte) *Buttle {
	return &Buttle{
		guid: guid,
		cmd:  cmd,
		data: data,
	}
}
func ReadFrom(rd io.Reader) (*Buttle, error) {
	b := &Buttle{}
	buff := make([]byte, 1024)
	n, err := rd.Read(buff[:OffsetLength])
	if err != nil {
		return nil, err
	}
	if n < OffsetLength {
		return nil, ErrInvalidLength
	}
	length := binary.BigEndian.Uint32(buff[:OffsetLength])
	if length > 1024 {
		return nil, ErrInvalidLength
	}
	n, err = rd.Read(buff[:length+1])
	if err != nil {
		return nil, err
	}
	for n < int(length) {
		n2, err := rd.Read(buff[n : length+1])
		if err != nil {
			return nil, err
		}
		n += n2
	}
	b.guid = binary.BigEndian.Uint64(buff[:OffsetGuid])
	b.cmd = buff[OffsetGuid]
	b.data = buff[OffsetGuid+OffsetCmd : length+1]
	return b, nil
}

func (b *Buttle) GetGuid() uint64 {
	return b.guid
}
func (b *Buttle) GetCmd() uint8 {
	return b.cmd
}

func (b *Buttle) GetData() []byte {
	return b.data
}

func (b *Buttle) Bytes() []byte {
	buff := bytes.NewBuffer([]byte{})
	binary.Write(buff, binary.BigEndian, uint32(8+len(b.data))) // 4 bytes for length
	binary.Write(buff, binary.BigEndian, b.guid)                // 8 bytes for guid
	buff.Write([]byte{b.cmd})                                   // 1 byte for cmd
	buff.Write(b.data)
	return buff.Bytes()
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
