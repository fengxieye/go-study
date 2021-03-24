package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"zTcp/zInterface"
)

type GlobalObj struct {
	TcpServer zInterface.IServer
	Host      string
	TcpPort   int
	Name      string
	Version   string

	MaxPacketSize uint32
	MaxConn       int
}

var GlobalObject *GlobalObj

func init() {
	GlobalObject = &GlobalObj{
		Name:          "ZinxServerApp",
		Version:       "v0.4",
		TcpPort:       7777,
		Host:          "0.0.0.0",
		MaxConn:       12000,
		MaxPacketSize: 4096,
	}

	GlobalObject.Reload()
}

func (g *GlobalObj) Reload() {
	data, err := ioutil.ReadFile("conf/zinx.json")
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		log.Println(err)
	}
}
