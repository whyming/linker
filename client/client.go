package main

import (
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	connServer()
}

func connServer() {
	for {
		log.Info().Msg("start to connect server")
		conn, err := net.Dial("tcp", "127.0.0.1:8777")
		if err != nil {
			log.Error().Err(err).Msg("connect to server fail")
		} else {
			// 绑定反向数据通路
			defaultRemote.Bind(conn)
			// 读取连接数据，直到断开
			readRequest(conn)
		}
		log.Info().Msg("connect to server fail,try reconnect")
		time.Sleep(10 * time.Second)
	}
}

func readRequest(conn net.Conn) {
	defer conn.Close()
	for {
		req := ReadRequest(conn)
		if req == nil {
			log.Info().Msg("get nil request")
			return
		}
		defaultLocal.Send(req)
	}
}

func NewConn(guid uint64) net.Conn {
	conn, err := net.Dial("tcp", "10.138.35.98:8001")
	if err != nil {
		log.Error().Err(err).Msg("connect to server fail")
	}
	sessions.Store(guid, conn)
	go func() {
		buff := make([]byte, 1000)
		defer conn.Close()
		defer sessions.Delete(guid)
		defer func() {
			log.Info().Msg("connect target close")
		}()
		for {
			n, err := conn.Read(buff)
			if err != nil {
				return
			}
			if n > 0 {
				log.Debug().Bytes("response", buff[0:n]).Msg("from local server")
				data := NewRequestData(guid, buff[0:n])
				defaultRemote.Send(data)
			} else {
				log.Debug().Msg("get zero bytes")
			}
			// time.Sleep(time.Millisecond)
		}
	}()
	return conn
}
