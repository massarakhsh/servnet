package controller

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likdom"
	"github.com/massarakhsh/servnet/ruler"
	"strings"
)

type DataControl struct {
	ruler.DataControl
}

type Controller interface {
	ruler.Controller
}

func (it *DataControl) BuildPart(part string) string {
	return "/" + ruler.GetIdLevel(it.Level) + "/" + part
}

func (it *DataControl) BuildUrl(rule ruler.DataRuler, part string) string {
	return rule.BuildUrl(it.BuildPart(part))
}

func (it *DataControl) BuildProc(proc string, part string, parm string) string {
	parms := "'" + it.BuildPart(part) + "'"
	if parm != "" {
		parms += "," + parm
	}
	return proc + "(" + parms + ")"
}

func (it *DataControl) MenuPrepare(rule ruler.DataRuler, state bool) likdom.Domer {
	id := fmt.Sprintf("menu_%d", it.Level)
	tbl := likdom.BuildTableClass("menu", "id", id)
	if !state && it.Level+1 >= rule.GetLevel() {
		it.Mode = ""
	}
	return tbl
}

func (it *DataControl) MenuItemText(rule ruler.DataRuler, row likdom.Domer, text string) {
	it.MenuItemCmd(rule, row, "", text, "")
}

func (it *DataControl) MenuItemImg(rule ruler.DataRuler, row likdom.Domer, mode string, txt string, img string, cmd string) {
	text := ""
	if img != "" {
		item := likdom.BuildUnpairItem("img", "src", img)
		if txt != "" {
			item.SetAttr("title", txt)
		}
		text = item.ToString()
	}
	it.MenuItemCmd(rule, row, mode, text, cmd)
}

func (it *DataControl) MenuItemCmd(rule ruler.DataRuler, row likdom.Domer, mode string, txt string, cmd string) {
	proc := ""
	if cmd != "" {
		path := it.BuildPart(cmd)
		proc = fmt.Sprintf("front_get('%s')", path)
	}
	it.MenuItemProc(rule, row, mode, txt, proc)
}

func (it *DataControl) MenuItemProc(rule ruler.DataRuler, row likdom.Domer, mode string, txt string, proc string) {
	cls := "menu"
	if mode != "" && mode == it.GetMode() {
		cls += " menu_select"
	}
	td := row.BuildTdClass(cls)
	if proc != "" {
		a := td.BuildItem("a", "href=#", "onclick", proc)
		a.BuildString(txt)
	} else {
		td.BuildString(txt)
	}
}

func (it *DataControl) MenuItemSep(rule ruler.DataRuler, row likdom.Domer) {
	td := row.BuildTdClass("menu fill")
	td.BuildString("&nbsp;")
}

func (it *DataControl) MenuTools(rule ruler.DataRuler, row likdom.Domer) {
	it.MenuItemSep(rule, row)
	if it.Level == 0 {
		row.BuildTdClass("menu").BuildString("<span id=srvtime class=srvtime></span>")
	} else {
		it.MenuItemImg(rule, row, "", "Закрыть", "/images/menuexit.png", "exit")
	}
}

func (it *DataControl) LinkItemImg(pic string, txt string, cmd string, cls string) likdom.Domer {
	img := likdom.BuildUnpairItem("img", "src", pic)
	if txt != "" {
		img.SetAttr("title", txt)
	}
	return it.LinkItemCmd(img.ToString(), cmd, cls)
}

func (it *DataControl) LinkItemCmd(txt string, cmd string, cls string) likdom.Domer {
	proc := ""
	if cmd != "" {
		path := it.BuildPart(cmd)
		proc = fmt.Sprintf("front_get('%s')", path)
	}
	return it.LinkItemProc(txt, proc, cls)
}

func (it *DataControl) LinkItemProc(txt string, proc string, cls string) likdom.Domer {
	a := likdom.BuildItemClass("a", cls, "href=#", "onclick", proc)
	a.BuildString(txt)
	return a
}

func (it *DataControl) CollectParms(rule ruler.DataRuler, prefix string) lik.Seter {
	parms := lik.BuildSet()
	if context := rule.GetAllContext(); context != nil {
		for _, set := range context.Values() {
			if strings.HasPrefix(set.Key, prefix) && set.Val != nil {
				str := lik.StringFromXS(set.Val.ToString())
				parms.SetItem(str, set.Key[len(prefix):])
			}
		}
	}
	return parms
}

func (it *DataControl) ExecuteController(rule ruler.DataRuler, cmd string) {
	if cmd == "exit" {
		rule.SetControl(it.GetLevel(), nil)
	} else if cmd == "seek" {
		rule.SetControl(it.GetLevel()+1, nil)
	}
}

func (it *DataControl) HeadTableString(head string, body likdom.Domer) likdom.Domer {
	return it.HeadTable(likdom.BuildString(head), body)
}

func (it *DataControl) HeadTable(head likdom.Domer, body likdom.Domer) likdom.Domer {
	code := likdom.BuildTableClass("panel")
	code.BuildTrTdClass("panelhead").AppendItem(head)
	code.BuildTrTdClass("panelbody").AppendItem(body)
	return code
}
