package main

import (
	"flag"
	"linker/modules/bullet"
	"linker/modules/session"
	"linker/modules/tunnel"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	ss := session.NewSessions()
	startSesionServer(ss)

	// 启动tunnel
	tun := startTunnel()

	SignalProcess(ss)
	go bullet.Copy("tunnel->session", ss, tun)
	bullet.Copy("session->tunnel", tun, ss)
}

func startSesionServer(ss *session.Sessions) {
	server, err := net.Listen("tcp", *sessionAddr)
	if err != nil {
		panic(err)
	}
	log.Info().Msg("start http server success")
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msgf("startSesionServer panic: %v", err)
			}
		}()
		for {
			conn, err := server.Accept()
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
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msgf("startTunnel panic: %v", err)
			}
		}()
		for {
			lock.Lock()
			conn, err := server.Accept()
			if err != nil {
				lock.Unlock()
				break
			}
			log.Info().Msg("get new client")
			tun.Bind(conn, func() {
				log.Info().Msg("client tunnel closed")
				lock.Unlock()
			})
		}
	}()
	return tun
}

func SignalProcess(ss *session.Sessions) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		for {
			s := <-c
			if s == syscall.SIGQUIT {
				log.Info().Msg("list session guids")
				ss.ListConn()
			} else {
				*debug = !*debug
				log.Info().Bool("change to", *debug).Msg("change debug mode")
				if *debug {
					zerolog.SetGlobalLevel(zerolog.InfoLevel)
				} else {
					zerolog.SetGlobalLevel(zerolog.DebugLevel)
				}
			}
		}
	}()
}
