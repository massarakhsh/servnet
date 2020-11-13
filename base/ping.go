package base

import (
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/ruler"
	"sync"
	"time"
)

type ElmPing struct {
	SysNum  lik.IDB
	Roles   int
	IP      string
	MAC     string
	Namely  string
	TimeOn  int
	TimeOff int
}

type ElmIM struct {
	Pings []*ElmPing
}

var PingSync sync.Mutex
var PingList []*ElmPing
var PingMapSys map[lik.IDB]*ElmPing
var PingMapIM map[string]*ElmIM

func LoadPing() {
	list := GetList("Ping")
	if list != nil {
		PingSync.Lock()
		PingList = []*ElmPing{}
		PingMapSys = make(map[lik.IDB]*ElmPing)
		PingMapIM = make(map[string]*ElmIM)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sys := elm.GetIDB("SysNum")
				ip := elm.GetString("IP")
				mac := elm.GetString("MAC")
				namely := elm.GetString("Namely")
				roles := elm.GetInt("Roles")
				if it := AddPing(sys, ip, mac, namely, roles); it == nil {
					DeleteElm("Ping", sys)
				} else {
					it.TimeOn = elm.GetInt("TimeOn")
					it.TimeOff = elm.GetInt("TimeOff")
				}
			}
		}
		PingSync.Unlock()
	}
}

func SetPingOffline(ip string) {
	PingSync.Lock()
	for _, elm := range PingMapSys {
		if ip == elm.IP {
			if (elm.Roles & 0x1000) != 0 {
				elm.Roles ^= 0x1000
				elm.TimeOff = int(time.Now().Unix())
				UpdatePing(elm)
			}
		}
	}
	PingSync.Unlock()
}

func SetPingOnline(ip string, mac string) {
	PingSync.Lock()
	if ipelm, _ := IPMapIP[ip]; ipelm != nil {
		ipelm.OnlineMAC = mac
		ipelm.SetIPOnline()
	}
	found := false
	for _, it := range PingMapSys {
		if ip == "" || ip == it.IP {
			if mac == "" || mac == it.MAC {
				found = true
				if (it.Roles & 0x1000) == 0 {
					it.Roles ^= 0x1000
					it.TimeOn = int(time.Now().Unix())
					UpdatePing(it)
					if ruler.DebugLevel > 0 {
						lik.SayInfo("Online " + IPToShow(ip) + "(" + MACToShow(mac) + ")")
					}
				}
			} else if ip != "" && (it.Roles&0x1000) != 0 {
				it.Roles ^= 0x1000
				it.TimeOff = int(time.Now().Unix())
				UpdatePing(it)
				if ruler.DebugLevel > 0 {
					lik.SayInfo("OFF " + IPToShow(ip) + "(" + MACToShow(mac) + ")")
				}
			}
		}
	}
	if !found {
		if it := AddPing(0, ip, mac, "", 0x1000); it != nil {
			it.TimeOn = int(time.Now().Unix())
			UpdatePing(it)
			if ruler.DebugLevel > 0 {
				lik.SayInfo("New online " + IPToShow(ip) + "(" + MACToShow(mac) + ")")
			}
		}
	}
	PingSync.Unlock()
}

func AddPing(sys lik.IDB, ip string, mac string, name string, roles int) *ElmPing {
	var it *ElmPing
	if ip != "" {
		AddAsk(ip, (roles&0x1000) != 0)
		if mac != "" {
			found := false
			im := ip + mac
			itim, _ := PingMapIM[im]
			if itim != nil {
				for _, ep := range itim.Pings {
					if name == "" || ep.Namely == "" || name == ep.Namely {
						found = true
						break
					}
				}
			}
			if !found {
				it = &ElmPing{SysNum: sys, IP: ip, MAC: mac, Namely: name, Roles: roles}
				PingList = append(PingList, it)
				if sys > 0 {
					PingMapSys[sys] = it
				}
				if itim == nil {
					itim = &ElmIM{}
				}
				itim.Pings = append(itim.Pings, it)
			}
		}
	}
	return it
}

func UpdatePing(elm *ElmPing) {
	set := lik.BuildSet()
	set.SetItem(elm.Roles, "Roles")
	set.SetItem(elm.IP, "IP")
	set.SetItem(elm.MAC, "MAC")
	set.SetItem(elm.Namely, "Namely")
	set.SetItem(elm.TimeOn, "TimeOn")
	set.SetItem(elm.TimeOff, "TimeOff")
	if elm.SysNum > 0 {
		UpdateElm("Ping", elm.SysNum, set)
	} else {
		elm.SysNum = InsertElm("Ping", set)
		PingMapSys[elm.SysNum] = elm
	}
}
