package base

import (
	"os"
	"time"
)

var ConfServ = "localhost"
var ConfBase = "rptp"
var ConfUser = "rptp"
var ConfPass = "Shaman1961"
var ConfVirtual = false
var ConfPort = 0

var HostName = ""
var HostPid = 0
var HostSignal = ""
var HostChan chan os.Signal

var DebugLevel = 0
var IsStoping = false

const ROLE_ONLINE = 0x1000
const ROLE_LINKED = 0x2000
const ROLE_ROOT = 0x10000

var TimeoutIP = time.Second * 60
var TimeoutMAC = time.Second * 50
var TimeoutAlarm = time.Second * 5
var TimeoutFull = time.Second * 600
var TimeoutOffline = time.Hour * 24 * 30

