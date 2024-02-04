package tunnel

import (
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
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msgf("readTunnel panic: %v", err)
			}
		}()
		for {
			b, err := bullet.ReadFrom(conn)
			if err != nil {
				log.Error().Err(err).Msg("read from tunnel fail")
				return
			}
			t.buff <- b
		}
	}()
}
