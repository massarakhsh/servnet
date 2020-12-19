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
	TimeOn  int
	TimeOff int

	SeekOn  time.Time
}

var PingSync sync.Mutex
var PingList []*ElmPing
var PingMapSys map[lik.IDB]*ElmPing
var PingMapIM map[string]*ElmPing

func LoadPing() {
	list := GetList("Ping")
	if list != nil {
		PingSync.Lock()
		PingList = []*ElmPing{}
		PingMapSys = make(map[lik.IDB]*ElmPing)
		PingMapIM = make(map[string]*ElmPing)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sys := elm.GetIDB("SysNum")
				ip := elm.GetString("IP")
				mac := elm.GetString("MAC")
				roles := elm.GetInt("Roles")
				if it := AddPing(sys, ip, mac, roles); it == nil {
					DeleteElm("Ping", sys)
				} else {
					it.TimeOn = elm.GetInt("TimeOn")
					it.TimeOff = elm.GetInt("TimeOff")
					it.SeekOn = time.Unix(int64(it.TimeOn), 0)
				}
			}
		}
		PingSync.Unlock()
	}
}

func SetPingsOffline(ip string) {
	PingSync.Lock()
	for _, it := range PingMapSys {
		if ip == "" || ip == it.IP {
			if (it.Roles & 0x1000) != 0 {
				it.Roles ^= 0x1000
				it.TimeOff = int(time.Now().Unix())
				it.Update()
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
				it.SeekOn = time.Now()
				if (it.Roles & 0x1000) == 0 {
					it.Roles ^= 0x1000
					it.TimeOn = int(time.Now().Unix())
					it.Update()
					AddEvent(it.IP, it.MAC, "", "ON ping")
				}
			} else if ip != "" && (it.Roles&0x1000) != 0 {
				if time.Now().Sub(it.SeekOn) > time.Minute * 2 {
					it.Roles ^= 0x1000
					it.TimeOff = int(time.Now().Unix())
					it.Update()
					AddEvent(it.IP, it.MAC, "", "OFF ping")
				}
			}
		}
	}
	if !found {
		if it := AddPing(0, ip, mac, 0x1000); it != nil {
			it.TimeOn = int(time.Now().Unix())
			it.SeekOn = time.Now()
			it.Update()
			AddEvent(ip, mac, "", "ON new")
		}
	}
	PingSync.Unlock()
}

func AddPing(sys lik.IDB, ip string, mac string, roles int) *ElmPing {
	var it *ElmPing
	if ip != "" {
		AddAsk(ip, (roles&0x1000) != 0)
		if mac != "" {
			if elm,_ := PingMapIM[ip+mac]; elm != nil {
				return nil
			}
			it = &ElmPing{SysNum: sys, IP: ip, MAC: mac, Roles: roles}
			PingList = append(PingList, it)
			if sys > 0 {
				PingMapSys[sys] = it
			}
			PingMapIM[ip+mac] = it
		}
	}
	return it
}

func (it *ElmPing) Update() {
	set := lik.BuildSet()
	set.SetItem(it.Roles, "Roles")
	set.SetItem(it.IP, "IP")
	set.SetItem(it.MAC, "MAC")
	set.SetItem(it.TimeOn, "TimeOn")
	set.SetItem(it.TimeOff, "TimeOff")
	if it.SysNum > 0 {
		UpdateElm("Ping", it.SysNum, set)
	} else {
		it.SysNum = InsertElm("Ping", set)
		PingMapSys[it.SysNum] = it
	}
}
