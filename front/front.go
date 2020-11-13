package front

import (
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/ruler"
)

type DataRule struct {
	ruler.DataRule
}

type DataRuler interface {
	ruler.DataRuler
}

func BuildRule(page *ruler.DataPage) *DataRule {
	rule := &DataRule{}
	rule.BindPage(page)
	return rule
}

func (rule *DataRule) Execute() lik.Seter {
	rule.SeekPageSize()
	rule.execute()
	if rule.IsNeedRedraw {
		rule.ShowRedraw()
	}
	return rule.GetAllResponse()
}

func (rule *DataRule) execute() {
	if rule.IsShift("front") {
		rule.execute()
	} else if ctrl := rule.seekControl(); ctrl != nil {
		rule.Shift()
		ctrl.Execute(rule, rule.GetPath())
	}
}

func (rule *DataRule) seekControl() ruler.Controller {
	if cid := rule.Top(); cid != "" {
		for lev := len(rule.ItPage.Controls) - 1; lev >= 0; lev-- {
			if ctrl := rule.ItPage.Controls[lev]; ctrl != nil {
				if cid == ruler.GetIdLevel(lev) {
					return ctrl
				}
			}
		}
	}
	return nil
}
