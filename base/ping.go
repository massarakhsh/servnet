package base

import (
	"github.com/massarakhsh/lik"
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

var PingMapSys map[lik.IDB]*ElmPing
var PingMapIP map[string][]*ElmPing
var PingMapIM map[string]*ElmPing
var PingMapOld map[lik.IDB]*ElmPing

func LoadPing() {
	list := GetList("Ping")
	if list != nil {
		PingMapOld = PingMapSys
		PingMapSys = make(map[lik.IDB]*ElmPing)
		PingMapIP = make(map[string][]*ElmPing)
		PingMapIM = make(map[string]*ElmPing)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sys := elm.GetIDB("SysNum")
				ip := elm.GetString("IP")
				mac := elm.GetString("MAC")
				roles := elm.GetInt("Roles")
				ton := elm.GetInt("TimeOn")
				toff := elm.GetInt("TimeOff")
				tlast := ton
				if toff > ton { tlast = toff }
				if (roles & ROLE_ONLINE) == 0 && time.Now().Sub(time.Unix(int64(tlast),0)) > TimeoutOffline {
					DeleteElm("Ping", sys)
				} else if it := AddPing(sys, ip, mac, roles); it == nil {
					DeleteElm("Ping", sys)
				} else {
					it.TimeOn = ton
					it.TimeOff = toff
					if PingMapOld == nil {
						it.SeekOn = time.Now()
					} else if old := PingMapOld[sys]; old == nil {
						it.SeekOn = time.Now()
					} else {
						it.SeekOn = old.SeekOn
					}
				}
			}
		}
	}
}

func PingSetOffline(ip string) {
	if lip := PingMapIP[ip]; lip != nil {
		for _, it := range lip {
			if (it.Roles & ROLE_ONLINE) != 0 {
				it.Roles ^= ROLE_ONLINE
				it.TimeOff = int(time.Now().Unix())
				it.Update()
				AddEvent(it.IP, it.MAC, "", "OFF ping")
			}
		}
	}
}

func PingSetOnline(ip string, mac string) {
	if ipelm, _ := IPMapIP[ip]; ipelm != nil {
		ipelm.OnlineMAC = mac
		ipelm.SetOnline()
	}
	found := false
	for _, it := range PingMapSys {
		if ip == "" || ip == it.IP {
			if mac == "" || mac == it.MAC {
				found = true
				it.SeekOn = time.Now()
				if (it.Roles & ROLE_ONLINE) == 0 {
					it.Roles ^= ROLE_ONLINE
					it.TimeOn = int(time.Now().Unix())
					it.Update()
					AddEvent(it.IP, it.MAC, "", "ON ping")
				}
			} else if ip != "" && (it.Roles&ROLE_ONLINE) != 0 {
				if time.Now().Sub(it.SeekOn) > TimeoutMAC {
					it.Roles ^= ROLE_ONLINE
					it.TimeOff = int(time.Now().Unix())
					it.Update()
					AddEvent(it.IP, it.MAC, "", "OFF ping")
				}
			}
		}
	}
	if !found {
		if it := AddPing(0, ip, mac, ROLE_ONLINE); it != nil {
			it.TimeOn = int(time.Now().Unix())
			it.SeekOn = time.Now()
			it.Update()
			AddEvent(ip, mac, "", "ON new")
		}
	}
}

func AddPing(sys lik.IDB, ip string, mac string, roles int) *ElmPing {
	var it *ElmPing
	if ip != "" {
		AddAsk(ip, (roles&ROLE_ONLINE) != 0)
		if mac != "" {
			if elm, _ := PingMapIM[ip+mac]; elm != nil {
				return nil
			}
			it = &ElmPing{SysNum: sys, IP: ip, MAC: mac, Roles: roles}
			if ipelm, _ := IPMapIP[ip]; ipelm == nil {
				AddIP(0, ip, mac, roles&ROLE_ONLINE)
			}
			if sys > 0 {
				PingMapSys[sys] = it
			}
			PingMapIM[ip+mac] = it
			if lip := PingMapIP[ip]; lip != nil {
				PingMapIP[ip] = append(PingMapIP[ip], it)
			} else {
				PingMapIP[ip] = []*ElmPing{it}
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
	set.SetItem(it.TimeOn, "TimeOn")
	set.SetItem(it.TimeOff, "TimeOff")
	set.SetItem("CURRENT_TIMESTAMP", "updated_at")
	if it.SysNum > 0 {
		UpdateElm("Ping", it.SysNum, set)
	} else {
		set.SetItem("CURRENT_TIMESTAMP", "created_at")
		it.SysNum = InsertElm("Ping", set)
		PingMapSys[it.SysNum] = it
	}
}
