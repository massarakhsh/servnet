package baser

import (
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"github.com/mostlygeek/arp"
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
