package main

import (
	"fmt"
	"net"
	"os"

	"github.com/Yash-sudo-web/redis-implementation-golang/redis-server/internal/config"
	"github.com/Yash-sudo-web/redis-implementation-golang/redis-server/internal/db"
	"github.com/Yash-sudo-web/redis-implementation-golang/redis-server/internal/network"
)

func main() {
	config.ParseFlags()

	db.LoadRDBFile()

	l, err := net.Listen("tcp", "0.0.0.0:"+config.Port)
	if err != nil {
		fmt.Printf("Failed to bind to port %s: %v\n", config.Port, err)
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Server is running on 0.0.0.0:" + config.Port)

	if db.DbInfo["role"] == "slave" {
		go network.HandleSlaveConn(config.MasterHost, config.MasterPort, config.Port)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go network.HandleConnection(conn)
	}
}
