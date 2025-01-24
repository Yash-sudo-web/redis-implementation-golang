# Redis Implementation in Go

A lightweight Redis server and a Redis Client implementation written in Go featuring support for various Redis commands, RESP protocol handling, loading RDB files and basic master-slave replication. 

---

## Table of Contents
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
  - [Command Line Arguments](#command-line-arguments)
  - [Supported Commands](#supported-command)
- [Contributing](#contributing)
- [License](#license)

---

## Features
- Basic Redis commands (GET, SET, DEL, PING, ECHO)
- Key expiration support
- Master-slave replication
- RDB file persistence
- RESP (Redis Serialization Protocol) implementation
- Docker support with master-slave configuration
- Compatibility with Redis: The custom server and client are fully compatible with Redis's official server and client. This means the custom server can interact with the Redis client, and the custom client can work with the Redis server.

---

## Getting Started

### Prerequisites
- Go version 1.21 or higher.
- Basic understanding of Redis and its protocol.
- Docker (optional)
- Docker Compose (optional)

## Installation

### Local Installation
1. Clone this repository:
   
   ```bash
   git clone https://github.com/Yash-sudo-web/redis-implementation-golang.git
   cd redis-implementation-golang
   cd redis-server
3. Build the Server:
   
   ```bash
   go build ./cmd/main.go
4. Run the Server
   
   ```bash
   # Standalone mode
    ./main --port 6379 --dir ./test-rdb --dbfilename dump.rdb
    
    # Slave mode
    ./main --port 6380 --dir ./test-rdb --dbfilename dump.rdb --replicaof "localhost 6379"

     #RDB files can be loaded into the server, with a sample file provided in the test-rdb directory for reference.
5. Bulid the Client
   ```bash
   cd .. && cd redis-client
   go build .

6. Run the Client
   ```bash
   ./redis-client

### Docker Deployment
1. Run with Docker Compose (master-slave setup):
   This will deploy and expose 1 master and 2 slave instances of the server and 1 Client on port 6379, 6380, 6381 and 8080 respectively.
   
   ```bash
   # Build docker images
   docker-compose build
   # Start all services
   docker-compose up
## Usage
### Connection Details
When using the custom client, specify the hostnames and ports as redis-master:6379 for the master, redis-slave-1:6379 for slave 1, and redis-slave-2:6379 for slave 2. These settings can be adjusted in the docker-compose file if needed. However, when using Redis's official CLI, specifying these hostnames is not required and only ports would be required to establish the connection.

### Command Line Arguments
- --port: Server port (default: 6379)
- --dir: Directory for RDB file storage
- --dbfilename: Name of the RDB file
- --replicaof: Master node configuration for slave mode (format: "host port")

### Supported Commands
|      Command     |      Description     |              Usage              |
|:----------------:|:--------------------:|:-------------------------------:|
| PING             | Test connection      | PING                            |
| ECHO             | Echo input           | ECHO message                    |
| SET              | Set key-value        | SET key value [PX milliseconds] |
| GET              | Get value            | GET key                         |
| DEL              | Delete key           | DEL key                         |
| KEYS             | List all keys        | KEYS *                          |
| CONFIG GET       | Get configuration    | CONFIG GET parameter            |
| INFO REPLICATION | Get replication info | INFO REPLICATION                |

## Contributing
Contributions are welcome! Feel free to open issues or submit pull requests.

## License
This project is licensed under the [MIT License](https://github.com/Yash-sudo-web/redis-implementation-golang/blob/main/LICENSE).
