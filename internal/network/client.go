package network

import (
	"context"
	"net"

	"github.com/mr-panta/go-logger"
)

func NewClient(tcpAddrA, tcpAddrB string) Client {
	return &client{
		id:           randString(clientIDLength),
		tcpAddrA:     tcpAddrA,
		tcpAddrB:     tcpAddrB,
		tcpConnAPool: make(chan *Connection, poolSize),
		listenBuffer: make(chan []byte, listenBufferSize),
		closeSig:     make(chan bool, 1),
	}
}

type client struct {
	id           string
	tcpAddrA     string
	tcpAddrB     string
	tcpConnAPool chan *Connection
	tcpConnBList []*Connection
	listenBuffer chan []byte
	closeSig     chan bool
	isClosed     bool
}

func (c *client) Start() (err error) {
	for i := 0; i < poolSize; i++ {
		tcpConnA, err := net.Dial("tcp", c.tcpAddrA)
		if err != nil {
			return err
		}
		c.tcpConnAPool <- NewConnection(tcpConnA)
	}
	for i := 0; i < poolSize; i++ {
		tcpConnB, err := net.Dial("tcp", c.tcpAddrB)
		if err != nil {
			return err
		}
		c.tcpConnBList = append(c.tcpConnBList, NewConnection(tcpConnB))
		go c.listenConnB(i)
	}
	return nil
}

func (c *client) Wait() {
	<-c.closeSig
}

func (c *client) Close() error {
	c.closeSig <- true
	c.isClosed = true
	// TODO
	return nil
}

func (c *client) Send(req []byte) (resp []byte, err error) {
	tcpConn := <-c.tcpConnAPool
	// Compress data
	if req, err = compressData(req); err != nil {
		return nil, err
	}
	// Write data
	if err = tcpConn.Write(req); err != nil {
		return nil, err
	}
	// Read data
	if resp, err = tcpConn.Read(); err != nil {
		return nil, err
	}
	// Uncompress data
	if resp, err = uncompressData(resp); err != nil {
		return nil, err
	}
	// Return connect
	c.tcpConnAPool <- tcpConn
	return resp, nil
}

func (c *client) listenConnB(i int) {
	ctx := context.Background()
	buffer := []byte(c.id)
	conn := c.tcpConnBList[i]
	if err := conn.Write(buffer); err != nil {
		logger.Errorf(ctx, err.Error())
		return
	}
	for !c.isClosed {
		data, err := conn.Read()
		if err != nil {
			logger.Errorf(ctx, "%v|%v", err.Error(), conn.LocalAddr())
		}
		// Uncompress data
		data, err = uncompressData(data)
		if err != nil {
			logger.Errorf(ctx, err.Error())
			break
		}
		c.listenBuffer <- data
	}
}

func (c *client) Listen() <-chan []byte {
	return c.listenBuffer
}
