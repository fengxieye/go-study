package zInterface

import "net"

type IConnection interface {
	Start()

	Stop()

	GetTcpConnection() *net.TCPConn

	GetConnID() uint32

	RemoteAddr() net.Addr
}

type HandFunc func(*net.TCPConn, []byte, int) error
