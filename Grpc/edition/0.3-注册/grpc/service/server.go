package service

import (
	"encoding/json"
	"errors"
	"grpc/grpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
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
	server.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

func (server *Server) serveCodec(c codec.Codec) {
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
		go server.handleRequest(c, req, sending, wg)
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

func (server *Server) handleRequest(c codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	//log.Println(req.h, req.argv.Elem())
	//req.replyv = reflect.ValueOf(fmt.Sprintf("rpc response %d", req.h.Seq))
	//server.sendResponse(c, req.h, req.replyv.Interface(), sending)
	err := req.svc.call(req.mtype, req.argv, req.replyv)
	if err != nil {
		//调用方法错误
		req.h.Error = err.Error()
		server.sendResponse(c, req.h, invalidRequest, sending)
		return
	}
	server.sendResponse(c, req.h, req.replyv.Interface(), sending)
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
