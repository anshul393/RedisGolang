package resp

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var errInvalidSimpleString = errors.New("invalid simple string")
var errInvalidBulkString = errors.New("invlaid bulk string")
var errInvalidErrors = errors.New("invlaid simple errors")
var errInvalidIntegers = errors.New("invalid integers")
var errInvalidArrays = errors.New("invalid arrays")

type RespDeserializer struct{}

func NewRespDeserializer() RespDeserializer {
	return RespDeserializer{}
}

func (rd RespDeserializer) Deserialize(data []byte) (*RespDataType, error) {
	var result *RespDataType
	var err error

	switch data[0] {
	case SIMPLESTRING:
		result, _, err = rd.SimpleString(data, 0)
	case ERRORS:
		result, _, err = rd.SimpleErrors(data, 0)
	case INTEGERS:
		result, _, err = rd.Integers(data, 0)
	case BULKSTRINGS:
		result, _, err = rd.BulkString(data, 0)
	case ARRAYS:
		result, _, err = rd.Arrays(data, 0)
	}

	return result, err
}

func (rd RespDeserializer) SimpleString(data []byte, start int) (*RespDataType, int, error) {
	if data[start] != SIMPLESTRING {
		return nil, -1, errInvalidSimpleString
	}

	data = data[start:]
	crlfIdx := bytes.Index(data, []byte("\r\n"))

	nextIndex := start + crlfIdx + 2
	return &RespDataType{Type: SIMPLESTRING, Value: string(data[1:crlfIdx])}, nextIndex, nil
}

func (rd RespDeserializer) SimpleErrors(data []byte, start int) (*RespDataType, int, error) {
	if data[start] != ERRORS {
		return nil, -1, errInvalidErrors
	}

	data = data[start:]
	crlfIdx := bytes.Index(data, []byte("\r\n"))

	nextIndex := start + crlfIdx + 2
	return &RespDataType{Type: ERRORS, Value: string(data[1:crlfIdx])}, nextIndex, nil
}

func (rd RespDeserializer) Integers(data []byte, start int) (*RespDataType, int, error) {
	if data[start] != INTEGERS {
		return nil, -1, errInvalidIntegers
	}

	data = data[start:]
	crlfIdx := bytes.Index(data, []byte("\r\n"))
	num, err := strconv.Atoi(string(data[1:crlfIdx]))
	if err != nil {
		return nil, -1, errInvalidIntegers
	}

	nextIndex := start + crlfIdx + 2
	return &RespDataType{Type: INTEGERS, Value: num}, nextIndex, nil
}

func (rd RespDeserializer) BulkString(data []byte, start int) (*RespDataType, int, error) {
	if data[start] != BULKSTRINGS {
		return nil, -1, errInvalidBulkString
	}

	data = data[start:]

	crlfIdx := bytes.Index(data, []byte("\r\n"))
	length, err := strconv.Atoi(string(data[1:crlfIdx]))
	if err != nil {
		return nil, -1, errInvalidBulkString
	}

	nextIndex := start + crlfIdx + 2 + length + 2
	return &RespDataType{Type: BULKSTRINGS, Value: string(data[crlfIdx+2 : crlfIdx+2+length])}, nextIndex, nil
}

// *length\r\nelement1...elementN
func (rd RespDeserializer) Arrays(data []byte, start int) (*RespDataType, int, error) {
	if data[start] != ARRAYS {
		return nil, -1, errInvalidArrays
	}

	data = data[start:]

	crlfIdx := bytes.Index(data, []byte("\r\n"))
	arrL, err := strconv.Atoi(string(data[1:crlfIdx]))
	if err != nil {
		return nil, -1, errInvalidArrays
	}

	arr := make([]*RespDataType, arrL)
	arrIdx := 0

	next := crlfIdx + 2
	deserializeFuncs := map[byte]func([]byte, int) (*RespDataType, int, error){
		SIMPLESTRING: rd.SimpleString,
		ERRORS:       rd.SimpleErrors,
		INTEGERS:     rd.Integers,
		BULKSTRINGS:  rd.BulkString,
		ARRAYS:       rd.Arrays,
	}

	for arrIdx < arrL {
		deserializeFunc, ok := deserializeFuncs[data[next]]
		if !ok {
			return nil, -1, fmt.Errorf("unknown: %c", data[next])
		}

		r, nextIdx, err := deserializeFunc(data, next)
		if err != nil {
			return nil, -1, err
		}

		next = nextIdx
		arr[arrIdx] = r
		arrIdx++
	}

	return &RespDataType{Type: ARRAYS, Value: arr}, start + next, nil
}
