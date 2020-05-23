package network

import (
	"encoding/binary"
	"net"
)

// Connection

type Connection struct {
	conn net.Conn
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) Read() (data []byte, err error) {
	length := 4
	lengthBytes := make([]byte, length)
	for length > 0 {
		n, err := c.conn.Read(lengthBytes[len(lengthBytes)-length:])
		if err != nil {
			return nil, err
		}
		length -= n
	}
	length = int(binary.LittleEndian.Uint32(lengthBytes))
	data = make([]byte, length)
	for length > 0 {
		n, err := c.conn.Read(data[len(data)-length:])
		if err != nil {
			return nil, err
		}
		length -= n
	}
	return data, nil
}

func (c *Connection) Write(data []byte) error {
	length := len(data)
	extendedLength := length + 4
	dataWithLength := make([]byte, extendedLength)
	binary.LittleEndian.PutUint32(dataWithLength[:4], uint32(length))
	copy(dataWithLength[4:], data)
	for extendedLength > 0 {
		n, err := c.conn.Write(dataWithLength[len(dataWithLength)-extendedLength:])
		if err != nil {
			return err
		}
		extendedLength -= n
	}
	return nil
}
