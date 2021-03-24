package znet

import (
	"log"
	"zTcp/zInterface"
)

type MsgHandler struct {
	Apis map[uint32]zInterface.IRouter
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{Apis: make(map[uint32]zInterface.IRouter)}
}

func (m *MsgHandler) DoMsgHandler(request zInterface.IRequest) {
	handler, ok := m.Apis[request.GetMsgID()]
	if !ok {
		log.Println("api msgid = ", request.GetMsgID(), " is not found")
		return
	}

	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

func (m *MsgHandler) AddRouter(msgId uint32, router zInterface.IRouter) {
	if _, ok := m.Apis[msgId]; ok {
		log.Println("repeated api, msgid = ", msgId)
		return
	}

	m.Apis[msgId] = router
	log.Println("add api msgid = ", msgId)
}
