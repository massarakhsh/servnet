package base

import (
	"github.com/massarakhsh/lik"
)

type ElmUnit struct {
	SysNum lik.IDB
	Roles  int
	Namely string
	Path		string
	Map			string
	IPs		[]lik.IDB
}

var MapSysUnit map[lik.IDB]*ElmUnit
var MapNamelyUnit map[string]*ElmUnit

func LoadUnit() {
	if list := GetList("Unit"); list != nil {
		MapSysUnit = make(map[lik.IDB]*ElmUnit)
		MapNamelyUnit = make(map[string]*ElmUnit)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sysnum := elm.GetIDB("SysNum")
				it := &ElmUnit{SysNum: sysnum}
				it.Roles = elm.GetInt("Roles")
				it.Namely = elm.GetString("Namely")
				it.Path = elm.GetString("Path")
				it.Map = elm.GetString("Map")
				MapSysUnit[sysnum] = it
				if _, ok := MapNamelyUnit[it.Namely]; !ok {
					MapNamelyUnit[it.Namely] = it
				}
			}
		}
	}
}

func (it *ElmUnit) NetUpdate() {
	ip := ""
	for _,sysip := range it.IPs {
		if elmip,_ := IPMapSys[sysip]; elmip != nil {
			if (elmip.Roles & 0x1000) != 0 {
				ip = elmip.IP
				break
			}
		}
	}
	if ip != "" && (it.Roles & 0x1000) == 0 {
		it.Roles |= 0x1000
		it.Update()
		AddEvent(ip, "", it.Namely, "ON unit")
	} else if ip == "" && (it.Roles & 0x1000) != 0 {
		it.Roles ^= 0x1000
		it.Update()
		AddEvent("", "", it.Namely, "OFF unit")
	}
}

func (it *ElmUnit) Update() {
	set := lik.BuildSet()
	set.SetItem(it.Roles, "Roles")
	set.SetItem(it.Path, "Path")
	set.SetItem(it.Map, "Map")
	if it.SysNum > 0 {
		UpdateElm("Unit", it.SysNum, set)
	}
}

