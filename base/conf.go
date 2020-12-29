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

var HostName = ""
var HostPid = 0
var HostSignal = ""
var HostChan chan os.Signal

var DebugLevel = 0
var IsStoping = false

const ROLE_ONLINE = 0x1000
const ROLE_LINKED = 0x2000
const ROLE_ROOT = 0x10000

var TimeoutIP = time.Second * 180
var TimeoutMAC = time.Second * 60
var TimeoutAlarm = time.Second * 15
var TimeoutFull = time.Second * 900

