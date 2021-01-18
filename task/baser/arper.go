package baser

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likssh"
	"github.com/massarakhsh/lik/liktel"
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
		it.callRoot("192.168.234.62")
	}
	it.callRouter("192.168.234.3")
	for _,ip := range []string { "192.168.0.15",
		"192.168.0.241", "192.168.0.242", "192.168.0.243", "192.168.0.244", "192.168.0.245", "192.168.0.246"} {
		it.callSwitch(ip)
	}
	base.Lock()
	for _, arp := range it.Arps {
		base.PingSetOnline(arp.IP, arp.MAC)
	}
	for _, loc := range it.Locs {
		base.TouchOnline(loc.SysUnit, loc.Port, loc.MAC, loc.Secs)
	}
	base.TouchTerminate()
	base.Unlock()
	it.SetPause(time.Second * 15)
}

func (it *ARPer) callLocal() {
	if table := arp.Table(); table != nil {
		if base.DebugLevel > 1 {
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

func (it *ARPer) callRoot(ip string) {
	if touch := likssh.Open(base.IPToShow(ip) + ":22", "root", "", "var/host_rsa"); touch != nil {
		if answer := touch.Execute("arp -an"); answer != "" {
			lines := strings.Split(answer, "\n")
			if base.DebugLevel > 1 {
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

func (it *ARPer) callRouter(ip string) {
	var sysunit lik.IDB
	if ipelm := base.IPMapIP[base.IPFromShow(ip)]; ipelm != nil {
		sysunit = ipelm.SysNum
	}
	if touch := likssh.Open(base.IPToShow(ip) + ":22", "admin", "", "var/host_rsa"); touch != nil {
		if answer := touch.Execute("ip arp print without-paging"); answer != "" {
			lines := strings.Split(answer, "\n")
			if base.DebugLevel > 1 {
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

func (it *ARPer) callSwitch(ip string) {
	var sysunit lik.IDB
	if ipelm := base.IPMapIP[base.IPFromShow(ip)]; ipelm != nil {
		sysunit = ipelm.SysNum
	}
	if touch := liktel.Open(base.IPToShow(ip) + ":23", "cisco", "gamilto17"); touch != nil {
		if _,ok := touch.Execute("terminal datadump"); ok {
			if answer,ok := touch.Execute("show arp"); ok {
				lines := strings.Split(answer, "\n")
				if base.DebugLevel > 1 {
					fmt.Printf("Load switch %s ARP: %d\n", base.IPToShow(ip), len(lines))
				}
				for _, line := range lines {
					if match := lik.RegExParse(line, "(\\d+\\.\\d+\\.\\d+\\.\\d+).+(\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S)"); match != nil {
						ip := base.IPFromShow(match[1])
						mac := base.MACFromShow(match[2])
						it.addArp(ip, mac)
						//fmt.Println(ip, ", ", mac)
					}
				}
			}
			if answer,ok := touch.Execute("show mac addr"); ok {
				lines := strings.Split(answer, "\n")
				if base.DebugLevel > 1 {
					fmt.Printf("Load switch %s MACs: %d\n", base.IPToShow(ip), len(lines))
				}
				for _, line := range lines {
					if match := lik.RegExParse(line, "(\\S+)\\s+(\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S:\\S\\S)\\s+gi(\\d+)\\s+"); match != nil {
						mac := base.MACFromShow(match[2])
						port := lik.StrToInt(match[3])
						it.addArp("", mac)
						it.addLoc(sysunit, port, mac, 0)
						//fmt.Println(mac, ", ", port)
					}
				}
			}
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
	it.Locs = append(it.Locs, LocElm{ SysUnit: sysunit, Port: port, MAC: mac, Secs: secs })
}

