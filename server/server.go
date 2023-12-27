package main

import (
	"flag"
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

var (
	sessionAddr = flag.String("s", ":8777", "session server address")
	tunnelAddr  = flag.String("t", ":8765", "tunnel server address")
	debug       = flag.Bool("d", false, "debug mode")
)

func init() {
	flag.Parse()
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// 启动http服务
	ss := session.NewSessions(*debug)
	startSesionServer(ss)

	// 启动tunnel
	tun := startTunnel()

	go bullet.Copy(ss, tun)
	bullet.Copy(tun, ss)
}

func startSesionServer(ss *session.Sessions) {
	httpServer, err := net.Listen("tcp", *sessionAddr)
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
	server, err := net.Listen("tcp", *tunnelAddr)
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
