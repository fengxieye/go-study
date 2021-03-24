package zInterface

//定义接口，可以实现部分的函数
type IServer interface {
	Start()

	Stop()

	Serve()

	AddRouter(router IRouter)
}
