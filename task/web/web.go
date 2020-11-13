package web

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likapi"
	"github.com/massarakhsh/servnet/api"
	"github.com/massarakhsh/servnet/front"
	"github.com/massarakhsh/servnet/ruler"
	"log"
	"net/http"
)

func StartHttp() {
	go runHttp()
}

func runHttp() {
	http.HandleFunc("/", routerMain)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", ruler.HostPort), nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func routerMain(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PROPFIND" {
		return
	}
	isapi := lik.RegExCompare(r.RequestURI, "^/api")
	isfront := lik.RegExCompare(r.RequestURI, "^/front")
	ismarshal := lik.RegExCompare(r.RequestURI, "/marshal")
	if !isfront && !ismarshal &&
		lik.RegExCompare(r.RequestURI, "\\.(js|css|htm|html|ico|gif|png|jpg|jpeg|pdf|doc|docx|xls|xlsx)(\\?|$)") {
		likapi.ProbeRouteFile(w, r, r.RequestURI)
		return
	}
	var page *ruler.DataPage
	if sp := lik.StrToInt(likapi.GetParm(r, "_sp")); sp > 0 {
		if pager := likapi.FindPage(sp); pager != nil {
			page = pager.(ruler.DataPager).GetItPage()
		}
	}
	if page == nil {
		page = ruler.StartPage()
	}
	var rule ruler.DataRuler
	if isapi {
		rule = api.BuildRule(page)
	} else {
		rule = front.BuildRule(page)
	}
	rule.LoadRequest(r)
	if !ismarshal {
		rule.RuleLog()
	}
	if !rule.Authority() && !isfront && !ismarshal {

	}
	if isfront {
		json := rule.Execute()
		likapi.RouteJson(w, 200, json, false)
	} else if ismarshal {
		json := rule.Marshal()
		likapi.RouteJson(w, 200, json, false)
	} else if !rule.Authority() {
		likapi.Route401(w, 401, "realm=\"SERVNET\"")
	} else if isapi {
		json := rule.Execute()
		likapi.RouteJson(w, 200, json, false)
	} else {
		html := rule.ShowPage()
		likapi.RouteCookies(w, rule.GetAllCookies())
		likapi.RouteHtml(w, 200, html.ToString())
	}
}
