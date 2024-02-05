package bullet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/rs/zerolog/log"
)

type Bullet struct {
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
	ErrConnClosed    = errors.New("connection closed")
)

func NewBullet(guid uint64, cmd uint8, data []byte) *Bullet {
	return &Bullet{
		guid: guid,
		cmd:  cmd,
		data: data,
	}
}
func ReadFrom(rd io.Reader) (*Bullet, error) {
	b := &Bullet{}
	buff := make([]byte, 1024)
	n, err := rd.Read(buff[:OffsetLength])
	if err != nil {
		return nil, err
	}
	for n < OffsetLength {
		n2, err := rd.Read(buff[n:OffsetLength])
		if err != nil {
			return nil, err
		}
		n += n2
	}
	if n < OffsetLength {
		log.Error().Int("read from buff length n", n).Msg("read length bad")
		return nil, ErrInvalidLength
	}
	length := binary.BigEndian.Uint32(buff[:OffsetLength])
	if length > 1024 || length+1 < OffsetGuid+OffsetCmd {
		log.Error().Int("buff head length bad length n", n).Bytes("buff", buff[:OffsetLength]).Msg("read length not valid")
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

func (b *Bullet) GetGuid() uint64 {
	return b.guid
}
func (b *Bullet) GetCmd() uint8 {
	return b.cmd
}

func (b *Bullet) GetData() []byte {
	return b.data
}

func (b *Bullet) Bytes() []byte {
	buff := bytes.NewBuffer([]byte{})
	binary.Write(buff, binary.BigEndian, uint32(8+len(b.data))) // 4 bytes for length
	binary.Write(buff, binary.BigEndian, b.guid)                // 8 bytes for guid
	buff.Write([]byte{b.cmd})                                   // 1 byte for cmd
	buff.Write(b.data)
	return buff.Bytes()
}

type Writer interface {
	Write(*Bullet) error
}
type Reader interface {
	Read() (*Bullet, error)
}

func Copy(direction string, dst Writer, src Reader) error {
	for {
		b, err := src.Read()
		log.Debug().Str("direction", direction).Int("data-len", len(b.data)).Err(err).Msg("READ")
		if err != nil {
			return err
		}
		err = dst.Write(b)
		log.Debug().Str("direction", direction).Int("data-len", len(b.data)).Err(err).Msg("Write")
		if err != nil {
			log.Error().Err(err).Msg("write session into to tunnel fail")
		}
	}
}
