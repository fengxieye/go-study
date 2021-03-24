package znet

import (
	"errors"
	"log"
	"sync"
	"zTcp/zInterface"
)

type ConnManager struct {
	connections map[uint32]zInterface.IConnection

	connLock sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]zInterface.IConnection),
	}
}

func (t *ConnManager) Add(conn zInterface.IConnection) {
	t.connLock.Lock()
	defer t.connLock.Unlock()

	t.connections[conn.GetConnID()] = conn

	log.Println("connections add connid = ", conn.GetConnID(), " totallen = ", t.Len())
}

func (t *ConnManager) Remove(conn zInterface.IConnection) {
	t.connLock.Lock()
	defer t.connLock.Unlock()

	delete(t.connections, conn.GetConnID())

	log.Println("romove conn connid = ", conn.GetConnID())
}

func (t *ConnManager) Get(connID uint32) (zInterface.IConnection, error) {

	t.connLock.Lock()
	defer t.connLock.Unlock()

	if conn, ok := t.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

func (t *ConnManager) Len() int {
	return len(t.connections)
}

func (t *ConnManager) ClearConn() {
	t.connLock.Lock()
	defer t.connLock.Unlock()

	for connID, conn := range t.connections {
		conn.Stop()
		delete(t.connections, connID)
	}

	log.Println("clear connections")
}
