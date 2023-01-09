package network

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/go-logger"
)

func NewServer(tcpAddrA, tcpAddrB string, process Process) Server {
	return &server{
		tcpAddrA:      tcpAddrA,
		tcpAddrB:      tcpAddrB,
		process:       process,
		closeSig:      make(chan bool, 1),
		clientConnMap: make(map[string]chan *Connection),
		clientBuffMap: make(map[string]chan []byte),
	}
}

type server struct {
	tcpAddrA      string
	tcpAddrB      string
	process       Process
	closeSig      chan bool
	tcpListenerA  net.Listener
	tcpListenerB  net.Listener
	clientConnMap map[string]chan *Connection
	clientBuffMap map[string]chan []byte
	clientLock    sync.RWMutex
	isClosed      bool
}

func (s *server) Start() (err error) {
	ctx := context.Background()
	s.tcpListenerA, err = net.Listen("tcp", s.tcpAddrA)
	if err != nil {
		return err
	}
	logger.Infof(ctx, "listen tcp (a) connection|%+v", s.tcpListenerA.Addr())
	go s.listenTCPA()
	s.tcpListenerB, err = net.Listen("tcp", s.tcpAddrB)
	if err != nil {
		return err
	}
	logger.Infof(ctx, "listen tcp (b) connection|%+v", s.tcpListenerB.Addr())
	go s.listenTCPB()
	return nil
}

func (s *server) listenTCPA() {
	for !s.isClosed {
		conn, err := s.tcpListenerA.Accept()
		if err != nil {
			logger.Errorf(context.Background(), err.Error())
		} else {
			logger.Infof(context.Background(), "connection is created via tcp (a)|%+v", conn.RemoteAddr())
			go s.handleTCP(NewConnection(conn))
		}
	}
}

func (s *server) handleTCP(conn *Connection) {
	ctx := context.Background()
	for {
		var err error
		// Read data
		req, err := conn.Read()
		if err != nil {
			logger.Errorf(ctx, err.Error())
			return
		}
		// Uncompress data
		if req, err = uncompressData(req); err != nil {
			logger.Errorf(ctx, err.Error())
			return
		}
		// Process data
		resp := s.process(req)
		// Compress data
		if resp, err = compressData(resp); err != nil {
			logger.Errorf(ctx, err.Error())
			return
		}
		// Write data
		if err = conn.Write(resp); err != nil {
			logger.Errorf(ctx, err.Error())
			return
		}
	}
}

func (s *server) addClientPool(clientID string, conn *Connection) {
	s.clientLock.Lock()
	defer s.clientLock.Unlock()
	chPool := make(chan *Connection, poolSize)
	if pool, exists := s.clientConnMap[clientID]; exists {
		chPool = pool
	} else {
		s.clientConnMap[clientID] = chPool
	}
	chPool <- conn
}

func (s *server) sendToClient(clientID string, data []byte) (err error) {
	s.clientLock.RLock()
	defer s.clientLock.RUnlock()
	var pool chan *Connection
	if p, exists := s.clientConnMap[clientID]; exists {
		pool = p
	} else {
		return errors.New("client connection pool not found")
	}
	timeout := time.After(time.Second)
	var conn *Connection
	select {
	case <-timeout:
		s.clientLock.RUnlock()
		s.clientLock.Lock()
		if buffer, exists := s.clientBuffMap[clientID]; exists {
			close(buffer)
		}
		delete(s.clientBuffMap, clientID)
		delete(s.clientConnMap, clientID)
		s.clientLock.Unlock()
		s.clientLock.RLock()
		return errors.New("no connection available")
	case conn = <-pool:
	}
	// Compress data
	if data, err = compressData(data); err != nil {
		return err
	}
	// Write data
	if err = conn.Write(data); err != nil {
		return err
	}
	pool <- conn
	return nil
}

func (s *server) getClientIDs() (list []string) {
	s.clientLock.Lock()
	defer s.clientLock.Unlock()
	for clientID := range s.clientConnMap {
		list = append(list, clientID)
	}
	return list
}

func (s *server) listenTCPB() {
	ctx := context.Background()
	for !s.isClosed {
		c, err := s.tcpListenerB.Accept()
		if err != nil {
			logger.Errorf(ctx, err.Error())
			continue
		}
		conn := NewConnection(c)
		clientIDBytes, err := conn.Read()
		if err != nil {
			logger.Errorf(ctx, err.Error())
			continue
		}
		clientID := string(clientIDBytes)
		logger.Infof(ctx, "client_id:%s|connection is created via tcp (b)|%+v", clientID, conn.RemoteAddr())
		s.addClientPool(clientID, conn)
		if s.newBuffer(clientID) {
			go s.startBroadcastWorker(clientID)
		}
	}
}

func (s *server) Wait() {
	<-s.closeSig
}

func (s *server) Close() error {
	if err := s.tcpListenerA.Close(); err != nil {
		return err
	}
	if err := s.tcpListenerB.Close(); err != nil {
		return err
	}
	s.isClosed = true
	s.closeSig <- true
	return nil
}

func (s *server) getBuffer(clientID string) (buffer chan []byte, exists bool) {
	s.clientLock.RLock()
	defer s.clientLock.RUnlock()
	buffer, exists = s.clientBuffMap[clientID]
	if exists {
		return buffer, true
	}
	return nil, false
}

func (s *server) newBuffer(clientID string) (ok bool) {
	s.clientLock.Lock()
	defer s.clientLock.Unlock()
	if _, exists := s.clientBuffMap[clientID]; !exists {
		buffer := make(chan []byte, config.BufferSize)
		s.clientBuffMap[clientID] = buffer
		return true
	}
	return false
}

func (s *server) startBroadcastWorker(clientID string) {
	ctx := context.Background()
	buffer, exists := s.getBuffer(clientID)
	if !exists {
		logger.Errorf(ctx, "client_id:%s|err:buffer not found", clientID)
		return
	}
	for data := range buffer {
		if err := s.sendToClient(clientID, data); err != nil {
			logger.Errorf(ctx, "client_id:%s|err:%v", clientID, err)
		}
	}
}

func (s *server) Broadcast(data []byte) {
	clientIDs := s.getClientIDs()
	for _, clientID := range clientIDs {
		if buffer, exists := s.getBuffer(clientID); exists {
			buffer <- data
		}
	}
}
