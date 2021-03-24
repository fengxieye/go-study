package zInterface

//定义接口，可以实现部分的函数
type IServer interface {
	Start()

	Stop()

	Serve()

	AddRouter(msgID uint32, router IRouter)

	GetConnMgr() IConnManager

	SetOnConnStart(func(IConnection))

	SetOnConnStop(func(IConnection))

	CallOnConnStart(conn IConnection)

	CallOnConnStop(conn IConnection)
}
