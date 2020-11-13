package front

import (
	"fmt"
	"github.com/massarakhsh/lik/likdom"
	"github.com/massarakhsh/servnet/base"
)

func (rule *DataRule) ShowPage() likdom.Domer {
	rule.SayInfo(fmt.Sprintf("ShowPage id=%d", rule.ItPage.GetPageId()))
	html := rule.InitializePage(base.Version)
	if head, _ := html.GetDataTag("head"); head != nil {
		head.BuildItem("title").BuildString("Servnet")
		head.BuildString("<script type='text/javascript' src='/lib/jquery.js'></script>")
		head.BuildString("<script type='text/javascript' src='/lib/datatables.js'></script>")
		head.BuildString("<link rel='stylesheet' href='/lib/datatables.css'/>")
		head.BuildString("<script type='text/javascript' src='/lib/dropzone.js'></script>")
		head.BuildString("<link rel='stylesheet' href='/lib/dropzone.css'/>")
		head.BuildString("<script type='text/javascript' src='/lib/tcal.js'></script>")
		head.BuildString("<link rel='stylesheet' href='/lib/tcal.css'/>")
		head.BuildString("<script type='text/javascript' src='/lib/jquery.timepicker.js'></script>")
		head.BuildString("<link rel='stylesheet' href='/lib/jquery.timepicker.css'/>")
		head.BuildString("<link rel='stylesheet' href='/js/grid.css'/>")
		head.BuildString("<script type='text/javascript' src='/js/lik.js'></script>")
		head.BuildString("<script type='text/javascript' src='/js/script.js'></script>")
		head.BuildString("<script type='text/javascript' src='/js/request.js'></script>")
		head.BuildString("<script type='text/javascript' src='/js/grid.js'></script>")
		head.BuildString("<script type='text/javascript' src='/js/form.js'></script>")
		head.BuildString("<link rel='stylesheet' href='/js/styles.css'/>")
	}
	if body, _ := html.GetDataTag("body"); body != nil {
		if script := body.BuildItem("script"); script != nil {
			code := "script_start();\r\n"
			script.BuildString("jQuery(document).ready(function () { " + code + " });")
		}
		tbl := body.BuildTableClass("fill")
		tbl.BuildTrTd().AppendItem(rule.showMainGen())
		//tbl.BuildTrTd().AppendItem(rule.showProtocol())
		tbl.BuildTrTdClass("fill")
	}
	return html
}

/*func (rule *DataRule) showProtocol() likdom.Domer {
	div := likdom.BuildDivClassId("protocol", "protocol")
	tbl := div.BuildTable()
	tbl.BuildTrTd().BuildString("<hr>")
	if protos := lik.GetProtoLog(250); protos != nil {
		for np := len(protos) - 1; np >= 0; np-- {
			tbl.BuildTrTdClass("protocol").BuildString(protos[np])
		}
	}
	return div
}
*/
