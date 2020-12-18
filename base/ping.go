package base

import (
	"github.com/massarakhsh/lik"
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
	TimeOnline	time.Time
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
				elm.Update()
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
				it.TimeOnline = time.Now()
				if (it.Roles & 0x1000) == 0 {
					it.Roles ^= 0x1000
					it.TimeOn = int(time.Now().Unix())
					it.Update()
					AddEvent(it.IP, it.MAC, it.Namely, "ON ping")
				}
			} else if ip != "" && (it.Roles&0x1000) != 0 {
				if time.Now().Sub(it.TimeOnline) > time.Second * 120 {
					it.Roles ^= 0x1000
					it.TimeOff = int(time.Now().Unix())
					it.Update()
					AddEvent(it.IP, it.MAC, it.Namely, "OFF ping")
				}
			}
		}
	}
	if !found {
		if it := AddPing(0, ip, mac, "", 0x1000); it != nil {
			it.TimeOn = int(time.Now().Unix())
			it.Update()
			AddEvent(ip, mac, "", "ON new")
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

func (it *ElmPing) Update() {
	set := lik.BuildSet()
	set.SetItem(it.Roles, "Roles")
	set.SetItem(it.IP, "IP")
	set.SetItem(it.MAC, "MAC")
	set.SetItem(it.Namely, "Namely")
	set.SetItem(it.TimeOn, "TimeOn")
	set.SetItem(it.TimeOff, "TimeOff")
	if it.SysNum > 0 {
		UpdateElm("Ping", it.SysNum, set)
	} else {
		it.SysNum = InsertElm("Ping", set)
		PingMapSys[it.SysNum] = it
	}
}
