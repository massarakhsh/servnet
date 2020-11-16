package unit

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likdom"
	"github.com/massarakhsh/lik/liktable"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/controller"
	"github.com/massarakhsh/servnet/ruler"
)

type Unit struct {
	controller.DataControl
	Table  *liktable.Table
	Start  int
	Length int
	Total  int
	IdSel  lik.IDB
}

type Uniter interface {
	controller.Controller
}

var UnitColumns = []controller.Column{
	{"Number", "#", "32px"},
	{"Namely", "Наименование", "120px"},
	{"IP", "IP", "80px"},
	{"MAC", "MAC", "80px"},
}

func BuildUnit(rule ruler.DataRuler, level int, path []string) Uniter {
	it := &Unit{}
	it.Table = liktable.New("server=true", "page=15")
	for _, col := range UnitColumns {
		it.Table.AddColumn("searchable=true",
			"data", col.Name, "title", col.Title, "width", col.Width)
	}
	rule.SetControl(level, it)
	it.Execute(rule, path)
	return it
}

func (it *Unit) ShowMenu(rule ruler.DataRuler) likdom.Domer {
	return nil
}

func (it *Unit) ShowInfo(rule ruler.DataRuler) likdom.Domer {
	path := it.BuildPart("unit")
	return it.Table.Initialize(path)
}

func (it *Unit) Execute(rule ruler.DataRuler, path []string) {
	if cmd := ruler.PopCommand(&path); cmd == "" {
	} else if cmd == "unitinit" {
		it.execGridInit(rule)
	} else if cmd == "unitdata" {
		it.execGridData(rule)
	} else if cmd == "select" {
		it.execSelect(rule)
	} else {
		it.ExecuteController(rule, cmd)
	}
}

func (it *Unit) Marshal(rule ruler.DataRuler) {
}

func (it *Unit) execGridInit(rule ruler.DataRuler) {
	grid := it.Table.Show()
	rule.SetResponse(grid, "grid")
}

func (it *Unit) execGridData(rule ruler.DataRuler) {
	if parm := rule.GetContext("draw"); parm != "" {
		rule.SetResponse(lik.StrToInt(parm), "draw")
	}
	if parm := rule.GetContext("start"); parm != "" {
		it.Start = lik.StrToInt(parm)
	}
	if parm := rule.GetContext("length"); parm != "" {
		it.Length = lik.StrToInt(parm)
	}
	if it.Length == 0 {
		it.Length = 20
	}
	sort := ""
	if parm := rule.GetContext("order[0][column]"); parm != "" {
		if iprm := lik.StrToInt(parm); iprm > 0 && iprm < len(UnitColumns) {
			sort = UnitColumns[iprm].Name
		}
	}
	desc := rule.GetContext("order[0][dir]") == "desc"
	order := ""
	if sort == "Namely" {
		order = "Unit.Namely"
		if desc {
			order += " desc"
		}
	} else if sort == "IP" {
		order = "IP.IP"
		if desc {
			order += " desc"
		}
	} else if sort == "MAC" {
		order = "IP.MAC"
		if desc {
			order += " desc"
		}
	}
	if order != "" {
		order += ","
	}
	order += "Unit.SysNum,Unit.Namely,IP.IP"
	data := lik.BuildList()
	var list lik.Lister
	if base.DB != nil {
		list = base.DB.GetListElm("Unit.SysNum,Unit.Namely,IP.IP,IP.MAC",
			"Unit LEFT JOIN IP ON Unit.SysNum=IP.SysUnit",
			"(Unit.Roles&1)", order)
	}
	if list != nil {
		it.Total = list.Count()
		rule.SetResponse(it.Total, "recordsTotal")
		rule.SetResponse(it.Total, "recordsFiltered")
		for n := 0; n < it.Length && it.Start+n < it.Total; n++ {
			nr := (it.Start + n)
			if elm := list.GetSet(nr); elm != nil {
				id := elm.GetIDB("Id")
				row := lik.BuildSet("DT_RowId", id)
				row.SetItem(1+n, "Number")
				row.SetItem(elm.GetString("Namely"), "Namely")
				ip := ""
				if match := lik.RegExParse(elm.GetString("IP"), "(\\d\\d\\d)(\\d\\d\\d)(\\d\\d\\d)(\\d\\d\\d)"); match != nil {
					ip = fmt.Sprintf("%d.%d.%d.%d",
						lik.StrToInt(match[1]),
						lik.StrToInt(match[2]),
						lik.StrToInt(match[3]),
						lik.StrToInt(match[4]))
				}
				row.SetItem(ip, "IP")
				mac := ""
				if match := lik.RegExParse(elm.GetString("MAC"), "(\\w\\w)(\\w\\w)(\\w\\w)(\\w\\w)(\\w\\w)(\\w\\w)"); match != nil {
					mac = fmt.Sprintf("%s:%s:%s:%s:%s:%s", match[1], match[2], match[3], match[4], match[5], match[6])
				}
				row.SetItem(mac, "MAC")
				data.AddItems(row)
			}
		}
	}
	rule.SetResponse(data, "data")
}

func (it *Unit) execSelect(rule ruler.DataRuler) {
	it.IdSel = lik.IDB(lik.StrToInt(rule.Shift()))
}
