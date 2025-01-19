package network

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/Yash-sudo-web/redis-server/internal/commands"
	"github.com/Yash-sudo-web/redis-server/internal/db"
)

func HandleSlaveConn(MasterHost string, MasterPort string, port string) {
	if MasterHost == "localhost" {
		MasterHost = "0.0.0.0"
	}

	conn, err := net.Dial("tcp", MasterHost+":"+MasterPort)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	fmt.Println("Connected to master at", MasterHost+":"+MasterPort)

	reader := bufio.NewReader(conn)

	// PING
	message := fmt.Sprintf("*1\r\n$%d\r\n%s\r\n", 4, "PING")
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error writing PING:", err)
		conn.Close()
		return
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading PING response:", err)
		conn.Close()
		return
	}
	if line != "+PONG\r\n" {
		fmt.Println("Error: Expected PONG, got", line)
		conn.Close()
		return
	}

	// REPLCONF listening-port
	command := "REPLCONF"
	parameter := "listening-port"
	request := fmt.Sprintf("*3\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
		len(command), command,
		len(parameter), parameter,
		len(port), port)
	_, err = conn.Write([]byte(request))
	if err != nil {
		fmt.Println("Error writing REPLCONF listening-port:", err)
		conn.Close()
		return
	}

	line, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading REPLCONF response:", err)
		conn.Close()
		return
	}
	if line != "+OK\r\n" {
		fmt.Println("Error: Expected OK, got", line)
		conn.Close()
		return
	}

	// REPLCONF capa
	parameter = "capa"
	replication_capability := "psync2"
	request = fmt.Sprintf("*3\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
		len(command), command,
		len(parameter), parameter,
		len(replication_capability), replication_capability)
	_, err = conn.Write([]byte(request))
	if err != nil {
		fmt.Println("Error writing REPLCONF capa:", err)
		conn.Close()
		return
	}

	line, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading REPLCONF capa response:", err)
		conn.Close()
		return
	}
	if line != "+OK\r\n" {
		fmt.Println("Error: Expected OK, got", line)
		conn.Close()
		return
	}

	request = fmt.Sprintf("*3\r\n$5\r\nPSYNC\r\n$1\r\n%s\r\n$2\r\n%s\r\n", "?", "-1")
	_, err = conn.Write([]byte(request))
	if err != nil {
		fmt.Println("Error writing PSYNC:", err)
		conn.Close()
		return
	}

	line, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading PSYNC response:", err)
		conn.Close()
		return
	}
	if !strings.HasPrefix(line, "+FULLRESYNC") {
		fmt.Printf("Unexpected PSYNC response: %s\n", line)
		conn.Close()
		return
	}

	line, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading RDB size:", err)
		conn.Close()
		return
	}

	if !strings.HasPrefix(line, "$") {
		fmt.Printf("Unexpected RDB size format: %s\n", line)
		conn.Close()
		return
	}

	rdbSize, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		fmt.Printf("Error parsing RDB size: %v\n", err)
		conn.Close()
		return
	}

	rdbData := make([]byte, rdbSize)
	_, err = io.ReadFull(reader, rdbData)
	if err != nil {
		fmt.Printf("Error reading RDB data: %v\n", err)
		conn.Close()
		return
	}

	fmt.Println("Successfully received RDB file")

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
			db.DbInfo["slave_repl_offset"] = db.DbInfo["slave_repl_offset"].(int) + offset
		case "ECHO":
			commands.HandleEcho(conn, args)
			db.DbInfo["slave_repl_offset"] = db.DbInfo["slave_repl_offset"].(int) + offset
		case "SET":
			commands.HandleSet(conn, args, db.Db, db.SlaveConnections)
			db.DbInfo["slave_repl_offset"] = db.DbInfo["slave_repl_offset"].(int) + offset
		case "GET":
			commands.HandleGet(conn, args, db.Db)
			db.DbInfo["slave_repl_offset"] = db.DbInfo["slave_repl_offset"].(int) + offset
		case "CONFIG":
			commands.HandleConfig(conn, args, db.Rdbconfig)
			db.DbInfo["slave_repl_offset"] = db.DbInfo["slave_repl_offset"].(int) + offset
		case "KEYS":
			commands.HandleKeys(conn, args, db.Db)
			db.DbInfo["slave_repl_offset"] = db.DbInfo["slave_repl_offset"].(int) + offset
		case "INFO":
			commands.HandleInfo(conn, args, db.DbInfo)
			db.DbInfo["slave_repl_offset"] = db.DbInfo["slave_repl_offset"].(int) + offset
		case "DELETE":
			commands.HandleDelete(conn, args, db.Db)
			db.DbInfo["slave_repl_offset"] = db.DbInfo["slave_repl_offset"].(int) + offset
		case "REPLCONF":
			subCommand := strings.ToUpper(args[1])
			if subCommand == "GETACK" {
				resp := fmt.Sprintf("*3\r\n$8\r\n%s\r\n$3\r\n%s\r\n$%d\r\n%d\r\n", "REPLCONF", "ACK", len(strconv.Itoa(db.DbInfo["slave_repl_offset"].(int))), db.DbInfo["slave_repl_offset"])
				conn.Write([]byte(resp))
				db.DbInfo["slave_repl_offset"] = db.DbInfo["slave_repl_offset"].(int) + offset
			} else {
				conn.Write([]byte("-ERR unsupported 'REPLCONF' subcommand\r\n"))
			}
		}
	}

	fmt.Println("Replication connection terminated")
	conn.Close()
}
