package api

import (
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likdom"
	"github.com/massarakhsh/servnet/ruler"
)

const MAX_SEARCH = 9999

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
	responce := rule.GetAllResponse()
	if responce == nil || responce.Count() == 0 {
		responce = lik.BuildSet("disagnosis=error")
	}
	return responce
}

func (rule *DataRule) Marshal() lik.Seter {
	return nil
}

func (rule *DataRule) ShowPage() likdom.Domer {
	return nil
}

func (rule *DataRule) execute() {
	if rule.IsShift("api") {
		rule.execute()
	}
}
