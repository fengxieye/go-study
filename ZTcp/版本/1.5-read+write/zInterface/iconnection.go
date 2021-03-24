package zInterface

import "net"

type IConnection interface {
	Start()

	Stop()

	GetTcpConnection() *net.TCPConn

	GetConnID() uint32

	RemoteAddr() net.Addr

	SendMsg(msgID uint32, data []byte) error
}

type HandFunc func(*net.TCPConn, []byte, int) error
