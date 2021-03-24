package znet

import (
	"errors"
	"io"
	"log"
	"net"
	"zTcp/zInterface"
)

type Connection struct {
	Conn *net.TCPConn

	ConnID uint32

	isClosed bool

	MsgHandler zInterface.IMsgHandler

	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool

	//读写的通道
	msgChan chan []byte
}

func NewConnection(conn *net.TCPConn, connID uint32, handler zInterface.IMsgHandler) *Connection {

	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   handler, //指向同一个router，但interface的值是不一样的
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte),
	}

	return c
}

func (c *Connection) StartReader() {
	log.Println("Reader Goroutine is running")
	defer log.Println(c.RemoteAddr().String(), " conn reader exit") //离开
	defer c.Stop()

	for {
		dp := NewDataPackString()

		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTcpConnection(), headData); err != nil {
			log.Println("ReadFull err ", err)
			break
		}

		msg, err := dp.UnPack(headData)
		if err != nil {
			log.Println("unpack error ", err)
			break
		}

		var data []byte
		if msg.GetDataLen() > 0 {
			//设置包长度的byte
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTcpConnection(), data); err != nil {
				log.Println("read msg error", err)
			} else {
				log.Println("recv data ", data)
			}
		}

		msg.SetData(data)

		req := Request{
			c,
			msg,
		}
		//log.Println("&c.Router %p",&c.Router)
		go c.MsgHandler.DoMsgHandler(&req)
	}
}

func (c *Connection) StartWriter() {
	defer log.Println(c.RemoteAddr().String(), " conn write exit")

	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				log.Println("send data error : ", err, " conn writer exit")
			} else {
				log.Println("send data suce ", data)
			}
		case <-c.ExitBuffChan:
			return
		}
	}
}

func (c *Connection) Start() {
	go c.StartReader()
	go c.StartWriter()
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

func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed {
		return errors.New("connection closed when send data")
	}

	dp := NewDataPackString()

	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		return errors.New("pack error msg")
	}

	//这样不会堵住，在协程里面写
	c.msgChan <- msg

	return nil
}
