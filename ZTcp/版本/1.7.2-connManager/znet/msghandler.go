package znet

import (
	"log"
	"zTcp/utils"
	"zTcp/zInterface"
)

type MsgHandler struct {
	Apis map[uint32]zInterface.IRouter

	WorkerPoolSize uint32

	TaskQueue []chan zInterface.IRequest
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Apis:           make(map[uint32]zInterface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		TaskQueue:      make([]chan zInterface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
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

func (m *MsgHandler) StartOneWorker(workerID int, taskQueue chan zInterface.IRequest) {
	log.Println("worker start id = ", workerID)
	for {
		select {
		case request := <-taskQueue:
			m.DoMsgHandler(request)
		}
	}
}

//启动多个go，读取对应的chan
func (m *MsgHandler) StartWorkerPool() {
	for i := 0; i < int(m.WorkerPoolSize); i++ {
		m.TaskQueue[i] = make(chan zInterface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		go m.StartOneWorker(i, m.TaskQueue[i])
	}
}

func (m *MsgHandler) SendMsgToTaskQueue(request zInterface.IRequest) {
	workerID := request.GetConnection().GetConnID() % m.WorkerPoolSize
	log.Println("add requset conn id ", request.GetConnection().GetConnID(), " request msgID=", request.GetMsgID(), "to workerID=", workerID)

	m.TaskQueue[workerID] <- request
}
