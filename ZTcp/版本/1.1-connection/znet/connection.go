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

	handleAPI zInterface.HandFunc

	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
}

func NewConnection(conn *net.TCPConn, connID uint32, callback_api zInterface.HandFunc) *Connection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		handleAPI:    callback_api,
		ExitBuffChan: make(chan bool, 1),
	}

	return c
}

func (c *Connection) StartReader() {
	log.Println("Reader Goroutine is running")
	defer log.Println(c.RemoteAddr().String(), " conn reader exit") //离开
	defer c.Stop()

	for {
		buf := make([]byte, 512)
		cnt, err := c.Conn.Read(buf)
		if err != nil {
			log.Panicln("recv buf err ", err, "conn ", c.ConnID)
			c.ExitBuffChan <- true //这里直接调close是否更好
		} else if err := c.handleAPI(c.Conn, buf, cnt); err != nil {
			log.Panicln("connID ", c.ConnID, " handle is error")
			c.ExitBuffChan <- true
			return
		}
	}
}

func (c *Connection) Start() {
	go c.StartReader()
	//阻塞等待
	for {
		select {
		case <-c.ExitBuffChan:
			return
		}
	}
}

func (c *Connection) Stop() {

	if c.isClosed {
		return
	}
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
