package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"

	"github.com/rs/zerolog/log"
)

type Request struct {
	Conn net.Conn
	Guid uint64
}

func (r *Request) send() {
	go func() {
		buff := make([]byte, 1024)
		defer r.Conn.Close()
		defer sessions.Delete(r.Guid)
		for {
			n, err := r.Conn.Read(buff)
			if err != nil {
				log.Info().Uint64("guid", r.Guid).Msg("request end")
				return
			}
			if n > 0 {
				data := NewRequestData(r.Guid, buff[0:n])
				log.Debug().Bytes("forward-data", buff[0:n]).Send()
				defaultForward.Forward(data)
			}
		}
	}()
}

type RequestData struct {
	Guid uint64
	Data []byte
}

func (r *RequestData) Bytes() []byte {
	buff := bytes.NewBuffer([]byte{})
	binary.Write(buff, binary.BigEndian, uint32(8+len(r.Data)))
	binary.Write(buff, binary.BigEndian, r.Guid)
	buff.Write(r.Data)
	return buff.Bytes()
}

func ReadRequest(reader io.Reader) *RequestData {
	buff := make([]byte, 1024)
	// 读取包长
	n, err := reader.Read(buff[:4])
	if err != nil {
		log.Error().Err(err).Msg("read data length fail")
		return nil
	}
	if n != 4 {
		log.Error().Msg("data length is not 4")
		return nil
	}
	var len uint32
	binary.Read(bytes.NewBuffer(buff[:4]), binary.BigEndian, &len)
	// 长度超过1024，错误数据，重连
	if len > 1024 || len < 8 {
		return nil
	}
	var readLen uint32
	for readLen < len {
		n, err := reader.Read(buff[readLen : len-readLen])
		if err != nil {
			log.Error().Err(err).Msg("read back data fail")
			return nil
		}
		readLen += uint32(n)
	}

	var guid uint64
	binary.Read(bytes.NewBuffer(buff[:8]), binary.BigEndian, &guid)
	return &RequestData{
		Guid: guid,
		Data: buff[8:len],
	}
}

func NewRequestData(guid uint64, data []byte) *RequestData {
	return &RequestData{
		Guid: guid,
		Data: data,
	}
}
