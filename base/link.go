package base

import (
	"github.com/massarakhsh/lik"
)

type ElmLink struct {
	SysNum lik.IDB
	Roles  int
	Path		string
	Map			string
	SysUnitA	lik.IDB
	PortA		string
	SysUnitB	lik.IDB
	PortB		string
}

var LinkMapSys map[lik.IDB]*ElmLink

func LoadLink() {
	if list := GetList("Link"); list != nil {
		LinkMapSys = make(map[lik.IDB]*ElmLink)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sysnum := elm.GetIDB("SysNum")
				it := &ElmLink{SysNum: sysnum}
				it.Path = elm.GetString("Path")
				it.Map = elm.GetString("Map")
				it.Roles = elm.GetInt("Roles")
				it.SysUnitA = elm.GetIDB("SysUnitA")
				it.PortA = elm.GetString("PortA")
				it.SysUnitB = elm.GetIDB("SysUnitB")
				it.PortB = elm.GetString("PortB")
				LinkMapSys[sysnum] = it
			}
		}
	}
}

func (it *ElmLink) Update() {
	set := lik.BuildSet()
	set.SetItem(it.Roles, "Roles")
	set.SetItem(it.Path, "Path")
	set.SetItem(it.Map, "Map")
	set.SetItem(it.SysUnitA, "SysUnitA")
	set.SetItem(it.PortA, "PortA")
	set.SetItem(it.SysUnitB, "SysUnitB")
	set.SetItem(it.PortB, "PortB")
	if it.SysNum > 0 {
		UpdateElm("Link", it.SysNum, set)
	}
}

