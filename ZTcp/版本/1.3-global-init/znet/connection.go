package znet

import (
	"log"
	"net"
	"zTcp/zInterface"
)

type Connection struct {
	Conn *net.TCPConn

	ConnID uint32

	isClosed bool

	Router zInterface.IRouter

	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
}

func NewConnection(conn *net.TCPConn, connID uint32, router zInterface.IRouter) *Connection {

	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		Router:       router, //指向同一个router，但interface的值是不一样的
		ExitBuffChan: make(chan bool, 1),
	}
	log.Println(&c.Router)
	return c
}

func (c *Connection) StartReader() {
	log.Println("Reader Goroutine is running")
	defer log.Println(c.RemoteAddr().String(), " conn reader exit") //离开
	defer c.Stop()

	for {
		buf := make([]byte, 512)
		_, err := c.Conn.Read(buf)
		if err != nil {
			log.Println("recv buf err ", err, "conn ", c.ConnID)
			//c.ExitBuffChan <- true //这里直接调Stop是否更好
			c.Stop()
			return
		}

		req := Request{
			c,
			buf,
		}
		log.Println("&c.Router %p", &c.Router)
		go func(request zInterface.IRequest) {
			c.Router.PreHandle(request)
			c.Router.Handle(request)
			c.Router.PostHandle(request)
		}(&req)
	}
}

func (c *Connection) Start() {
	go c.StartReader()
	//阻塞等待
	for {
		select {
		case <-c.ExitBuffChan:
			{
				log.Println("start out") //子协程不随父协程退出
				return
			}
		}
	}
}

func (c *Connection) Stop() {
	//log.Println("conn stop ", c.ConnID)
	if c.isClosed {
		return
	}
	log.Println("conn stop 2 ", c.ConnID)
	c.isClosed = true

	c.Conn.Close()

	c.ExitBuffChan <- true

	close(c.ExitBuffChan)
}

func (c *Connection) GetTcpConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
