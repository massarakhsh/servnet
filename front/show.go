package front

import (
	"github.com/massarakhsh/lik/likdom"
	"github.com/massarakhsh/servnet/ruler"
)

func (rule *DataRule) ShowRedraw() {
	rule.StoreItem(rule.showMainGen())
}

func (rule *DataRule) showMainGen() likdom.Domer {
	div := likdom.BuildDivClassId("main_page", "page")
	if len(rule.ItPage.Controls) == 0 {
		if ruler.RootCreator != nil {
			ruler.RootCreator(rule, 0, rule.GetPath())
		}
		//root.BuildRoot(rule, 0, rule.GetPath())
	}
	rule.SeekPath()
	dat := div.BuildDivClass("main_data fill")
	if len(rule.ItPage.Controls) > 0 {
		dat.AppendItem(rule.showControlGen(rule.ItPage.Controls[0]))
	}
	return div
}

func (rule *DataRule) showControlGen(controller ruler.Controller) likdom.Domer {
	tbl := likdom.BuildTableClass("main_data")
	if menu := controller.ShowMenu(rule); menu != nil {
		tbl.BuildTrTdClass("main_data").AppendItem(menu)
	}
	tbl.BuildTrTdClass("main_space")
	dat := tbl.BuildTrTdClass("main_info")
	lev := controller.GetLevel()
	if lev+1 < len(rule.ItPage.Controls) {
		dat.AppendItem(rule.showControlGen(rule.ItPage.Controls[lev+1]))
	} else {
		dat.AppendItem(controller.ShowInfo(rule))
	}
	return tbl
}
