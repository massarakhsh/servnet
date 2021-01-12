package base

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"time"
)

type ElmTouch struct {
	SysNum lik.IDB
	SysUnit lik.IDB
	Roles  	int
	Port	int
	MAC    	string
	TimeAt	int
}

var TouchMapSys map[lik.IDB]*ElmTouch
var TouchMapIPM map[string]*ElmTouch
var TouchMapOld map[lik.IDB]*ElmTouch

func LoadTouch() {
	if list := GetList("Touch"); list != nil {
		TouchMapOld = TouchMapSys
		TouchMapSys = make(map[lik.IDB]*ElmTouch)
		TouchMapIPM = make(map[string]*ElmTouch)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sys := elm.GetIDB("SysNum")
				sysunit := elm.GetIDB("SysUnit")
				port := elm.GetInt("Port")
				roles := elm.GetInt("Roles")
				mac := elm.GetString("MAC")
				tat := elm.GetInt("TimeAt")
				if UnitMapSys[sysunit] == nil || mac == "" {
					DeleteElm("Touch", sys)
				} else if TouchFind(sysunit, port, mac) != nil {
					lik.SayError("Touch duplicate daleted")
					DeleteElm("Touch", sys)
				} else if (roles & ROLE_ONLINE) == 0 && time.Now().Sub(time.Unix(int64(tat),0)) > TimeoutOffline {
					DeleteElm("Touch", sys)
				} else {
					AddTouch(sys, sysunit, port, mac, tat, roles)
				}
			}
		}
	}
}

func TouchFind(sysunit lik.IDB, port int, mac string) *ElmTouch {
	ipm := fmt.Sprintf("%d_%d_%s", sysunit, port, mac)
	touch := TouchMapIPM[ipm]
	return touch
}

func AddTouch(sys lik.IDB, sysunit lik.IDB, port int, mac string, at int, roles int) *ElmTouch {
	it := &ElmTouch{SysNum: sys, SysUnit: sysunit, Port: port, MAC: mac, TimeAt: at, Roles: roles}
	ipm := fmt.Sprintf("%d_%d_%s", sysunit, port, mac)
	TouchMapIPM[ipm] = it
	return it
}

func (it *ElmTouch) Update() {
	set := lik.BuildSet()
	set.SetItem(it.SysUnit, "SysUnit")
	set.SetItem(it.Roles, "Roles")
	set.SetItem(it.Port, "Port")
	set.SetItem(it.MAC, "MAC")
	set.SetItem(it.TimeAt, "TimeAt")
	set.SetItem("CURRENT_TIMESTAMP", "updated_at")
	if it.SysNum > 0 {
		UpdateElm("Touch", it.SysNum, set)
	} else {
		set.SetItem("CURRENT_TIMESTAMP", "created_at")
		it.SysNum = InsertElm("Touch", set)
		TouchMapSys[it.SysNum] = it
	}
}
