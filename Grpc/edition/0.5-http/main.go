package main

import (
	"golang.org/x/net/context"
	"grpc/grpc/client"
	"grpc/grpc/service"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func startServer(addr chan string) {
	var foo Foo
	if err := service.Register(&foo); err != nil {
		log.Fatal("register err：", err)
	}

	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal("startServer listen err:", err)
	}

	service.HandleHTTP()
	addr <- l.Addr().String()
	http.Serve(l, nil)

	////随机端口
	//l,err := net.Listen("tcp", ":0")
	//if err != nil{
	//	log.Fatal("startServer listen err:", err)
	//}
	//
	//log.Println("startServer rpc server on:", l.Addr())
	//addr <- l.Addr().String()
	//service.Accept(l)
}

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func call(addrCh chan string) {
	//addr := make(chan string)
	//go startServer(addr)
	//等待服务器建立后建立client
	client, _ := client.DialHTTP("tcp", <-addrCh)
	defer func() { client.Close() }()

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{i, i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("main: call Foo.Sum error:", err)
			}
			log.Println("reply:", reply, ",i:", i)
		}(i)
	}
	wg.Wait()
}

func main() {
	ch := make(chan string)
	go call(ch)
	startServer(ch)
}
