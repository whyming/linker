package session

import (
	"errors"
	"io"
	"linker/modules/bullet"
	"os"
	"path"
	"strconv"
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
	buff  chan *bullet.Buttle
	new   func(guid uint64) io.ReadWriteCloser
}

func NewSessions() *Sessions {
	return &Sessions{
		conns: sync.Map{},
		buff:  make(chan *bullet.Buttle, 128),
	}
}

// 找不到时，被动触发
func (s *Sessions) SetNew(f func(guid uint64) io.ReadWriteCloser) {
	s.new = f
}

// 主动加入
func (s *Sessions) AddConn(guid uint64, conn io.ReadWriteCloser) {
	s.conns.Store(guid, conn)
	s.readSession(guid, conn)
}
func (s *Sessions) RemoveConn(guid uint64) {
	s.conns.Delete(guid)
}

func (s *Sessions) readSession(guid uint64, conn io.ReadWriteCloser) {
	go func() {
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

var (
	ErrSessionNotFound = errors.New("session not found")
)

func (s *Sessions) Write(b *bullet.Buttle) error {
	conn, ok := s.conns.Load(b.GetGuid())
	if !ok {
		if s.new == nil {
			return ErrSessionNotFound
		} else {
			c := s.new(b.GetGuid())
			if c != nil {
				s.AddConn(b.GetGuid(), c)
				conn = c
			}
		}
	}
	if conn == nil {
		return ErrSessionNotFound
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

func (s *Sessions) Read() (*bullet.Buttle, error) {
	b, ok := <-s.buff
	if !ok {
		return nil, errors.New("session read error")
	}
	writeFile(strconv.FormatInt(int64(b.GetGuid()), 10), b.GetData())
	return b, nil
}

func writeFile(name string, data []byte) error {
	f, err := os.OpenFile(path.Join("data/", name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}
