package main

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var sessions sync.Map

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// 监听端口，接收http请求
	startHttp()
	// 监听端口，连接客户端使用
	startServer()
}

func startHttp() {
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
			guid := guid()
			log.Info().Uint64("guid", guid).Msg("new request")
			req := &Request{
				Conn: conn,
				Guid: guid,
			}
			sessions.Store(guid, conn)
			req.send()
		}
	}()
}

func guid() uint64 {
	return uint64(time.Now().UnixNano())
}

func startServer() {
	server, err := net.Listen("tcp", "127.0.0.1:8777")
	if err != nil {
		panic(err)
	}
	log.Info().Msg("start server success")

	lock := sync.Mutex{}
	for {
		lock.Lock()
		conn, err := server.Accept()
		if err != nil {
			break
		}
		log.Info().Msg("get new client")
		// 正向数据绑定
		defaultForward.Bind(conn)
		// 反向数据读取
		readBack(conn)
		lock.Unlock()
		log.Info().Msg("client close")
	}

}

func readBack(conn net.Conn) {
	defer conn.Close()
	for {
		req := ReadRequest(conn)
		if req == nil {
			log.Info().Msg("get nil request")
			// time.Sleep(100 * time.Millisecond)
			return
		}
		defaultBack.Back(req)
	}
}
