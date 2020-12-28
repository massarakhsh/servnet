package base

import "time"

var HostServ = "localhost"
var HostBase = "rptp"
var HostUser = "rptp"
var HostPass = "Shaman1961"
var HostModes = 0
var HostName = ""

var DebugLevel = 0
var IsStoping = false

const MODE_BASE = 0x1
const MODE_PING = 0x2
const MODE_ARP = 0x4
const MODE_REAL = 0x10

const ROLE_ONLINE = 0x1000
const ROLE_LINKED = 0x2000
const ROLE_ROOT = 0x10000

var TimeoutIP = time.Second * 180
var TimeoutMAC = time.Second * 60
var TimeoutAlarm = time.Second * 15
var TimeoutFull = time.Second * 900

