package baser

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likssh"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"github.com/mostlygeek/arp"
	"strings"
	"time"
)

type ARPer struct {
	task.Task
	Elms	[]ArpElm
}

type ArpElm struct {
	IP	string
	MAC	string
}

func StartARP() {
	go func() {
		arper := &ARPer{}
		arper.Initialize("ARPer", arper)
	}()
}

func (it *ARPer) DoStep() {
	it.Elms = []ArpElm{}
	if base.HostName == "root" {
		it.callLocal()
	}
	if base.HostName != "root" {
		it.callRoot()
	}
	it.callRouter()
	//it.callSwitch()
	base.LockDB()
	for _,elm := range it.Elms {
		//fmt.Printf("%s : %s\n", base.IPToShow(elm.IP), base.MACToShow(elm.MAC))
		base.SetPingOnline(elm.IP, elm.MAC)
	}
	base.UnlockDB()
	it.SetPause(time.Second * 15)
}

func (it *ARPer) callLocal() {
	if table := arp.Table(); table != nil {
		for ips, ipa := range table {
			mac := ""
			if match := lik.RegExParse(ipa, "(\\w\\w:\\w\\w:\\w\\w:\\w\\w:\\w\\w:\\w\\w)"); match != nil {
				mac = base.MACFromShow(match[1])
				if mac != "000000000000" && mac != "ffffffffffff" {
					ip := base.IPFromShow(ips)
					it.addElm(ip, mac)
				}
			}
		}
	}
}

func (it *ARPer) callRoot() {
	if touch := likssh.Open("192.168.234.62:22", "root", "", "root.opn"); touch != nil {
		if answer := touch.Execute("arp -an"); answer != "" {
			lines := strings.Split(answer, "\n")
			for _, line := range lines {
				if match := lik.RegExParse(line, "(\\d+\\.\\d+\\.\\d+\\.\\d+).+(\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S)"); match != nil {
					ip := base.IPFromShow(match[1])
					mac := base.MACFromShow(match[2])
					it.addElm(ip, mac)
				}
			}
		}
	}
}

func (it *ARPer) callRouter() {
	if touch := likssh.Open("192.168.0.3:22", "admin", "", "root.opn"); touch != nil {
		if answer := touch.Execute("ip arp print without-paging"); answer != "" {
			lines := strings.Split(answer, "\n")
			for _, line := range lines {
				if match := lik.RegExParse(line, "\\s+(\\w+)\\s+(\\d+\\.\\d+\\.\\d+\\.\\d+).+(\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S)"); match != nil {
					if lik.RegExCompare(match[1], "(c|C)") {
						ip := base.IPFromShow(match[2])
						mac := base.MACFromShow(match[3])
						it.addElm(ip, mac)
					}
				}
			}
		}
		if answer := touch.Execute("interface bridge host print without-paging"); answer != "" {
			lines := strings.Split(answer, "\n")
			for _, line := range lines {
				if match := lik.RegExParse(line, "(\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S)"); match != nil {
					mac := base.MACFromShow(match[1])
					it.addElm("", mac)
				}
			}
		}
		touch.Close()
	}
}

func (it *ARPer) callSwitch() {
	if touch := likssh.Open("192.168.0.241:22", "cisco", "gamilto17", ""); touch != nil {
		if answer := touch.Execute("dir"); answer != "" {
			fmt.Println(answer)
		}
		touch.Close()
	}
}

func (it *ARPer) addElm(ip string, mac string) {
	for p := 0; p < len(it.Elms); p++ {
		if ip == "" || it.Elms[p].IP == "" || ip == it.Elms[p].IP {
			if mac == "" || it.Elms[p].MAC == "" || mac == it.Elms[p].MAC {
				if it.Elms[p].IP == "" && ip != "" {
					it.Elms[p].IP = ip
				}
				if it.Elms[p].MAC == "" && mac != "" {
					it.Elms[p].MAC = mac
				}
				return
			}
		}
	}
	it.Elms = append(it.Elms, ArpElm{ ip, mac })
}

