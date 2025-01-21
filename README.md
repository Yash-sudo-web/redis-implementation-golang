# Redis Implementation in Go

A lightweight Redis server implementation written in Go featuring support for various Redis commands, RESP protocol handling, and basic master-slave replication. 

---

## Table of Contents
- [Features](#features)
- [Architecture Overview](#architecture-overview)
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

---

## Architecture/Folder Structure Overview
This implementation is organized into several key components:

1. **Database (`db`)**:
   - Handles in-memory key-value storage.
   - Implements RDB file loading for persistence.

2. **Network Layer (`network`)**:
   - Handles client-server communication.
   - Implements RESP parsing and replication protocol.

3. **Commands (`commands`)**:
   - Implements various Redis commands like `PING`, `SET`, and `GET`.

4. **Utilities (`utils`)**:
   - Includes helper functions, such as random hex string generation and RDB file transmission.

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
3. Build the project:
   
   ```bash
   go build ./cmd/main.go
4. Run the project
   
   ```bash
   # Standalone mode
    ./main --port 6379 --dir ./test-rdb --dbfilename dump.rdb
    
    # Slave mode
    ./main --port 6380 --dir ./test-rdb --dbfilename dump.rdb --replicaof "localhost 6379"

     #RDB files can be loaded into the server, with a sample file provided in the test-rdb directory for reference.
### Docker Deployment
1. Run with Docker Compose (master-slave setup):
   This will deploy and expose 1 master and 2 slave instances of the server on port 6379, 6380 and 6381 respectively.
   
   ```bash
   # Build docker images
   docker-compose build
   # Start all services
   docker-compose up
## Usage
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
