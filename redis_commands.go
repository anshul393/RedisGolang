package main

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"redis-server/resp"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type StoreData struct {
	Value  string
	Expiry time.Time
}

// func (st *s)

func (rs *RedisServer) handleCommand(conn io.Writer, command []*resp.RespDataType) {

	rs.Mu.Lock()
	defer rs.Mu.Unlock()

	recvCom := command[0].Value.(string)
	// rs.Logger.Infow("commands", ":::", command)
	var err error

	switch strings.ToUpper(recvCom) {
	case "PING":
		_, err = conn.Write(rs.handlePing())
	case "COMMAND":
		_, err = conn.Write(rs.Serializer.Arrays([]*resp.RespDataType{}))
	case "ECHO":
		_, err = conn.Write(rs.handleEcho(command))
	case "SET":
		_, err = conn.Write(rs.handleSet(command))
	case "GET":
		_, err = conn.Write(rs.handleGet(command))
	case "CONFIG":
		_, err = conn.Write(rs.handleConfig(command))
	case "EXISTS":
		_, err = conn.Write(rs.handleExists(command))
	case "DEL":
		_, err = conn.Write(rs.handleDelete(command))
	case "INCR":
		_, err = conn.Write(rs.handleINCR(command))
	case "DECR":
		_, err = conn.Write(rs.handleDECR(command))
	case "LPUSH":
		_, err = conn.Write(rs.handleLPush(command))
	case "RPUSH":
		_, err = conn.Write(rs.handleRPush(command))
	case "SAVE":
		_, err = conn.Write(rs.handleSave())
	}

	if err != nil {
		// rs.Logger.Infow("error in writing to the connection", "error", err)
	}

}

func (rs *RedisServer) handlePing() []byte {
	return rs.Serializer.SimpleString("PONG")
}

func (rs *RedisServer) handleEcho(command []*resp.RespDataType) []byte {
	sb := &strings.Builder{}
	for _, args := range command[1:] {
		s := args.Value.(string)
		_, err := sb.WriteString(s + " ")
		if err != nil {
			rs.Logger.Infow("error writing to string builder", "function", "handleEcho", "error", err)
			break
		}
	}
	return rs.Serializer.SimpleString(sb.String())
}

func (rs *RedisServer) handleSet(command []*resp.RespDataType) []byte {

	key, value := command[1].Value.(string), command[2]
	var ExpiryTime time.Time

	if len(command) > 3 && len(command) <= 5 {

		tVal, err := strconv.Atoi(command[4].Value.(string))
		if err != nil {
			return rs.Serializer.SimpleErrors("invalid expiry time")
		}
		switch command[3].Value.(string) {
		case "EX":
			ExpiryTime = time.Now().Add(time.Duration(tVal) * time.Second)
		case "PX":
			ExpiryTime = time.Now().Add(time.Duration(tVal) * time.Millisecond)
		case "EXAT":
			ExpiryTime = time.Unix(int64(tVal), 0)
		case "PXAT":
			ExpiryTime = time.UnixMilli(int64(tVal))

		}
	}

	s := &StoreData{
		Value:  value.Value.(string),
		Expiry: ExpiryTime,
	}

	rs.Store[key] = s

	return rs.Serializer.SimpleString("OK")
}

func (rs *RedisServer) handleGet(command []*resp.RespDataType) []byte {

	key := command[1].Value.(string)
	v, ok := rs.Store[key]
	if !ok {
		return rs.Serializer.Null()
	}

	if reflect.TypeOf(v).String() != "*main.StoreData" {
		return rs.Serializer.SimpleErrors("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	val := v.(*StoreData)
	isExpired, isValid := Validate(val)
	if isExpired {
		delete(rs.Store, key)
	}

	if isValid {
		return rs.Serializer.BulkString(val.Value)
	}

	return nil

}

func (rs *RedisServer) handleConfig(command []*resp.RespDataType) []byte {
	if len(command) < 2 {
		return rs.Serializer.SimpleErrors("ERR wrong number of arguments for 'config' command")
	}

	subcommand := strings.ToUpper(command[1].Value.(string))
	switch subcommand {
	case "GET":
		if len(command) < 3 {
			return rs.Serializer.SimpleErrors("ERR wrong number of arguments for 'config get' command")
		}
		parameter := command[2].Value.(string)
		return rs.handleConfigGet(parameter)
	default:
		return rs.Serializer.SimpleErrors("ERR unsupported CONFIG subcommand")
	}
}

func (rs *RedisServer) handleConfigGet(parameter string) []byte {
	switch parameter {
	case "save":
		return rs.Serializer.SimpleString("")

	default:
		return rs.Serializer.Arrays([]*resp.RespDataType{})
	}
}

func (rs *RedisServer) handleExists(command []*resp.RespDataType) []byte {
	count := 0
	for _, cmd := range command[1:] {
		key := cmd.Value.(string)
		if _, ok := rs.Store[key]; ok {
			count += 1
		}
	}

	return rs.Serializer.Integers(count)
}

func (rs *RedisServer) handleDelete(command []*resp.RespDataType) []byte {
	count := 0
	for _, cmd := range command[1:] {
		key := cmd.Value.(string)
		if _, ok := rs.Store[key]; ok {
			delete(rs.Store, key)
			count += 1
		}
	}

	return rs.Serializer.Integers(count)
}

func (rs *RedisServer) handleINCR(command []*resp.RespDataType) []byte {
	key := command[1].Value.(string)
	val, ok := rs.Store[key]

	if !ok {
		s := &StoreData{Value: "1"}
		rs.Store[key] = s
		return rs.Serializer.Integers(1)
	}

	s := val.(*StoreData)
	isExpired, _ := Validate(s)

	if isExpired {
		s := &StoreData{Value: "1"}
		rs.Store[key] = s
		return rs.Serializer.Integers(1)
	}

	n, err := strconv.Atoi(s.Value)
	if err != nil {
		return rs.Serializer.SimpleErrors("invalid operation or out of range")
	}

	n, err = Incr(n)
	if err != nil {
		return rs.Serializer.SimpleErrors("invalid operation or out of range")
	}
	s.Value = strconv.Itoa(n)
	rs.Store[key] = s

	return rs.Serializer.Integers(n)

}

func (rs *RedisServer) handleDECR(command []*resp.RespDataType) []byte {
	key := command[1].Value.(string)
	val, ok := rs.Store[key]

	if !ok {
		s := &StoreData{Value: "-1"}
		rs.Store[key] = s
		return rs.Serializer.Integers(-1)
	}

	s := val.(*StoreData)
	isExpired, _ := Validate(s)

	if isExpired {
		s := &StoreData{Value: "-1"}
		rs.Store[key] = s
		return rs.Serializer.Integers(-1)
	}

	n, err := strconv.Atoi(s.Value)
	if err != nil {
		return rs.Serializer.SimpleErrors("invalid operation or out of range")
	}

	n, err = Decr(n)
	if err != nil {
		return rs.Serializer.SimpleErrors("invalid operation or out of range")
	}
	s.Value = strconv.Itoa(n)
	rs.Store[key] = s

	return rs.Serializer.Integers(n)
}

func (rs *RedisServer) handleLPush(command []*resp.RespDataType) []byte {

	key := command[1].Value.(string)
	value, ok := rs.Store[key]

	if ok && reflect.TypeOf(value).String() != "*main.List" {
		return rs.Serializer.SimpleErrors("type other than list is present")
	}

	var list *List

	if !ok {
		list = NewList()
	} else {
		list = value.(*List)
	}

	for _, e := range command[2:] {
		list.LPush(e.Value)
	}
	rs.Store[key] = list

	return rs.Serializer.Integers(list.Len())
}

func (rs *RedisServer) handleRPush(command []*resp.RespDataType) []byte {
	key := command[1].Value.(string)
	value, ok := rs.Store[key]

	if ok && reflect.TypeOf(value).String() != "*main.List" {
		return rs.Serializer.SimpleErrors("type other than list is present")
	}

	var list *List
	if !ok {
		list = NewList()
	} else {
		list = value.(*List)
	}

	for _, e := range command[2:] {
		list.RPush(e.Value)
	}
	rs.Store[key] = list

	return rs.Serializer.Integers(list.Len())
}

func (rs *RedisServer) handleSave() []byte {
	f, err := os.Create(rs.File)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	if err := enc.Encode(rs.Store); err != nil {
		log.Fatal(err)
	}

	return rs.Serializer.SimpleString("OK")
}
