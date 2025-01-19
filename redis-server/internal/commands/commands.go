package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Yash-sudo-web/redis-server/internal/config"
	"github.com/Yash-sudo-web/redis-server/internal/db"
	"github.com/Yash-sudo-web/redis-server/internal/utils"
)

func HandlePing(conn net.Conn) {
	conn.Write([]byte("+PONG\r\n"))
}

func HandleEcho(conn net.Conn, args []string) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'echo' command\r\n"))
		return
	}
	response := fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])
	conn.Write([]byte(response))
}

func HandleSet(conn net.Conn, args []string, Db map[string]map[string]interface{}, slaveConnections map[string]net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
		return
	}

	key := args[1]
	value := args[2]
	expTime := int64(0)

	if len(args) >= 4 && strings.ToUpper(args[3]) == "PX" {
		if len(args) != 5 {
			conn.Write([]byte("-ERR wrong number of arguments for 'set' with PX option\r\n"))
			return
		}
		pxTime, err := strconv.Atoi(args[4])
		if err != nil || pxTime <= 0 {
			conn.Write([]byte("-ERR invalid expiration time\r\n"))
			return
		}
		expTime = time.Now().UnixMilli() + int64(pxTime)
	}

	Db[key] = map[string]interface{}{
		"value":   value,
		"expTime": expTime,
	}
	remoteAddr := conn.RemoteAddr().String()
	remotePort := strings.Split(remoteAddr, ":")[1]

	if remotePort != config.MasterPort {
		conn.Write([]byte("+OK\r\n"))
	}

	if db.DbInfo["role"] == "master" {
		for slaveID, slaveConn := range slaveConnections {
			fmt.Printf("Propagating to slave %s\n", slaveID)
			command := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
			_, err := slaveConn.Write([]byte(command))
			if err != nil {
				fmt.Printf("Error propagating to slave %s: %v\n", slaveID, err)
				delete(slaveConnections, slaveID)
			}
		}
	}
}

func HandleGet(conn net.Conn, args []string, db map[string]map[string]interface{}) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
		return
	}

	key := args[1]
	response := "$-1\r\n"
	if entry, ok := db[key]; ok {
		if expTime, ok := entry["expTime"].(int64); ok {
			if expTime == 0 || expTime > time.Now().UnixMilli() {
				value := entry["value"].(string)
				response = fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
			} else {
				delete(db, key)
			}
		} else {
			value := entry["value"].(string)
			response = fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
		}
	}

	conn.Write([]byte(response))
}

func HandleConfig(conn net.Conn, args []string, rdbconfig map[string]string) {
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'config' command\r\n"))
		return
	}
	subCommand := strings.ToUpper(args[1])
	if subCommand == "GET" {
		if len(args) != 3 {
			conn.Write([]byte("-ERR wrong number of arguments for 'config get' command\r\n"))
			return
		}

		configKey := args[2]
		response := "$-1\r\n"

		if value, ok := rdbconfig[configKey]; ok {
			response = fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(configKey), configKey, len(value), value)
		}

		conn.Write([]byte(response))
	} else {
		conn.Write([]byte("-ERR unsupported 'config' subcommand\r\n"))
	}
}

func HandleKeys(conn net.Conn, args []string, db map[string]map[string]interface{}) {
	if len(args) != 2 || args[1] != "*" {
		conn.Write([]byte("-ERR wrong arguments for 'keys' command\r\n"))
		return
	}

	response := fmt.Sprintf("*%d\r\n", len(db))
	for key := range db {
		response += fmt.Sprintf("$%d\r\n%s\r\n", len(key), key)
	}
	conn.Write([]byte(response))
}

func HandleInfo(conn net.Conn, args []string, dbinfo map[string]interface{}) {
	if len(args) == 1 {
		return
	}

	subCommand := strings.ToUpper(args[1])
	if subCommand == "REPLICATION" {
		var dblen = 0
		for key := range dbinfo {
			dblen += len(fmt.Sprintf("%v", dbinfo[key])) + len(key) + 3
		}
		dblen -= 2
		response := fmt.Sprintf("$%d\r\n", dblen)
		for key := range dbinfo {
			value := dbinfo[key]
			response += fmt.Sprintf("%s\r\n", key+":"+fmt.Sprintf("%v", value))
		}
		fmt.Printf("response: %s\n", response)
		conn.Write([]byte(response))
	}
}

func HandleReplConf(conn net.Conn) {
	conn.Write([]byte("+OK\r\n"))
}

func HandlePSync(conn net.Conn, dbinfo map[string]interface{}, slaveConnections map[string]net.Conn, slavesCount *int) {
	(*slavesCount)++
	conn.Write([]byte("+FULLRESYNC" + " " + dbinfo["master_replid"].(string) + " " + dbinfo["master_repl_offset"].(string) + "\r\n"))
	utils.SendRDBFile(conn)
	slaveConnections[strconv.Itoa(*slavesCount)] = conn
}

func HandleDelete(conn net.Conn, args []string, Db map[string]map[string]interface{}) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'delete' command\r\n"))
		return
	}
	key := args[1]
	remoteAddr := conn.RemoteAddr().String()
	remotePort := strings.Split(remoteAddr, ":")[1]
	if _, ok := Db[key]; ok {
		delete(Db, key)
		if remotePort != config.MasterPort {
			conn.Write([]byte(":1\r\n"))
		}
	} else {
		if remotePort != config.MasterPort {
			conn.Write([]byte(":0\r\n"))
		}
	}
	if db.DbInfo["role"] == "master" {
		for _, slaveConn := range db.SlaveConnections {
			command := fmt.Sprintf("*2\r\n$6\r\nDELETE\r\n$%d\r\n%s\r\n", len(key), key)
			_, err := slaveConn.Write([]byte(command))
			if err != nil {
				fmt.Printf("Error propagating to slave: %v\n", err)
			}
		}
	}
}

func HandleQuit(conn net.Conn) {
	conn.Write([]byte("+OK\r\n"))
}

func HandleUnknownCommand(conn net.Conn) {
	conn.Write([]byte("-ERR unknown command\r\n"))
}
