package base

import (
	"github.com/massarakhsh/lik"
)

type ElmUnit struct {
	SysNum lik.IDB
	Roles  int
	Namely string
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
				MapSysUnit[sysnum] = it
				if _, ok := MapNamelyUnit[it.Namely]; !ok {
					MapNamelyUnit[it.Namely] = it
				}
			}
		}
	}
}
