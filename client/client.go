package main

import (
	"flag"
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

var (
	sessionAddr = flag.String("s", "127.0.0.1:8001", "session server address")
	tunnelAddr  = flag.String("t", "127.0.0.1:8777", "tunnel server address")
	debug       = flag.Bool("d", false, "debug mode")
)

func init() {
	flag.Parse()
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	ss := session.NewSessions()
	ss.SetNew(func() io.ReadWriteCloser {
		conn, err := net.Dial("tcp", *sessionAddr)
		if err != nil {
			log.Error().Err(err).Msg("connect to local session fail")
		} else {
			log.Info().Msg("connect to local session success")
		}
		return conn
	})

	tun := tunnel.NewTunnel()
	connServer(tun, ss)
	go bullet.Copy("session->tunnel", tun, ss)
	bullet.Copy("tunnel->session", ss, tun)
}

func connServer(tun *tunnel.Tunnel, ss *session.Sessions) {
	reConnect := func() {
		ss.CleanUp()
		log.Info().Msg("connect to server fail,after 10s retry to connect")
		time.Sleep(10 * time.Second)
		connServer(tun, ss)
	}

	log.Info().Msg("start to connect server")
	conn, err := net.Dial("tcp", *tunnelAddr)
	if err != nil {
		log.Error().Err(err).Msg("connect to server fail")
		reConnect()
	} else {
		log.Info().Msg("connect to server succes")
		tun.Bind(conn, reConnect)
	}
}
