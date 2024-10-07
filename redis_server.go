package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"redis-server/resp"
	"reflect"
	"sync"

	"go.uber.org/zap"
)

type RedisServer struct {
	Addr              string
	Serializer        resp.RespSerializer
	Deserializer      resp.RespDeserializer
	Logger            *zap.SugaredLogger
	MaxReadBufferSize int
	Store             map[string]any
	Mu                *sync.RWMutex
	File              string
}

func NewRedisServer(Addr string, filename string) *RedisServer {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil
	}
	defer logger.Sync()

	f, err := os.OpenFile(filename, os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	gob.Register(&StoreData{})
	gob.Register(&List{})

	rs := &RedisServer{
		Addr:              Addr,
		Serializer:        resp.NewRespSerializer(),
		Deserializer:      resp.NewRespDeserializer(),
		Logger:            logger.Sugar(),
		MaxReadBufferSize: 1024,
		Store:             map[string]any{},
		Mu:                &sync.RWMutex{},
		File:              filename,
	}

	if err := rs.Load(); errors.Is(err, io.EOF) {
		rs.Logger.Infow("no data to load")
	} else if err != nil {
		log.Fatal(err)
	}

	for key, value := range rs.Store {
		fmt.Printf("%#v::%#v\n", key, value)
	}

	return rs
}

func (rs *RedisServer) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}

func (rs *RedisServer) Accept(l net.Listener) {
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			rs.Logger.Infow("failed to accept the connection")
		}

		go rs.HandleConn(conn)
	}
}

// It will handle the connection
// all the writes from the redis-cli are in form of array
func (rs *RedisServer) HandleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	// writer := bufio.NewWriter(conn)

	// defer writer.Flush()

	buf := make([]byte, rs.MaxReadBufferSize)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			// rs.Logger.Infow("failed to read conn", "connAddress", conn.LocalAddr(), "error", err)
			return
		}

		serData := buf[:n]
		deserData, err := rs.Deserializer.Deserialize(serData)
		if err != nil {
			rs.Logger.Infow("invalid resp data recieved", "error", err)
			continue
		}

		command, ok := deserData.Value.([]*resp.RespDataType)
		if !ok {
			rs.Logger.Infow("failed interface conversion", "type", reflect.TypeOf(deserData.Value))
			os.Exit(1)
		}

		// rs.Logger.Infow("command recieved", "command", command)

		rs.handleCommand(conn, command)

		// tp := reflect.TypeOf(deserData.Value)
		// rs.Logger.Infow("recieved resp data", "type", tp.String())
		// commands := deserData.Value
	}

}

func (rs *RedisServer) Load() error {
	f, err := os.Open(rs.File)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	if err := dec.Decode(&rs.Store); err != nil {
		return err
	}

	return nil

}
