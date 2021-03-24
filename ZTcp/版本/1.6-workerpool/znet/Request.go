package znet

import (
	"zTcp/zInterface"
)

type Request struct {
	conn zInterface.IConnection
	msg  zInterface.IMessage
}

func (r *Request) GetConnection() zInterface.IConnection {
	return r.conn
}

func (r *Request) GetData() []byte {
	return r.msg.GetData()
}

func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgId()
}
