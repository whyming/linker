package main

import (
	"net"

	"github.com/rs/zerolog/log"
)

var defaultForward = &Forward{
	Buffer: make(chan *RequestData, 32),
}

func init() {
	defaultForward.Start()
}

type Forward struct {
	Buffer chan *RequestData
	target net.Conn
}

func (f *Forward) Bind(conn net.Conn) {
	f.target = conn
}

func (f *Forward) Forward(data *RequestData) {
	f.Buffer <- data
}

func (f *Forward) Start() {
	log.Info().Msg("forward start")
	go func() {
		if f.Buffer == nil {
			return
		}
		for data := range f.Buffer {
			if f.target != nil {
				_, err := f.target.Write(data.Bytes())
				if err != nil {
					log.Error().Str("data-flow", "forward").Err(err).Msg("write data to conn fail")
				}
			} else {
				log.Debug().Msg("forward target not found")
			}
		}
	}()
}
func (f *Forward) Stop() {
	close(f.Buffer)
}
