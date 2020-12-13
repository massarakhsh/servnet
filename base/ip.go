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
	TimeOn  int
	TimeOff int

	OnlineMAC string
	SeekOn    time.Time
}

var IPMapSys map[lik.IDB]*ElmIP
var IPMapIP map[string]*ElmIP

func LoadIP() {
	if list := GetList("IP"); list != nil {
		IPMapSys = make(map[lik.IDB]*ElmIP)
		IPMapIP = make(map[string]*ElmIP)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sysnum := elm.GetIDB("SysNum")
				if ip := elm.GetString("IP"); ip == "" {
					DeleteElm("IP", sysnum)
				} else if _, ok := IPMapIP[ip]; ok {
					lik.SayError("IP duplicate " + IPToShow(ip) + " daleted")
					DeleteElm("IP", sysnum)
				} else {
					it := AddIP(sysnum, ip, elm.GetString("MAC"))
					it.Roles = elm.GetInt("Roles")
					it.TimeOn = elm.GetInt("TimeOn")
					it.TimeOff = elm.GetInt("TimeOff")
					it.SeekOn = time.Now()
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
	if (roles & 0x1000) != 0 {
		def += " On"
	}
	if def == "" {
		def = "Off"
	}
	return def
}

func SetIPOnline(ip string) {
	if it, _ := IPMapIP[ip]; it != nil {
		it.SetIPOnline()
	}
}

func (it *ElmIP) SetIPOnline() {
	it.SeekOn = time.Now()
	if (it.Roles & 0x1000) == 0 {
		it.Roles ^= 0x1000
		it.TimeOn = int(time.Now().Unix())
		UpdateIP(it)
		AddEvent(it.TimeOn, it.IP, it.OnlineMAC, "", "On")
	}
}

func SetIPOffline(ip string) {
	if it, _ := IPMapIP[ip]; it != nil {
		it.SetIPOffline()
	}
}

func (it *ElmIP) SetIPOffline() {
	if (it.Roles & 0x1000) != 0 {
		if time.Now().Sub(it.SeekOn) > time.Minute*1 {
			it.Roles ^= 0x1000
			it.OnlineMAC = ""
			it.TimeOff = int(time.Now().Unix())
			UpdateIP(it)
			SetPingOffline(it.IP)
			AddEvent(it.TimeOn, it.IP, it.OnlineMAC, "", "Off")
		}
	}
}

func AddIP(sys lik.IDB, ip string, mac string) *ElmIP {
	it := &ElmIP{SysNum: sys, IP: ip, MAC: mac}
	IPMapSys[sys] = it
	IPMapIP[ip] = it
	return it
}

func UpdateIP(elm *ElmIP) {
	set := lik.BuildSet()
	set.SetItem(elm.Roles, "Roles")
	set.SetItem(elm.IP, "IP")
	set.SetItem(elm.MAC, "MAC")
	set.SetItem(elm.TimeOn, "TimeOn")
	set.SetItem(elm.TimeOff, "TimeOff")
	if elm.SysNum > 0 {
		UpdateElm("IP", elm.SysNum, set)
	}
}

func AddEvent(at int, ip string, mac string, namely string, formula string) {
	set := lik.BuildSet()
	set.SetItem(at, "TimeAt")
	set.SetItem(ip, "IP")
	set.SetItem(mac, "MAC")
	set.SetItem(namely, "Namely")
	set.SetItem(formula, "Formula")
	InsertElm("Eventage", set)
	old := int(time.Now().Add(time.Hour * 24 * 30).Unix())
	DB.Execute(fmt.Sprintf("DELETE FROM Eventage WHERE TimeAt<%d", old))
	if DebugLevel > 0 {
		lik.SayInfo(fmt.Sprintf("IP %s: %s", IPToShow(ip), formula))
	}
}
