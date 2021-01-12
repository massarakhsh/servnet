package base

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"strings"
	"time"
)

type ElmIP struct {
	SysNum lik.IDB
	Roles  int
	IP     string
	MAC    string
	Namely	string
	TimeOn  int
	TimeOff int
	SysUnit	lik.IDB

	OnlineMAC string
	SeekOn    time.Time
	Host	string
}

var IPMapSys map[lik.IDB]*ElmIP
var IPMapIP map[string]*ElmIP
var IPMapOld map[string]*ElmIP

func LoadIP() {
	if list := GetList("IP"); list != nil {
		IPMapOld = IPMapIP
		IPMapSys = make(map[lik.IDB]*ElmIP)
		IPMapIP = make(map[string]*ElmIP)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sys := elm.GetIDB("SysNum")
				if ip := elm.GetString("IP"); ip == "" {
					DeleteElm("IP", sys)
				} else if _, ok := IPMapIP[ip]; ok {
					lik.SayError("IP duplicate " + IPToShow(ip) + " daleted")
					DeleteElm("IP", sys)
				} else {
					it := AddIP(sys, ip, elm.GetString("MAC"), elm.GetInt("Roles"))
					it.Namely = elm.GetString("Namely")
					it.TimeOn = elm.GetInt("TimeOn")
					it.TimeOff = elm.GetInt("TimeOff")
					it.SysUnit = elm.GetIDB("SysUnit")
					if IPMapOld == nil {
						it.SeekOn = time.Now()
					} else if old := IPMapOld[ip]; old == nil {
						it.SeekOn = time.Now()
					} else {
						it.SeekOn = old.SeekOn
					}
					if unit,_ := UnitMapSys[it.SysUnit]; unit != nil {
						unit.ListIP = append(unit.ListIP, sys)
					}
				}
			}
		}
	}
}

func IPToShow(ip string) string {
	ips := ip
	if match := lik.RegExParse(ip, "(\\d\\d\\d)(\\d\\d\\d)(\\d\\d\\d)(\\d\\d\\d)"); match != nil {
		ips = fmt.Sprintf("%d.%d.%d.%d",
			lik.StrToInt(match[1]),
			lik.StrToInt(match[2]),
			lik.StrToInt(match[3]),
			lik.StrToInt(match[4]))
	}
	return ips
}

func IPFromShow(ip string) string {
	ipd := ip
	if match := lik.RegExParse(ip, "(\\d+)\\.(\\d+)\\.(\\d+)\\.(\\d+)"); match != nil {
		ipd = fmt.Sprintf("%03d%03d%03d%03d",
			lik.StrToInt(match[1]),
			lik.StrToInt(match[2]),
			lik.StrToInt(match[3]),
			lik.StrToInt(match[4]))
	}
	return ipd
}

func MACToShow(mac string) string {
	macs := mac
	if match := lik.RegExParse(mac, "(\\w\\w)(\\w\\w)(\\w\\w)(\\w\\w)(\\w\\w)(\\w\\w)"); match != nil {
		macs = fmt.Sprintf("%s:%s:%s:%s:%s:%s", match[1], match[2], match[3], match[4], match[5], match[6])
		macs = strings.ToLower(macs)
	}
	return macs
}

func MACFromShow(mac string) string {
	macd := mac
	if match := lik.RegExParse(mac, "(\\w\\w).(\\w\\w).(\\w\\w).(\\w\\w).(\\w\\w).(\\w\\w)"); match != nil {
		macd = fmt.Sprintf("%s%s%s%s%s%s", match[1], match[2], match[3], match[4], match[5], match[6])
		macd = strings.ToLower(macd)
	}
	return macd
}

func RolesToShow(roles int) string {
	def := ""
	if (roles & ROLE_ONLINE) != 0 {
		def += " On"
	}
	if def == "" {
		def = "Off"
	}
	return def
}

func IPSetOnline(ip string) {
	it, _ := IPMapIP[ip]
	if it == nil && ip != "" {
		it = AddIP(0, ip, "", ROLE_ONLINE)
	}
	if it != nil {
		it.SetOnline()
	}
}

func IPSetOffline(ip string) {
	it, _ := IPMapIP[ip]
	if it == nil && ip != "" {
		it = AddIP(0, ip, "", 0)
	}
	if it != nil {
		it.SetOffline()
	}
}

func (it *ElmIP) SetOnline() {
	it.SeekOn = time.Now()
	if (it.Roles & ROLE_ONLINE) == 0 {
		it.Roles ^= ROLE_ONLINE
		it.TimeOn = int(time.Now().Unix())
		it.Update()
		AddEvent(it.IP, it.OnlineMAC, "", "ON ip")
	}
}

func (it *ElmIP) SetOffline() {
	if (it.Roles & ROLE_ONLINE) != 0 {
		if time.Now().Sub(it.SeekOn) > TimeoutIP {
			it.Roles ^= ROLE_ONLINE
			it.OnlineMAC = ""
			it.TimeOff = int(time.Now().Unix())
			it.Update()
			PingSetOffline(it.IP)
			AddEvent(it.IP, it.OnlineMAC, "", "OFF ip")
		}
	} else {
		PingSetOffline(it.IP)
	}
}

func (it *ElmIP) Update() {
	if it.SysNum > 0 {
		set := lik.BuildSet()
		set.SetItem(it.Roles, "Roles")
		set.SetItem(it.IP, "IP")
		set.SetItem(it.MAC, "MAC")
		set.SetItem(it.TimeOn, "TimeOn")
		set.SetItem(it.TimeOff, "TimeOff")
		set.SetItem("CURRENT_TIMESTAMP", "updated_at")
		UpdateElm("IP", it.SysNum, set)
	}
	if unit,_ := UnitMapSys[it.SysUnit]; unit != nil {
		unit.NetUpdate()
	}
}

func AddIP(sys lik.IDB, ip string, mac string, roles int) *ElmIP {
	AddAsk(ip, (roles&ROLE_ONLINE) != 0)
	it := &ElmIP{SysNum: sys, IP: ip, MAC: mac, Roles: roles }
	IPMapSys[sys] = it
	IPMapIP[ip] = it
	return it
}

