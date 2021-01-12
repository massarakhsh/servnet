package base

import (
	"github.com/massarakhsh/lik"
	"time"
)

type ElmTouch struct {
	SysNum lik.IDB
	SysUnit lik.IDB
	Roles  	int
	Port	int
	MAC    	string
	TimeOn  int
	TimeAt	int
}

var TouchMapSys map[lik.IDB]*ElmTouch
var TouchMapIP map[string]*ElmTouch
var TouchMapOld map[lik.IDB]*ElmTouch

func LoadTouch() {
	if list := GetList("Touch"); list != nil {
		TouchMapOld = TouchMapSys
		TouchMapSys = make(map[lik.IDB]*ElmTouch)
		TouchMapIP = make(map[string]*ElmTouch)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sys := elm.GetIDB("SysNum")
				roles := elm.GetInt("Roles")
				mac := elm.GetString("MAC")
				ton := elm.GetInt("TimeOn")
				tat := elm.GetInt("TimeAt")
				if (roles & ROLE_ONLINE) == 0 && time.Now().Sub(time.Unix(int64(tlast),0)) > TimeoutOffline {
					DeleteElm("Ping", sys)
				if ip := elm.GetString("IP"); ip == "" {
					DeleteElm("IP", sys)
				} else if _, ok := TouchMapIP[ip]; ok {
					lik.SayError("IP duplicate " + IPToShow(ip) + " daleted")
					DeleteElm("IP", sys)
				} else {
					it := AddIP(sys, ip, elm.GetString("MAC"), elm.GetInt("Roles"))
					it.Namely = elm.GetString("Namely")
					it.TimeOn = elm.GetInt("TimeOn")
					it.TimeOff = elm.GetInt("TimeOff")
					it.SysUnit = elm.GetIDB("SysUnit")
					if TouchMapOld == nil {
						it.SeekOn = time.Now()
					} else if old := TouchMapOld[ip]; old == nil {
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

func (it *ElmTouch) Update() {
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

func AddTouch(sys lik.IDB, ip string, mac string, roles int) *ElmTouch {
	it := &ElmTouch{SysNum: sys, IP: ip, MAC: mac, Roles: roles }
	TouchMapSys[sys] = it
	TouchMapIP[ip] = it
	return it
}

