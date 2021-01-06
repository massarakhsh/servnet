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
	PortA		int
	SysUnitB	lik.IDB
	PortB		int
	NewPath		string
}

var LinkMapSys map[lik.IDB]*ElmLink

func LoadLink() {
	if list := GetList("Link"); list != nil {
		LinkMapSys = make(map[lik.IDB]*ElmLink)
		for n := 0; n < list.Count(); n++ {
			if elm := list.GetSet(n); elm != nil {
				sys := elm.GetIDB("SysNum")
				it := &ElmLink{SysNum: sys}
				it.Path = elm.GetString("Path")
				it.Map = elm.GetString("Map")
				it.Roles = elm.GetInt("Roles")
				it.SysUnitA = elm.GetIDB("SysUnitA")
				it.PortA = elm.GetInt("PortA")
				it.SysUnitB = elm.GetIDB("SysUnitB")
				it.PortB = elm.GetInt("PortB")
				LinkMapSys[sys] = it
				if unit,_ := UnitMapSys[it.SysUnitA]; unit != nil {
					unit.ListLink = append(unit.ListLink, sys)
				}
				if unit,_ := UnitMapSys[it.SysUnitB]; unit != nil {
					unit.ListLink = append(unit.ListLink, sys)
				}
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
	set.SetItem("CURRENT_TIMESTAMP", "updated_at")
	if it.SysNum > 0 {
		UpdateElm("Link", it.SysNum, set)
	}
}

