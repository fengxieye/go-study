package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"grpc/grpc/codec"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.Type
	//
	ConnectTimeout time.Duration //客户端使用的连接时间
	HandleTimeout  time.Duration //由客户端请求传入
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
	//
	ConnectTimeout: time.Second * 10,
}

type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
	//
	mtype *methodType
	svc   *service
}

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

func Accept(listener net.Listener) {
	DefaultServer.Accept(listener)
}

var DefaultServer = NewServer()

//tcp 直连方式
func (server *Server) Accept(listener net.Listener) {
	log.Println("server accept start:")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("rpc server accept err :", err)
			return
		}

		go server.ServeConn(conn)
	}
}

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()

	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server options error: ", err)
		return
	}
	//根据codetype得到解码器
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server invalid codec type %s", opt.CodecType)
		return
	}
	server.serveCodec(f(conn), &opt)
}

var invalidRequest = struct{}{}

func (server *Server) serveCodec(c codec.Codec, opt *Option) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := server.readRequest(c)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			//得到请求头但读参数失败
			server.sendResponse(c, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(c, req, sending, wg, opt.HandleTimeout)
	}
	//等待写完成
	wg.Wait()
	c.Close()
}

func (server *Server) readRequestHeader(c codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := c.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("server readRequestHeader read err:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(c codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(c)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	//暂时默认为string
	//req.argv = reflect.New(reflect.TypeOf(""))
	//if err = c.ReadBody(req.argv.Interface()); err != nil{
	//	log.Println("server readRequest read argv err:", err)
	//}

	//获取类型
	req.svc, req.mtype, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		//read body 函数一定要是指针
		argvi = req.argv.Addr().Interface()
	}
	//log.Println("server.readRequest argv type", req.argv.Type())
	if err = c.ReadBody(argvi); err != nil {
		log.Println("server.readRequest : read body err:", err)
		return req, err
	}

	return req, nil
}

//写加锁
func (server *Server) sendResponse(c codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := c.Write(h, body); err != nil {
		log.Println("server sendResponse err:", err)
	}
}

//加入处理超时
func (server *Server) handleRequest(c codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	called := make(chan struct{})
	sent := make(chan struct{})

	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			server.sendResponse(c, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}

		server.sendResponse(c, req.h, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}

	select {
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server:requset timeout expect in %s", timeout)
		server.sendResponse(c, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
}

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		//存在
		return errors.New("server.Register : service has defined:" + s.name)
	}
	return nil
}

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

//server.findService
func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("server.findService service/method request wrong no . method:" + serviceMethod)
		return
	}

	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]

	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("server.findService: cant find service:" + serviceName)
		return
	}

	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("server.findService:cant find method" + serviceName + "." + methodName)
	}
	return
}

const (
	Connected        = "200 Connected to GRPC"
	DefaultRPCPath   = "/_grpc_"
	DefaultDebugPath = "/debug/grpc"
)

//http连接方式
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must connect\n")
		return
	}

	//conn 从管理的 conn 列表中去除,由我们接管socket的连接生命周期
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("server ServeHTPP hijacking ,", req.RemoteAddr, ",err:", err.Error())
		return
	}

	_, _ = io.WriteString(conn, "HTTP/1.0 "+Connected+"\n\n")
	server.ServeConn(conn)
}

func (server *Server) HandleHTTP() {
	http.Handle(DefaultDebugPath, debugHTTP{server})
	http.Handle(DefaultRPCPath, server)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
