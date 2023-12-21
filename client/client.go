package main

import (
	"io"
	"net"
	"os"
	"time"

	"linker/modules/bullet"
	"linker/modules/session"
	"linker/modules/tunnel"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	tun := tunnel.NewTunnel()
	connServer(tun)

	ss := session.NewSessions()
	ss.SetNew(func(guid uint64) io.ReadWriteCloser {
		conn, err := net.Dial("tcp", "10.138.35.98:8001")
		if err != nil {
			log.Error().Err(err).Msg("connect to local session fail")
		}
		return conn
	})
	go bullet.Copy(tun, ss)
	bullet.Copy(ss, tun)
}

func connServer(tun *tunnel.Tunnel) {
	reConnect := func() {
		time.Sleep(10 * time.Second)
		log.Info().Msg("connect to server fail,try reconnect")
		connServer(tun)
	}

	log.Info().Msg("start to connect server")
	conn, err := net.Dial("tcp", "127.0.0.1:8777")
	if err != nil {
		log.Error().Err(err).Msg("connect to server fail")
	} else {
		tun.Bind(conn, reConnect)
	}
}
