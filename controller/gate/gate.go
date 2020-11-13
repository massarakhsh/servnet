package gate

import (
	"fmt"
	"github.com/massarakhsh/lik/likdom"
	"github.com/massarakhsh/servnet/controller"
	"github.com/massarakhsh/servnet/ruler"
)

type Control struct {
	controller.DataControl
}

type Controller interface {
	controller.Controller
}

func BuildGate(rule ruler.DataRuler, level int, path []string) ruler.Controller {
	it := &Control{}
	rule.SetControl(level, it)
	it.Execute(rule, path)
	return it
}

func (it *Control) ShowMenu(rule ruler.DataRuler) likdom.Domer {
	if it.Mode == "" {
		it.Mode = "summary"
	}
	tbl := likdom.BuildTableClass("menu")
	row := tbl.BuildTr()
	it.MenuItemText(rule, row, "Шлюз")
	it.MenuItemText(rule, row, "|")
	it.MenuItemCmd(rule, row, "summary", "Сводка", "summary")
	it.MenuItemCmd(rule, row, "station", "Станции", "station")
	it.MenuItemCmd(rule, row, "proxy", "Прокси", "proxy")
	it.MenuTools(rule, row)
	return tbl
}

func (it *Control) ShowInfo(rule ruler.DataRuler) likdom.Domer {
	div := likdom.BuildDivClass("grid")
	div.AppendItem(it.buildDiagnFront(rule))
	return div
}

func (it *Control) Execute(rule ruler.DataRuler, path []string) {
	if cmd := ruler.PopCommand(&path); cmd == "" {
	} else if cmd == "summary" {
		it.Mode = "summary"
		rule.SetControl(it.GetLevel()+1, nil)
	} else if cmd == "station" {
		it.Mode = "station"
		BuildStation(rule, it.Level+1, path)
	} else if cmd == "proxy" {
		it.Mode = "proxy"
		BuildProxy(rule, it.Level+1, path)
	} else {
		it.ExecuteController(rule, cmd)
	}
}

func (it *Control) Marshal(rule ruler.DataRuler) {
}

//func (it *DiagnControl) buildLinePurge(rule ruler.DataRuler) likdom.Domer {
//	proc := it.buildProc("purge")
//	line := LinkTe("api", "Очистить", proc)
//	return line
//}

func (it *Control) buildProc(part string) string {
	path := it.BuildPart(part)
	return fmt.Sprintf("%s('%s')", "cmd_proxy", path)
}

func (it *Control) buildDiagnFront(rule ruler.DataRuler) likdom.Domer {
	tbl := likdom.BuildTableClass("")
	row := tbl.BuildTr()
	if td := row.BuildTdClass("column"); td != nil {
		clm := td.BuildTableClass("")
		clm.BuildTrTd().AppendItem(it.buildDiagnInterface(rule))
		clm.BuildTrTd().AppendItem(it.buildDiagnServer(rule))
		clm.BuildTrTd().AppendItem(it.buildDiagnInterface(rule))
	}
	if td := row.BuildTdClass("column"); td != nil {
		clm := td.BuildTableClass("")
		clm.BuildTrTd().AppendItem(it.buildDiagnServer(rule))
		clm.BuildTrTd().AppendItem(it.buildDiagnInterface(rule))
	}
	return tbl
}

func (it *Control) buildDiagnInterface(rule ruler.DataRuler) likdom.Domer {
	tbl := likdom.BuildTableClass("")
	it.buildAppendRow(tbl, "Концентратор", "Ok")
	it.buildAppendRow(tbl, "Концентратор", "Ok")
	it.buildAppendRow(tbl, "Маршрутизатор", "Не отвечает")
	return it.HeadTableString("Интерфейсы", tbl)
}

func (it *Control) buildDiagnServer(rule ruler.DataRuler) likdom.Domer {
	tbl := likdom.BuildTableClass("")
	it.buildAppendRow(tbl, "Сервер", "Ok")
	it.buildAppendRow(tbl, "Сервер", "Не отвечает")
	return it.HeadTableString("Серверы", tbl)
}

func (it *Control) buildAppendRow(tbl likdom.Domer, title string, diagn string) {
	row := tbl.BuildTr()
	row.BuildTdClass("panelinfo").BuildString(title)
	row.BuildTdClass("panelinfo").BuildString(diagn)
}
