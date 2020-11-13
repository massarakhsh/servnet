package gate

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likdom"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/controller"
	"github.com/massarakhsh/servnet/ruler"
)

type Station struct {
	controller.DataControl
	Start  int
	Length int
	Total  int
	IdSel  lik.IDB
}

type Stationer interface {
	controller.Controller
}

var StationColumns = []controller.Column{
	{"Number", "#", "32px"},
	{"Namely", "Наименование", "120px"},
	{"IP", "IP", "80px"},
	{"MAC", "MAC", "80px"},
}

func BuildStation(rule ruler.DataRuler, level int, path []string) Stationer {
	it := &Station{}
	rule.SetControl(level, it)
	it.Execute(rule, path)
	return it
}

func (it *Station) ShowMenu(rule ruler.DataRuler) likdom.Domer {
	return nil
}

func (it *Station) ShowInfo(rule ruler.DataRuler) likdom.Domer {
	div := likdom.BuildDivClass("grid")
	_, table := it.ShowGrid("station")
	div.AppendItem(table)
	return div
}

func (it *Station) Execute(rule ruler.DataRuler, path []string) {
	if cmd := ruler.PopCommand(&path); cmd == "" {
	} else if cmd == "stationinit" {
		it.execGridInit(rule)
	} else if cmd == "stationdata" {
		it.execGridData(rule)
	} else if cmd == "select" {
		it.execSelect(rule)
	} else {
		it.ExecuteController(rule, cmd)
	}
}

func (it *Station) Marshal(rule ruler.DataRuler) {
}

func (it *Station) execGridInit(rule ruler.DataRuler) {
	grid := lik.BuildSet()
	grid.SetItem(true, "serverSide")
	grid.SetItem(true, "processing")
	grid.SetItem(it.execInitLanguage(rule), "language")
	//grid.SetItem("400px", "scrollY")
	//grid.SetItem(true, "scrollCollapse")
	//grid.SetItem(false, "paging")
	grid.SetItem(15, "pageLength")
	grid.SetItem(false, "searching")
	grid.SetItem(false, "lengthChange")
	grid.SetItem("single", "select/style")
	if it.IdSel > 0 {
		grid.SetItem(it.IdSel, "likSelect")
	}
	columns := lik.BuildList()
	for _, col := range StationColumns {
		columns.AddItemSet("data", col.Name, "title", col.Title, "width", col.Width, "searchable=true")
	}
	grid.SetItem(columns, "columns")
	rule.SetResponse(grid, "grid")
}

func (it *Station) execInitLanguage(rule ruler.DataRuler) lik.Seter {
	data := lik.BuildSet()
	data.SetItem("Поиск", "search")
	data.SetItem("Таблица пуста", "emptyTable")
	data.SetItem("Строки от _START_ до _END_, всего _TOTAL_", "info")
	data.SetItem("Загрузка ...", "loadingRecords")
	data.SetItem("Обработка ...", "processing")
	data.SetItem("Нет строк в таблице", "infoEmpty")
	data.SetItem("В начало", "paginate/first")
	data.SetItem("Назад", "paginate/previos")
	data.SetItem("Вперёд", "paginate/next")
	data.SetItem("В конец", "paginate/last")
	return data
}

func (it *Station) execGridData(rule ruler.DataRuler) {
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
		if iprm := lik.StrToInt(parm); iprm > 0 && iprm < len(StationColumns) {
			sort = StationColumns[iprm].Name
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

func (it *Station) execSelect(rule ruler.DataRuler) {
	it.IdSel = lik.IDB(lik.StrToInt(rule.Shift()))
}
