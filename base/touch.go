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
	TimeOn	int
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
				ton := elm.GetInt("TimeOn")
				tat := elm.GetInt("TimeAt")
				if UnitMapSys[sysunit] == nil || mac == "" {
					lik.SayError("Touch switch daleted")
					DeleteElm("Touch", sys)
				} else if AddTouch(sys, sysunit, port, mac, ton, tat, roles) == nil {
					lik.SayError("Touch duplicate daleted")
					DeleteElm("Touch", sys)
				}
			}
		}
	}
}

func GetTouch(sysunit lik.IDB, port int, mac string) *ElmTouch {
	ipm := fmt.Sprintf("%d_%d_%s", sysunit, port, mac)
	touch := TouchMapIPM[ipm]
	return touch
}

func AddTouch(sys lik.IDB, sysunit lik.IDB, port int, mac string, ton int, at int, roles int) *ElmTouch {
	ipm := fmt.Sprintf("%d_%d_%s", sysunit, port, mac)
	if TouchMapIPM[ipm] != nil {
		return nil
	}
	touch := &ElmTouch{SysNum: sys, SysUnit: sysunit, Port: port, MAC: mac, TimeOn: ton, TimeAt: at, Roles: roles}
	TouchMapIPM[ipm] = touch
	if sys > 0 {
		TouchMapSys[sys] = touch
	}
	return touch
}

func (it *ElmTouch) Update() {
	set := lik.BuildSet()
	set.SetItem(it.SysUnit, "SysUnit")
	set.SetItem(it.Roles, "Roles")
	set.SetItem(it.Port, "Port")
	set.SetItem(it.MAC, "MAC")
	set.SetItem(it.TimeOn, "TimeOn")
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

func TouchOnline(sysunit lik.IDB, port int, mac string, secs int) {
	at := int(time.Now().Add(-time.Duration(secs) * time.Second).Unix())
	if touch := GetTouch(sysunit, port, mac); touch != nil {
		if (touch.Roles & ROLE_ONLINE) == 0 {
			touch.Roles |= ROLE_ONLINE
			touch.Update()
		}
		if at > touch.TimeAt {
			touch.TimeAt = at
			touch.Update()
		}
	} else if touch = AddTouch(0, sysunit, port, mac, at, at, ROLE_ONLINE); touch != nil {
		touch.Update()
		if DebugLevel > 0 {
			lik.SayError("New Touch created")
		}
	}
}

func TouchTerminate() {
	for _,touch := range TouchMapSys {
		at := int(time.Now().Unix())
		if touch.SysNum == 354609 {
			at += 0
		}
		if (touch.Roles & ROLE_ONLINE) != 0 && at - touch.TimeAt > 10 * 60 {
			touch.Roles ^= ROLE_ONLINE
			touch.Update()
		} else if (touch.Roles & ROLE_ONLINE) == 0 && at - touch.TimeAt > 3 * 24 * 60 * 60 {
			if DebugLevel > 0 {
				lik.SayError("Old Touch daleted")
			}
			DeleteElm("Touch", touch.SysNum)
			delete(TouchMapSys, touch.SysNum)
		}
	}
}

