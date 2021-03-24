package main

import (
	"log"
	"zTcp/zInterface"
	"zTcp/znet"
)

type PingRouter struct {
	znet.BaseRouter
	test int
}

//func (t *PingRouter)PreHandle(request zInterface.IRequest)  {
//	log.Println("call router prehandle")
//	_,err := request.GetConnection().GetTcpConnection().Write([]byte("before ping ... \n"))
//	if err != nil{
//		log.Println("call back ping error")
//	}
//}

func (t *PingRouter) Handle(request zInterface.IRequest) {
	t.test++
	log.Printf("call router handle %p", t)
	_, err := request.GetConnection().GetTcpConnection().Write([]byte(" ping ... \n"))
	if err != nil {
		log.Println("call back ping error")
	}
}

func (t *PingRouter) PostHandle(request zInterface.IRequest) {
	log.Println("call router PostHandle")
	_, err := request.GetConnection().GetTcpConnection().Write([]byte("after ping ... \n"))
	if err != nil {
		log.Println("call back ping error")
	}
}

func main() {
	server := znet.NewServer("mytcp test")
	server.AddRouter(&PingRouter{})
	server.Serve()
}
