package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"grpc/grpc/codec"
	"grpc/grpc/service"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"sync"
	"time"
)

type clientResult struct {
	client *Client
	err    error
}

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call
}

//结束通知
func (call *Call) done() {
	call.Done <- call
}

type Client struct {
	c        codec.Codec
	opt      *service.Option
	sending  sync.Mutex //发送为协程，接收是统一的
	header   codec.Header
	mu       sync.Mutex //pending 锁
	seq      uint64     //发送的call index
	pending  map[uint64]*Call
	closeing bool
	shutdown bool
}

var _ io.Closer = (*Client)(nil)

func NewClient(conn net.Conn, opt *service.Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("client.NewClient codec type err %s", opt.CodecType)
		log.Println("client.NewClient codec type err : ", opt.CodecType)
		return nil, err
	}

	//send option
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("client.NewClient opt encode err : ", err)
		conn.Close()
		return nil, err
	}

	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(c codec.Codec, opt *service.Option) *Client {
	client := &Client{
		seq:     1,
		c:       c,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closeing {
		return rpc.ErrShutdown
	}

	client.closeing = true
	return client.c.Close()
}

func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closeing
}

func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closeing || client.shutdown {
		return 0, rpc.ErrShutdown
	}
	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()

	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.c.ReadHeader(&h); err != nil {
			break
		}
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			err = client.c.ReadBody(nil)
		//服务器读取出错
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			log.Println("Client.receive:h.error", h.Error)
			err = client.c.ReadBody(nil)
			call.done()
		default:
			err = client.c.ReadBody(call.Reply)
			if err != nil {
				log.Println("Client.receive:ReadBody.error", err)
				call.Error = errors.New("client.receive read body err:" + err.Error())
			}
			call.done()
		}
	}
	//client发生错误，所有call停止
	client.terminateCalls(err)
}

func (client *Client) send(call *Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""
	//todo
	if err := client.c.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 1)
	} else if cap(done) == 0 {
		log.Panic("client.Go rpc client done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	//wait
	client.send(call)
	return call
}

//rpc请求,请求超时
func (client *Client) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	//等待写完和读完或者出错返回
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return errors.New("client.call: call fail :" + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
	return call.Error
}

func parseOptions(opts ...*service.Option) (*service.Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return service.DefaultOption, nil
	}

	if len(opts) != 1 {
		return nil, errors.New("client.parseOptions number is > 1")
	}

	opt := opts[0]
	opt.MagicNumber = service.DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = service.DefaultOption.CodecType
	}

	return opt, nil
}

type newClientFunc func(conn net.Conn, opt *service.Option) (client *Client, err error)

func dialTimeout(f newClientFunc, network, address string, opts ...*service.Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	ch := make(chan clientResult)
	go func() {
		client, err := f(conn, opt)
		//创建后返回结果
		ch <- clientResult{client: client, err: err}
	}()

	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}

	//超时或返回结果
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("clinet.dialTimeout connect timeout:%s", opt.ConnectTimeout)
	case result := <-ch:
		return result.client, result.err
	}
}

//链接服务器，发送option识别，建立client
func Dial(netWork, address string, opts ...*service.Option) (client *Client, err error) {
	return dialTimeout(NewClient, netWork, address, opts...)
}

func NewHTTPClient(conn net.Conn, opt *service.Option) (*Client, error) {
	_, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", service.DefaultRPCPath))

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})

	//在原来的方式上加入http的connect请求
	if err == nil && resp.Status == service.Connected {
		return NewClient(conn, opt)
	}

	if err == nil {
		err = errors.New("client.NewHTTPClient, unexpected HTTP response:" + resp.Status)
	}

	return nil, err
}

func DialHTTP(network, address string, opts ...*service.Option) (*Client, error) {
	return dialTimeout(NewHTTPClient, network, address, opts...)
}

func XDial(addr string, opts ...*service.Option) (*Client, error) {
	parts := strings.Split(addr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err : wrong fromat '%s'", addr)
	}

	protocal, addr := parts[0], parts[1]

	switch protocal {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		//tcp, unix
		return Dial(protocal, addr, opts...)
	}
}
