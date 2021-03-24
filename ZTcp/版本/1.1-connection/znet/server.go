package znet

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
	"zTcp/zInterface"
)

type Server struct {
	Name string

	IPVersion string

	IP string

	Port int
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
	log.Println("my server start", s.IP, s.Port)

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

			dealConn := NewConnection(conn, cid, CallBackToClient)
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

func NewServer(name string) zInterface.IServer {
	s := &Server{
		name,
		"tcp4",
		"0.0.0.0",
		7777,
	}
	return s
}
