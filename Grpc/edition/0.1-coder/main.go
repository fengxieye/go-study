package main

import (
	"encoding/json"
	"fmt"
	"grpc/grpc/codec"
	"grpc/grpc/service"
	"log"
	"net"
	"time"
)

func startServer(addr chan string) {
	//随机端口
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("startServer listen err:", err)
	}

	log.Println("startServer rpc server on:", l.Addr())
	addr <- l.Addr().String()
	service.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)
	//等待服务器建立后建立client
	conn, _ := net.Dial("tcp", <-addr)
	defer func() { conn.Close() }()

	time.Sleep(time.Second)
	json.NewEncoder(conn).Encode(service.DefaultOption)
	c := codec.NewGobCodec(conn)
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}

		c.Write(h, fmt.Sprintf("grpc req %d", h.Seq))
		c.ReadHeader(h)
		var reply string
		c.ReadBody(&reply)
		log.Println("reply:", reply)
	}
}
