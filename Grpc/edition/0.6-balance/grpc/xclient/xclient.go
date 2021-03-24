package xclient

import (
	"context"
	C "grpc/grpc/client"
	"grpc/grpc/service"
	"io"
	"log"
	"reflect"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *service.Option
	mu      sync.Mutex
	clients map[string]*C.Client
}

var _ io.Closer = (*XClient)(nil)

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	for key, client := range xc.clients {
		client.Close()
		delete(xc.clients, key)
	}
	return nil
}

func NewXClient(d Discovery, mode SelectMode, opt *service.Option) *XClient {
	return &XClient{d: d, mode: mode, opt: opt, clients: make(map[string]*C.Client)}
}

func (xc *XClient) dial(rpcAddr string) (*C.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	client, ok := xc.clients[rpcAddr]
	if ok && !client.IsAvailable() {
		client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
		log.Println("XClient.dial use old client")
	}

	if client == nil {
		var err error
		client, err = C.XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}

	return client, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}

	return client.Call(ctx, serviceMethod, args, reply)
}

func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}

	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error
	replyDone := reply == nil
	ctx, cancle := context.WithCancel(ctx)

	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var cloneReply interface{}
			if reply != nil {
				cloneReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			//仍然会发送请求和接收到结果,不使用分配策略
			err := xc.call(rpcAddr, ctx, serviceMethod, args, cloneReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancle() //ctx.done 通知所有协程
			}
			//已经回复了，不再返回结果
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(cloneReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}
