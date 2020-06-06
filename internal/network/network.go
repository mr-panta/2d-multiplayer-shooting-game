package network

const (
	listenBufferSize = 1000
	poolSize         = 8
	clientIDLength   = 16
)

type Client interface {
	Start() error
	Wait()
	Close() error
	Send(req []byte) (resp []byte, err error)
	Listen() <-chan []byte
}

type Server interface {
	Start() error
	Wait()
	Close() error
	Broadcast(data []byte)
}

type Process func(req []byte) (resp []byte)
