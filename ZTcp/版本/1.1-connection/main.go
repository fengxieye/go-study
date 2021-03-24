package main

import (
	"zTcp/znet"
)

func main() {
	server := znet.NewServer("mytcp test")
	server.Serve()
}
