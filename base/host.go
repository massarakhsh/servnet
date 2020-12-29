package base

import (
	"os"
	"time"
)

var HostServ = "localhost"
var HostBase = "rptp"
var HostUser = "rptp"
var HostPass = "Shaman1961"
var HostName = ""
var HostVirtual = false

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

