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
	Arps []ArpElm
	Locs []LocElm
}

type ArpElm struct {
	IP	string
	MAC	string
}

type LocElm struct {
	SysUnit	lik.IDB
	Port	int
	MAC		string
	Secs	int
}

func StartARP() {
	go func() {
		arper := &ARPer{}
		arper.Initialize("ARPer", arper)
	}()
}

func (it *ARPer) DoStep() {
	it.Arps = []ArpElm{}
	it.Locs = []LocElm{}
	if base.HostName == "root" {
		it.callLocal()
	}
	if base.HostName != "root" {
		it.callRoot()
	}
	it.callRouter()
	//it.callSwitch()
	base.Lock()
	for _, arp := range it.Arps {
		base.PingSetOnline(arp.IP, arp.MAC)
	}
	base.Unlock()
	it.SetPause(time.Second * 15)
}

func (it *ARPer) callLocal() {
	if table := arp.Table(); table != nil {
		if base.DebugLevel > 0 {
			fmt.Printf("Load locals ARP: %d\n", len(table))
		}
		for ips, ipa := range table {
			mac := ""
			if match := lik.RegExParse(ipa, "(\\w\\w:\\w\\w:\\w\\w:\\w\\w:\\w\\w:\\w\\w)"); match != nil {
				mac = base.MACFromShow(match[1])
				if mac != "000000000000" && mac != "ffffffffffff" {
					ip := base.IPFromShow(ips)
					it.addArp(ip, mac)
				}
			}
		}
	}
}

func (it *ARPer) callRoot() {
	if touch := likssh.Open("192.168.234.62:22", "root", "", "var/host_rsa"); touch != nil {
		if answer := touch.Execute("arp -an"); answer != "" {
			lines := strings.Split(answer, "\n")
			if base.DebugLevel > 0 {
				fmt.Printf("Load root ARP: %d\n", len(lines))
			}
			for _, line := range lines {
				if match := lik.RegExParse(line, "(\\d+\\.\\d+\\.\\d+\\.\\d+).+(\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S)"); match != nil {
					ip := base.IPFromShow(match[1])
					mac := base.MACFromShow(match[2])
					it.addArp(ip, mac)
				}
			}
		}
	}
}

func (it *ARPer) callRouter() {
	var sysunit lik.IDB
	if ipelm := base.IPMapIP[base.IPFromShow("192.168.0.3")]; ipelm != nil {
		sysunit = ipelm.SysNum
	}
	if touch := likssh.Open("192.168.0.3:22", "admin", "", "var/host_rsa"); touch != nil {
		if answer := touch.Execute("ip arp print without-paging"); answer != "" {
			lines := strings.Split(answer, "\n")
			if base.DebugLevel > 0 {
				fmt.Printf("Load router ARP: %d\n", len(lines))
			}
			for _, line := range lines {
				if match := lik.RegExParse(line, "\\s+(\\w+)\\s+(\\d+\\.\\d+\\.\\d+\\.\\d+).+(\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S)"); match != nil {
					if lik.RegExCompare(match[1], "(c|C)") {
						ip := base.IPFromShow(match[2])
						mac := base.MACFromShow(match[3])
						it.addArp(ip, mac)
					}
				}
			}
		}
		if answer := touch.Execute("interface bridge host print without-paging"); answer != "" {
			lines := strings.Split(answer, "\n")
			for _, line := range lines {
				if match := lik.RegExParse(line, "(\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S)\\s+ether(\\d+)\\s+\\S+\\s+(\\S+)"); match != nil {
					mac := base.MACFromShow(match[1])
					port := lik.StrToInt(match[2])
					dura := match[3]
					secs := 0
					if match := lik.RegExParse(dura, "^(\\d+)s"); match != nil {
						secs = lik.StrToInt(match[1])
					} else if match := lik.RegExParse(dura, "^(\\d+)m(\\d+)s"); match != nil {
						secs = lik.StrToInt(match[1]) * 60 + lik.StrToInt(match[2])
					}
					it.addArp("", mac)
					it.addLoc(sysunit, port, mac, secs)
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

func (it *ARPer) addArp(ip string, mac string) {
	for p := 0; p < len(it.Arps); p++ {
		if ip == "" || it.Arps[p].IP == "" || ip == it.Arps[p].IP {
			if mac == "" || it.Arps[p].MAC == "" || mac == it.Arps[p].MAC {
				if it.Arps[p].IP == "" && ip != "" {
					it.Arps[p].IP = ip
				}
				if it.Arps[p].MAC == "" && mac != "" {
					it.Arps[p].MAC = mac
				}
				return
			}
		}
	}
	it.Arps = append(it.Arps, ArpElm{ip, mac })
}

func (it *ARPer) addLoc(sysunit lik.IDB, port int, mac string, secs int) {
	//it.Locs = append(it.Locs, LocElm{ SysUnit: sysunit, Port: port, MAC: mac, Secs: secs })
}

