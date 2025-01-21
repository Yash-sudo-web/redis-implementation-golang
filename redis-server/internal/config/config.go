package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/Yash-sudo-web/redis-implementation-golang/redis-server/internal/db"
	"github.com/Yash-sudo-web/redis-implementation-golang/redis-server/internal/utils"
)

var Port = "6379"
var MasterHost = ""
var MasterPort = ""

func ParseFlags() {
	db.DbInfo["role"] = "master"
	db.DbInfo["master_replid"] = utils.GenerateRandomHexString(40)
	db.DbInfo["master_repl_offset"] = "0"

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--port":
			if i+1 < len(os.Args) {
				Port = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: Missing value for --port")
				os.Exit(1)
			}
		case "--dir":
			if i+1 < len(os.Args) {
				db.Rdbconfig["dir"] = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: Missing value for --dir")
				os.Exit(1)
			}
		case "--dbfilename":
			if i+1 < len(os.Args) {
				db.Rdbconfig["dbfilename"] = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: Missing value for --dbfilename")
				os.Exit(1)
			}
		case "--replicaof":
			if i+1 < len(os.Args) {
				db.DbInfo["role"] = "slave"
				db.DbInfo["slave_replid"] = utils.GenerateRandomHexString(40)
				db.DbInfo["slave_repl_offset"] = 0
				masterconf := os.Args[i+1]
				MasterHost = strings.Split(masterconf, " ")[0]
				MasterPort = strings.Split(masterconf, " ")[1]
				i++
			} else {
				fmt.Println("Error: Missing value for --replicaof")
				os.Exit(1)
			}
		default:
			fmt.Printf("Warning: Unrecognized argument %s\n", os.Args[i])
		}
	}
}
