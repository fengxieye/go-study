package znet

import (
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

		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				log.Println("accept err ", err)
				continue
			}
			log.Println("get connect ", conn)
			go func() {
				for {
					buf := make([]byte, 512)
					cnt, err := conn.Read(buf)
					if err != nil {
						log.Println("recv buf err ", err)
						continue
					}

					if _, err := conn.Write(buf[:cnt]); err != nil {
						log.Println("write buf err ", err)
						continue
					}
				}
			}()
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
