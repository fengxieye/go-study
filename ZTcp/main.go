package main

import (
	"fmt"
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
//	_,err := request.GetConnection().GetTcpConnection().Write([]byte("before ping0 ... \n"))
//	if err != nil{
//		log.Println("call back ping error")
//	}
//}

func (t *PingRouter) Handle(request zInterface.IRequest) {
	t.test++
	log.Printf("call PingRouter handle %p", t)
	err := request.GetConnection().SendBuffMsg(1, []byte("ping1..."))
	if err != nil {
		log.Println("call back PingRouter error")
	}
}

func (t *PingRouter) PostHandle(request zInterface.IRequest) {
	log.Println("call PingRouter PostHandle")
	err := request.GetConnection().SendBuffMsg(1, []byte("ping2..."))
	if err != nil {
		log.Println("call back PingRouter error")
	}
}

type HelloZinRouter struct {
	znet.BaseRouter
}

func (t *HelloZinRouter) Handle(request zInterface.IRequest) {
	log.Println("HelloZinRouter Handle")
	log.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendBuffMsg(1, []byte("hello zinx router v0.6"))
	if err != nil {
		log.Println(err)
	}
}

func ConnectionBegin(conn zInterface.IConnection) {
	log.Println("ConnectionBegin")
	err := conn.SendMsg(2, []byte("begin connection 111"))

	conn.SetProperty("Name", "test")
	if name, err := conn.GetProperty("Name"); err == nil {
		log.Println("conn name = ", name)
	}
	if err != nil {
		log.Println(err)
	}
}

func ConnectionLost(conn zInterface.IConnection) {
	if name, err := conn.GetProperty("Name"); err == nil {
		log.Println("conn name = ", name)
	}

	log.Println("ConnectionLost")
}

func main() {
	server := znet.NewServer()
	server.AddRouter(0, &PingRouter{})
	server.AddRouter(1, &HelloZinRouter{})

	server.SetOnConnStart(ConnectionBegin)
	server.SetOnConnStop(ConnectionLost)

	m := make(map[string]interface{})
	m["name"] = "test"
	if value, ok := m["name"]; ok {
		fmt.Println(value)
	} else {
		fmt.Println("property not found ")
	}

	server.Serve()
}
