package main

import (
	"linker/modules/bullet"
	"linker/modules/session"
	"linker/modules/tunnel"
	"net"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// 启动http服务
	ss := session.NewSessions()
	startHttp(ss)

	// 启动tunnel
	tun := startTunnel()

	go bullet.Copy(ss, tun)
	bullet.Copy(tun, ss)
}

func startHttp(ss *session.Sessions) {
	httpServer, err := net.Listen("tcp", ":8765")
	if err != nil {
		panic(err)
	}
	log.Info().Msg("start http server success")
	go func() {
		for {
			conn, err := httpServer.Accept()
			if err != nil {
				break
			}
			// 每收到一个请求，只是加入session就行了
			guid := guid()
			log.Info().Uint64("guid", guid).Msg("new request")
			ss.AddConn(guid, conn)
		}
	}()
}

func guid() uint64 {
	return uint64(time.Now().UnixNano())
}

func startTunnel() *tunnel.Tunnel {
	server, err := net.Listen("tcp", "127.0.0.1:8777")
	if err != nil {
		panic(err)
	}
	log.Info().Msg("start server success")
	tun := tunnel.NewTunnel()

	lock := sync.Mutex{}
	go func() {
		for {
			lock.Lock()
			conn, err := server.Accept()
			if err != nil {
				break
			}
			log.Info().Msg("get new client")
			tun.Bind(conn, func() {
				log.Info().Msg("client tunnel closed")
			})
			lock.Unlock()
		}
	}()
	return tun
}
