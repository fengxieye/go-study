package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"grpc/grpc/codec"
	"grpc/grpc/service"
	"io"
	"log"
	"net"
	"net/rpc"
	"sync"
)

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
			err = client.c.ReadBody(nil)
			call.done()
		default:
			err = client.c.ReadBody(call.Reply)
			if err != nil {
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

//rpc请求
func (client *Client) Call(serviceMethod string, args, reply interface{}) error {
	//等待写完和读完或者出错返回
	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
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

//链接服务器，发送option识别，建立client
func Dial(netWork, address string, opts ...*service.Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial(netWork, address)
	if err != nil {
		return nil, err
	}

	defer func() {
		if client == nil {
			conn.Close()
		}
	}()
	return NewClient(conn, opt)
}
