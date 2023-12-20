package main

import (
	"net"
	"sync"

	"github.com/rs/zerolog/log"
)

var (
	defaultLocal = &Local{}
	sessions     sync.Map
)

type Local struct {
}

func (l *Local) Send(data *RequestData) {
	conn, exists := sessions.Load(data.Guid)
	if !exists {
		conn = NewConn(data.Guid)
	}
	if c, ok := conn.(net.Conn); ok {
		log.Debug().Uint64("guid", data.Guid).Msg("write to local server")
		if _, err := c.Write(data.Data); err != nil {
			log.Error().Err(err).Msg("write to local server fail")
		}
	}
}
