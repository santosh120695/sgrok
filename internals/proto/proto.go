package proto

import (
	"encoding/binary"
	"io"
	"net"
)

const FrameSize = 8096

func WriteFrame(conn net.Conn, data []byte) error {
	dataLength := make([]byte, 4)
	binary.BigEndian.PutUint32(dataLength, uint32(len(data)))
	_, err := conn.Write(dataLength)
	if err != nil {
		return err
	}

	for len(data) > 0 {
		n, err := conn.Write(data)
		if err != nil {
			conn.Close()
			return err
		}
		data = data[n:]
	}
	return nil
}

func ReadFrame(conn net.Conn) ([]byte, error) {
	buffer := make([]byte, 4)

	_, err := io.ReadFull(conn, buffer)
	if err != nil {
		return nil, err
	}
	dataLength := binary.BigEndian.Uint32(buffer)
	buffer = make([]byte, dataLength)
	_, err = io.ReadFull(conn, buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
