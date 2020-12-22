package base

import "time"

var HostServ = "192.168.234.62"
var HostBase = "rptp"
var HostUser = "rptp"
var HostPass = "Shaman1961"
var DebugLevel = 1
var IsStoping = false

var HostModes = 0
const MODE_BASE = 0x1
const MODE_PING = 0x2
const MODE_ARP = 0x4
const MODE_REAL = 0x10

var TimeoutIP = time.Second * 600
var TimeoutMAC = time.Second * 180

