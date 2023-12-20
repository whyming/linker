package main

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/rs/zerolog/log"
)

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
		return nil
	}
	if n != 4 {
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
			return nil
		}
		readLen += uint32(n)
	}

	var guid uint64
	binary.Read(bytes.NewBuffer(buff[:len]), binary.BigEndian, &guid)
	log.Info().Bytes("forward-data", buff[8:len]).Msg("data")
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
