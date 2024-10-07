package resp

import (
	"bytes"
	"fmt"
)

const (
	SIMPLESTRING byte = '+'
	ERRORS       byte = '-'
	INTEGERS     byte = ':'
	BULKSTRINGS  byte = '$'
	ARRAYS       byte = '*'
	NULL         byte = '_'
)

type RespDataType struct {
	Value interface{}
	Type  byte
}

type RespSerializer struct{}

func NewRespSerializer() RespSerializer {
	return RespSerializer{}
}

func (rs RespSerializer) SimpleString(s string) []byte {
	return []byte(fmt.Sprintf("%v%s\r\n", string(SIMPLESTRING), s))
}

func (rs RespSerializer) SimpleErrors(err string) []byte {
	return []byte(fmt.Sprintf("%s%s\r\n", string(ERRORS), err))
}

func (rs RespSerializer) Integers(num int) []byte {
	return []byte(fmt.Sprintf("%s%d\r\n", string(INTEGERS), num))
}

func (rs RespSerializer) BulkString(s string) []byte {
	return []byte(fmt.Sprintf("%s%d\r\n%s\r\n", string(BULKSTRINGS), len(s), s))
}

func (rs RespSerializer) Arrays(arr []*RespDataType) []byte {

	prefix := fmt.Sprintf("%s%d\r\n", string(ARRAYS), len(arr))

	buf := bytes.NewBuffer([]byte(prefix))

	for _, respData := range arr {
		switch respData.Type {
		case SIMPLESTRING:
			buf.Write(rs.SimpleString(respData.Value.(string)))
		case ERRORS:
			buf.Write(rs.SimpleErrors(respData.Value.(string)))
		case INTEGERS:
			buf.Write(rs.Integers(respData.Value.(int)))
		case BULKSTRINGS:
			buf.Write(rs.BulkString(respData.Value.(string)))
		case NULL:
			buf.Write(rs.Null())
		case ARRAYS:
			buf.Write(rs.Arrays(respData.Value.([]*RespDataType)))

		}
	}

	return buf.Bytes()
}

func (rs RespSerializer) Null() []byte {
	return []byte("_\r\n")
}
