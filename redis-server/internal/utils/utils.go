package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
)

func GenerateRandomHexString(length int) string {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func SendRDBFile(conn net.Conn) {

	hexFile := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

	file, err := hex.DecodeString(hexFile)
	if err != nil {
		fmt.Println("Error decoding hex string:", err)
		return
	}

	rdbContent := fmt.Sprintf("$%d\r\n%s", len(file), file)

	_, err = conn.Write([]byte(rdbContent))
	if err != nil {
		fmt.Println("Error sending data:", err)
	}
}
