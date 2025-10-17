package session

import "net"

type ISession interface {
	Close() error
	CloseWithError(code uint16, reason string) error
	IsConnected() bool
	SendDatagram(data []byte) error
	ReceiveDatagram() ([]byte, error)
	GetRemoteAddr() net.Addr
}
