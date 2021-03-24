package znet

type Message struct {
	Id      uint32
	DataLen uint32
	Data    []byte
}

func NewMsgPackage(id uint32, data []byte) *Message {
	return &Message{
		id,
		uint32(len(data)),
		data,
	}
}

func (msg *Message) GetDataLen() uint32 {
	return msg.DataLen
}

func (msg *Message) GetData() []byte {
	return msg.Data
}

func (msg *Message) GetMsgId() uint32 {
	return msg.Id
}

func (msg *Message) SetDataLen(len uint32) {
	msg.DataLen = len
}

func (msg *Message) SetData(data []byte) {
	msg.Data = data
}

func (msg *Message) SetMsgId(id uint32) {
	msg.Id = id
}
