package main

import (
	"fmt"
	"grpc/grpc/client"
	"grpc/grpc/service"
	"log"
	"net"
	"sync"
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
	client, _ := client.Dial("tcp", <-addr)
	defer func() { client.Close() }()

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("grpc req %d", i)
			var reply string
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
	}
	wg.Wait()
}
