package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var connections = make(map[int]net.Conn)
var mutex = &sync.Mutex{}

type Message struct {
	Message string `json:"message"`
}

func encodeRESP(commands []string) string {
	var resp string
	for _, command := range commands {
		parts := splitCommand(command)
		resp += fmt.Sprintf("*%d\r\n", len(parts))
		for _, part := range parts {
			resp += fmt.Sprintf("$%d\r\n%s\r\n", len(part), part)
		}
	}
	return resp
}

func splitCommand(command string) []string {
	return strings.Fields(command)
}

func connectHandler(w http.ResponseWriter, r *http.Request) {
	portStr := r.URL.Path[len("/connect/"):]

	host := strings.Split(portStr, ":")[0]
	port, err := strconv.Atoi(strings.Split(portStr, ":")[1])
	if err != nil {
		http.Error(w, "Invalid port", http.StatusBadRequest)
		return
	}
	if host == "localhost" {
		host = "127.0.0.1"
	}

	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to %s: %v", address, err), http.StatusInternalServerError)
		return
	}

	mutex.Lock()
	connections[port] = conn
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Successfully connected to URL %d!"}`, port)
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	portStr := r.URL.Path[len("/send/"):]
	host := strings.Split(portStr, ":")[0]
	port, err := strconv.Atoi(strings.Split(portStr, ":")[1])
	if err != nil {
		http.Error(w, "Invalid port", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	conn, exists := connections[host:port]
	mutex.Unlock()

	if !exists {
		http.Error(w, fmt.Sprintf("No active connection on port %d", port), http.StatusBadRequest)
		return
	}

	var msg Message
	err = json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	commands := strings.Split(msg.Message, "\n")
	respMessage := encodeRESP(commands)

	_, err = conn.Write([]byte(respMessage))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(buf[:n])
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {

	http.HandleFunc("/", serveHome)

	http.HandleFunc("/connect/", connectHandler)

	http.HandleFunc("/send/", sendHandler)

	log.Println("Server running at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
