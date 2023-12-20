package main

import (
	"net"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

var defaultBack = &Back{
	Buffer: make(chan *RequestData, 1),
}

func init() {
	defaultBack.Start()
}

type Back struct {
	Buffer chan *RequestData
}

func (b *Back) Back(data *RequestData) {
	b.Buffer <- data
}
func (b *Back) Start() {
	log.Info().Msg("back start")
	go func() {
		if b.Buffer == nil {
			return
		}
		for data := range b.Buffer {
			fh, _ := os.OpenFile("last.data", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
			fh.Write(data.Bytes())
			fh.Close()

			log.Debug().Uint64("guid", data.Guid).Int("data-length", len(data.Data)).Msg("get data")
			if conn, exists := sessions.Load(data.Guid); exists {
				log.Debug().Uint64("guid", data.Guid).Msg("found session")
				if c, ok := conn.(net.Conn); ok {
					if n, err := c.Write(data.Data); err != nil {
						log.Error().Uint64("guid", data.Guid).Err(err).Msg("write data back to http server fail")
					} else {
						log.Debug().Uint64("guid", data.Guid).Int("write-back-length", n).Msg("send success")
					}
					time.Sleep(time.Millisecond)
				} else {
					log.Error().Msg("conn is not net.Conn")
				}
			} else {
				log.Debug().Uint64("guid", data.Guid).Msg("not found session")
			}
		}
	}()
}
