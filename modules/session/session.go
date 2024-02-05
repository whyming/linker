package session

import (
	"errors"
	"io"
	"linker/modules/bullet"
	"sync"

	"github.com/rs/zerolog/log"
)

/**
 * sesssion 一对多场景， 绑定一堆connection，
 * tunnel中读取出bullet，然后找到对应connection，发送原始数据
 * 写数据就是从connection中读取原始数据，然后将guid和data给tunnel
 * 如果连接不存在则，创建连接或是丢弃数据
 */

type Sessions struct {
	sync.RWMutex
	conns sync.Map // map[uint64]net.Conn
	buff  chan *bullet.Bullet
	new   func() io.ReadWriteCloser
}

func NewSessions() *Sessions {
	return &Sessions{
		conns: sync.Map{},
		buff:  make(chan *bullet.Bullet, 128),
	}
}

// 找不到时，被动触发
func (s *Sessions) SetNew(f func() io.ReadWriteCloser) {
	s.new = f
}

// 主动加入
func (s *Sessions) AddConn(guid uint64, conn io.ReadWriteCloser) {
	s.conns.Store(guid, conn)
	s.readSession(guid, conn)
}

// 连接断开后删除
func (s *Sessions) RemoveConn(guid uint64) {
	s.conns.Delete(guid)
}

func (s *Sessions) readSession(guid uint64, conn io.ReadWriteCloser) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msgf("readSession panic: %v", err)
			}
		}()
		for {
			buf := make([]byte, 1000)
			n, err := conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					s.buff <- bullet.NewBullet(guid, bullet.CmdClose, []byte{})
					log.Info().Uint64("guid", guid).Msg("session closed")
				} else {
					log.Error().Err(err).Msg("read session error")
				}
				s.RemoveConn(guid)
				return
			} else {
				log.Debug().Uint64("guid", guid).Msgf("read session %d bytes", n)
				s.buff <- bullet.NewBullet(guid, bullet.CmdData, buf[:n])
			}
		}
	}()
}

func (s *Sessions) ListConn() {
	s.conns.Range(func(key, _ any) bool {
		log.Info().Uint64("guid", key.(uint64)).Msg("session item")
		return true
	})
}

func (s *Sessions) CleanUp() {
	s.conns.Range(func(key, _ any) bool {
		conn, _ := s.conns.LoadAndDelete(key)
		if c, ok := conn.(io.Closer); ok {
			c.Close()
		}
		return true
	})
}

var (
	ErrSessionNotFound = errors.New("session not found")
)

func (s *Sessions) Write(b *bullet.Bullet) error {
	conn, ok := s.conns.Load(b.GetGuid())
	if !ok {
		if s.new == nil {
			return ErrSessionNotFound
		} else {
			c := s.new()
			if c == nil {
				return ErrSessionNotFound
			}
			s.AddConn(b.GetGuid(), c)
			conn = c
		}
	}

	var err error
	switch b.GetCmd() {
	case bullet.CmdClose:
		s.RemoveConn(b.GetGuid())
		conn.(io.ReadWriteCloser).Close()
	case bullet.CmdData:
		_, err = conn.(io.ReadWriteCloser).Write(b.GetData())
	default:
		err = errors.New("unknown cmd")
	}
	return err
}

func (s *Sessions) Read() (*bullet.Bullet, error) {
	b, ok := <-s.buff
	if !ok {
		return nil, errors.New("session read error")
	}
	return b, nil
}
