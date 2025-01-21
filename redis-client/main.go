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

// Map to keep track of active connections
var connections = make(map[int]net.Conn)
var mutex = &sync.Mutex{}

type Message struct {
	Message string `json:"message"`
}

// Encode commands to RESP format
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

// Helper function to split a command string into parts
func splitCommand(command string) []string {
	return strings.Fields(command)
}

// TCP connection handler
func connectHandler(w http.ResponseWriter, r *http.Request) {
	portStr := r.URL.Path[len("/connect/"):]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Invalid port", http.StatusBadRequest)
		return
	}

	// Attempt to establish a TCP connection
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to %s: %v", address, err), http.StatusInternalServerError)
		return
	}

	// Store the connection in the map
	mutex.Lock()
	connections[port] = conn
	mutex.Unlock()

	// Return a successful connection message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Successfully connected to port %d!"}`, port)
}

// Message sending handler
func sendHandler(w http.ResponseWriter, r *http.Request) {
	portStr := r.URL.Path[len("/send/"):]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Invalid port", http.StatusBadRequest)
		return
	}

	// Check if connection exists
	mutex.Lock()
	conn, exists := connections[port]
	mutex.Unlock()

	if !exists {
		http.Error(w, fmt.Sprintf("No active connection on port %d", port), http.StatusBadRequest)
		return
	}

	// Decode the incoming message
	var msg Message
	err = json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	// Split commands and encode to RESP format
	commands := strings.Split(msg.Message, "\n")
	respMessage := encodeRESP(commands)

	// Send the encoded message over the TCP connection
	_, err = conn.Write([]byte(respMessage))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Message sent: %s"}`, msg.Message)
}

// Serve HTML page from an external file
func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	// Serve the HTML page at the root URL
	http.HandleFunc("/", serveHome)
	// Handle TCP connection requests
	http.HandleFunc("/connect/", connectHandler)
	// Handle message sending requests
	http.HandleFunc("/send/", sendHandler)

	// Start the server on port 8080
	log.Println("Server running at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
