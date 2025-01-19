package db

import (
	"net"
)

var DbInfo = make(map[string]interface{})
var Rdbconfig = make(map[string]string)
var Db = make(map[string]map[string]interface{})
var SlavesCount = 0
var SlaveConnections = make(map[string]net.Conn)
var Offset = 0
