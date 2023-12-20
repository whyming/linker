package main

import (
	"net"

	"github.com/rs/zerolog/log"
)

var defaultRemote = &Remote{
	Buffer: make(chan *RequestData, 1),
}

func init() {
	defaultRemote.Start()
}

type Remote struct {
	target net.Conn
	Buffer chan *RequestData
}

func (r *Remote) Bind(conn net.Conn) {
	r.target = conn
}

func (r *Remote) Send(data *RequestData) {
	r.Buffer <- data
	log.Debug().Msg("send success")
}

func (r *Remote) Start() {
	go func() {
		log.Info().Msg("start remote channel")
		for data := range r.Buffer {
			log.Info().Msg("get new data")
			if _, err := r.target.Write(data.Bytes()); err != nil {
				log.Error().Err(err).Msg("write to remote conn fail")
			} else {
				log.Debug().Uint64("guid", data.Guid).Msg("send remote success")
			}
		}
	}()
}
