package znet

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"zTcp/utils"
	"zTcp/zInterface"
)

type DataPackString struct {
}

func NewDataPackString() *DataPackString {
	return &DataPackString{}
}

func (dp *DataPackString) GetHeadLen() uint32 {
	return 8
}

//使用string的话数字大小是按十进制位数，而不是字节位数
//原来4位的数字现在最大只能9999了
func (dp *DataPackString) Pack(msg zInterface.IMessage) ([]byte, error) {
	dataBuff := bytes.NewBuffer([]byte{})

	string := fmt.Sprintf("%04v", msg.GetDataLen())
	string += fmt.Sprintf("%04v", msg.GetMsgId())

	dataBuff.WriteString(string)
	dataBuff.Write(msg.GetData())
	return dataBuff.Bytes(), nil
}

func (dp *DataPackString) UnPack(data []byte) (zInterface.IMessage, error) {
	msg := &Message{}

	var str string = string(data[0:4])

	val, _ := strconv.ParseUint(str, 10, 32)
	msg.DataLen = uint32(val)

	str = string(data[4:8])
	val, _ = strconv.ParseUint(str, 10, 32)
	msg.Id = uint32(val)

	if utils.GlobalObject.MaxPacketSize > 0 && msg.DataLen > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("too large msg data reveived")
	}

	return msg, nil
}
