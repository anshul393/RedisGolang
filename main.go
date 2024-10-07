package main

import (
	"flag"
	"os"
	"runtime"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	filename := flag.String("file", "redis.gob", "file for persistence storage")
	flag.Parse()

	server := NewRedisServer("localhost:6666", *filename)
	l, err := server.Listen("tcp", server.Addr)
	if err != nil {
		server.Logger.Infow("failed to get listener", "error", err)
		os.Exit(1)
	}

	server.Logger.Infow("redis server start", "address", server.Addr)

	server.Accept(l)
}
