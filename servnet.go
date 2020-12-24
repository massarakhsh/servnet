package main

import (
	"fmt"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task/baser"
	"os"
	"strings"
	"time"

	"github.com/massarakhsh/lik"
)

func main() {
	lik.SetLevelInf()
	if host,_ := os.Hostname(); strings.ToLower(host) == "shaman" {
		base.HostServ = "192.168.234.62"
	}
	lik.SayError("System started")
	base.HostModes = base.MODE_BASE | base.MODE_PING | base.MODE_ARP | base.MODE_REAL
	//base.HostModes = base.MODE_ARP
	if !getArgs() {
		return
	}
	if !base.OpenDB(base.HostServ, base.HostBase, base.HostUser, base.HostPass) {
		return
	}
	base.WaitDB()
	if (base.HostModes & base.MODE_BASE) != 0 {
		baser.StartBaser()
	}
	if (base.HostModes & base.MODE_PING) != 0 {
		baser.StartPinger()
	}
	if (base.HostModes & base.MODE_ARP) != 0 {
		baser.StartARP()
	}
	for !base.IsStoping {
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second * 3)
}

func getArgs() bool {
	args, ok := lik.GetArgs(os.Args[1:])
	if val := args.GetInt("mode"); val > 0 {
		base.HostModes = val
	}
	if val := args.GetString("serv"); val != "" {
		base.HostServ = val
	}
	if val := args.GetString("base"); val != "" {
		base.HostBase = val
	}
	if val := args.GetString("user"); val != "" {
		base.HostUser = val
	}
	if val := args.GetString("pass"); val != "" {
		base.HostPass = val
	}
	if val := args.GetString("debug"); val != "" {
		base.DebugLevel = lik.StrToInt(val)
	}
	if !ok {
		fmt.Println("Usage: servnet [-key val | --key=val]...")
		fmt.Println("serv    - Database server")
		fmt.Println("base    - Database name")
		fmt.Println("user    - Database user")
		fmt.Println("pass    - Database pass")
	}
	return ok
}
