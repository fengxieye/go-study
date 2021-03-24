package znet

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
	"zTcp/utils"
	"zTcp/zInterface"
)

type Server struct {
	Name string

	IPVersion string

	IP string

	Port int

	msgHandler zInterface.IMsgHandler
}

func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
	log.Println("conn handle")
	if _, err := conn.Write(data[:cnt]); err != nil {
		log.Panicln("write back buf err ", err)
		return errors.New("call back to client error")
	}
	return nil
}

func (s *Server) Start() {
	log.Printf("[START] Server name: %s,listenner at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)
	log.Printf("[Zinx] Version: %s, MaxConn: %d,  MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)

	go func() {
		//获取tcp addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			log.Println("ResolveIPAddr tcp addr err ", err)
			return
		}

		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			log.Println("listen wrong ip ", s.IPVersion, " err ", err)
		}

		log.Println("server listening")

		var cid uint32
		cid = 0

		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				log.Println("accept err ", err)
				continue
			}
			log.Println("get connect ", conn)

			dealConn := NewConnection(conn, cid, s.msgHandler)
			cid++

			go dealConn.Start()
		}

	}()
}

func (s *Server) Stop() {
	log.Println("my server Stop")
}

func (s *Server) Serve() {
	log.Println("my server Serve")
	s.Start()

	for {
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) AddRouter(msgId uint32, router zInterface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
	log.Println("add router succ ")
}

func NewServer() zInterface.IServer {
	s := &Server{
		utils.GlobalObject.Name,
		"tcp4",
		utils.GlobalObject.Host,
		utils.GlobalObject.TcpPort,
		NewMsgHandler(),
	}
	return s
}
