package network

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/Yash-sudo-web/redis-implementation-golang/redis-server/internal/commands"
	"github.com/Yash-sudo-web/redis-implementation-golang/redis-server/internal/db"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, err := parseRESPArray(reader)
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading from connection:", err.Error())
			}
			break
		}

		if len(args) == 0 {
			continue
		}

		command := strings.ToUpper(args[0])
		switch command {
		case "PING":
			commands.HandlePing(conn)
		case "ECHO":
			commands.HandleEcho(conn, args)
		case "SET":
			commands.HandleSet(conn, args, db.Db, db.SlaveConnections)
		case "GET":
			commands.HandleGet(conn, args, db.Db)
		case "CONFIG":
			commands.HandleConfig(conn, args, db.Rdbconfig)
		case "KEYS":
			commands.HandleKeys(conn, args, db.Db)
		case "INFO":
			commands.HandleInfo(conn, args, db.DbInfo)
		case "REPLCONF":
			commands.HandleReplConf(conn)
		case "PSYNC":
			commands.HandlePSync(conn, db.DbInfo, db.SlaveConnections, &db.SlavesCount)
		case "WAIT":
			resp := fmt.Sprintf(":%d\r\n", db.SlavesCount)
			conn.Write([]byte(resp))
		case "DELETE":
			commands.HandleDelete(conn, args, db.Db)
		case "QUIT":
			commands.HandleQuit(conn)
		default:
			commands.HandleUnknownCommand(conn)
		}
	}
}
