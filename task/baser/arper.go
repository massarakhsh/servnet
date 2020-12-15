package baser

import (
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"github.com/mostlygeek/arp"
	"github.com/reiver/go-telnet"
	"time"
)

type ARPer struct {
	task.Task
}

func StartARP() {
	go func() {
		arper := &ARPer{}
		arper.Initialize("ARPer", arper)
	}()
}

func (it *ARPer) DoStep() {
	CallRouter()
	if table := arp.Table(); table != nil {
		base.LockDB()
		for ip, ipa := range table {
			mac := ""
			if match := lik.RegExParse(ipa, "(\\w\\w:\\w\\w:\\w\\w:\\w\\w:\\w\\w:\\w\\w)"); match != nil {
				mac = base.MACFromShow(match[1])
				if mac != "000000000000" && mac != "ffffffffffff" {
					//fmt.Printf("%s : %s\n", ip, base.MACToShow(mac))
					base.SetPingOnline(base.IPFromShow(ip), mac)
				}
			}
		}
		base.UnlockDB()
	}
	it.SetPause(time.Second * 15)
}

func CallRouter() {
	lik.SayInfo("Telnet")
	var caller telnet.Caller = telnet.StandardCaller
	telnet.DialToAndCall("192.168.0.3:23", caller)
	lik.SayInfo("Ok")
}

