package network

import (
	"context"
	"fmt"
	"net"
	"sync"

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
	lock         sync.RWMutex
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
	c.lock.Lock()
	defer c.lock.Unlock()
	c.closeSig <- true
	c.isClosed = true
	close(c.tcpConnAPool)
	close(c.listenBuffer)
	for conn := range c.tcpConnAPool {
		if err := conn.Close(); err != nil {
			return err
		}
	}
	for _, conn := range c.tcpConnBList {
		if err := conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) Send(req []byte) (resp []byte, err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.isClosed {
		return nil, fmt.Errorf("network client is closed")
	}
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
		c.lock.RLock()
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
		c.lock.RUnlock()
	}
}

func (c *client) Listen() <-chan []byte {
	return c.listenBuffer
}
