package base

import (
	"fmt"
	"github.com/massarakhsh/lik"
)

type netPoint struct {
	unit	*ElmUnit
}

var netList []netPoint
var netPos int

func NetLink() {
	netInit()
	netLarge()
	netSave()
}

func netInit() {
	netList = []netPoint{}
	netPos = 0
	for _, unit := range UnitMapSys {
		if (unit.Roles & ROLE_ROOT) != 0 {
			netList = append(netList, netPoint{unit})
			unit.NewPath = fmt.Sprintf("#%d", int(unit.SysNum))
		} else {
			unit.NewPath = ""
		}
	}
	for _, link := range LinkMapSys {
		link.NewPath = ""
	}
}

func netLarge() {
	for netPos < len(netList) {
		unit := netList[netPos].unit
		for _,syslink := range unit.ListLink {
			if link := LinkMapSys[syslink]; link != nil && link.NewPath == "" {
				path := unit.NewPath
				systo := lik.IDB(0)
				porto := 0
				if unit.SysNum == link.SysUnitA {
					path += fmt.Sprintf("_%02d", link.PortA)
					link.NewPath = path
					systo = link.SysUnitB
					porto = link.PortB
				} else if unit.SysNum == link.SysUnitB {
					path += fmt.Sprintf("_%02d", link.PortB)
					link.NewPath = path
					systo = link.SysUnitA
					porto = link.PortA
				}
				if unito := UnitMapSys[systo]; unito != nil && unito.NewPath == "" {
					path += fmt.Sprintf("#%d", int(link.SysNum))
					path += fmt.Sprintf("@%02d", porto)
					path += fmt.Sprintf("#%d", int(unito.SysNum))
					unito.NewPath = path
					netList = append(netList, netPoint{unito})
				}
			}
		}
		netPos++
	}
}

func netSave() {
	for _, unit := range UnitMapSys {
		oldpath := unit.Path
		modify := false
		if unit.Path != unit.NewPath {
			unit.Path = unit.NewPath
			modify = true
		}
		if unit.Path != "" && (unit.Roles & ROLE_LINKED) == 0 {
			unit.Roles |= ROLE_LINKED
			modify = true
		} else if unit.Path == "" && (unit.Roles & ROLE_LINKED) != 0 {
			unit.Roles ^= ROLE_LINKED
			modify = true
		}
		if modify {
			fmt.Printf("Unit %s: %s ===> %s\n", unit.Namely, unit.Path, oldpath)
			unit.Update()
		}
	}
	for _, link := range LinkMapSys {
		modify := false
		if link.Path != link.NewPath {
			link.Path = link.NewPath
			modify = true
		}
		if link.Path != "" && (link.Roles & ROLE_LINKED) == 0 {
			link.Roles |= ROLE_LINKED
			modify = true
		} else if link.Path == "" && (link.Roles & ROLE_LINKED) != 0 {
			link.Roles ^= ROLE_LINKED
			modify = true
		}
		if modify {
			link.Update()
		}
	}
}

