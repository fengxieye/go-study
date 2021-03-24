package znet

import (
	"zTcp/zInterface"
)

type Request struct {
	conn zInterface.IConnection
	data []byte
}

func (r *Request) GetConnection() zInterface.IConnection {
	return r.conn
}

func (r *Request) GetData() []byte {
	return r.data
}
