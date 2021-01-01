package api

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likapi"
	"github.com/massarakhsh/lik/likmarshal"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"log"
	"net/http"
	"strings"
	"time"
)

type APIer struct {
	task.Task
}

func StartAPI() {
	go func() {
		apier := &APIer{}
		apier.update()
		apier.runHttp()
		apier.Initialize("APIer", apier)
	}()
}

func (it *APIer) DoStep() {
	it.update()
	it.SetPause(time.Second * 15)
}

func (it *APIer) runHttp() {
	http.HandleFunc("/", it.routerHttp)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", base.ConfPort), nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (it *APIer) routerHttp(w http.ResponseWriter, r *http.Request) {
	lik.SayInfo("Request: " + r.RequestURI)
	json := likmarshal.Answer(0)
	if strings.HasPrefix(r.RequestURI, "/api/stop") {
		base.IsStoping = true
	}
	likapi.RouteJson(w, 200, json, false)
}

func (it *APIer) update() {
	likmarshal.UpdateBase(base.DB, []string { "IPZone", "IP", "Ping", "Unit", "Link" }, "SysNum")
}

