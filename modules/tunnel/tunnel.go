package tunnel

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"linker/modules/bullet"
	"sync"

	"github.com/rs/zerolog/log"
)

var (
	ErrTunnelClosed = errors.New("tunnel closed")
)

/**
 * 单对单发送，通过消息通道发送给确定的连接
 */

type Tunnel struct {
	sync.RWMutex
	conn io.ReadWriteCloser
	buff chan *bullet.Buttle
}

func NewTunnel() *Tunnel {
	return &Tunnel{
		buff: make(chan *bullet.Buttle, 128),
	}
}

func (t *Tunnel) Write(b *bullet.Buttle) error {
	t.RLock()
	defer t.RUnlock()
	if t.conn == nil {
		return nil
	}
	_, err := t.conn.Write(b.Bytes())
	return err
}
func (t *Tunnel) Read() (*bullet.Buttle, error) {
	b, ok := <-t.buff
	if !ok {
		return nil, ErrTunnelClosed
	}
	return b, nil
}

func (t *Tunnel) Bind(conn io.ReadWriteCloser, onClose func()) {
	t.Lock()
	defer t.Unlock()
	if t.conn != nil {
		t.conn.Close()
	}
	t.conn = conn
	t.readTunnel(conn, onClose)
}

func (t *Tunnel) readTunnel(conn io.ReadWriteCloser, onClose func()) {
	go func() {
		defer onClose()
		for {
			buff := make([]byte, 1024)
			// 读取包长
			n, err := t.conn.Read(buff[:4])
			if err != nil {
				log.Error().Err(err).Msg("read data length fail")
				return
			}
			if n != 4 {
				log.Error().Msg("data length is not 4")
				return
			}
			var len uint32
			binary.Read(bytes.NewBuffer(buff[:4]), binary.BigEndian, &len)
			// 长度超过1024，错误数据，重连
			if len > 1024 || len < 8 {
				log.Error().Uint32("length", len).Msg("data length not valid")
				return
			}
			var readLen uint32
			for readLen < len {
				n, err := t.conn.Read(buff[readLen : len-readLen])
				if err != nil {
					log.Error().Err(err).Msg("read back data fail")
					return
				}
				readLen += uint32(n)
			}

			var guid uint64
			binary.Read(bytes.NewBuffer(buff[:8]), binary.BigEndian, &guid)
			t.buff <- bullet.NewBullet(guid, buff[8:len])
		}
	}()
}
