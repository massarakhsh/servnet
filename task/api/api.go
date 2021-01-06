package api

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likapi"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type APIer struct {
	task.Task
}

type regTable struct {
	Elms	map[lik.IDB]*regElm
}

type regElm struct {
	Index	int
	DT		string
	Exists	bool
	Data	lik.Seter
}

var indexData = 0
var registerData = make(map[string]*regTable)
var syncData = sync.Mutex{}

func StartAPI() {
	go func() {
		apier := &APIer{}
		apier.update()
		apier.Initialize("APIer", apier)
		go apier.runHttp()
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
	json := Answer(0)
	if strings.HasPrefix(r.RequestURI, "/api/stop") {
		base.IsStoping = true
	}
	likapi.RouteJson(w, json, "relay=*")
}

func (it *APIer) update() {
	UpdateBase([]string { "IPZone", "IP", "Ping", "Unit", "Link" }, "SysNum")
}

func UpdateBase(tables []string, key string) {
	syncData.Lock()
	idxnext := indexData
	for _, table := range tables {
		tbl := registerData[table]
		if tbl == nil {
			tbl = &regTable{}
			tbl.Elms = make(map[lik.IDB]*regElm)
			registerData[table] = tbl
		}
		if list := base.GetList(table); list != nil {
			for _,val := range tbl.Elms {
				val.Exists = false
			}
			for ne := 0; ne < list.Count(); ne++ {
				if elm := list.GetSet(ne); elm != nil {
					id := elm.GetIDB(key)
					dt := elm.GetString("updated_at")
					telm := tbl.Elms[id]
					if telm == nil {
						if idxnext == indexData {
							idxnext++
						}
						telm = &regElm{}
						telm.DT = dt
						telm.Index = idxnext
						telm.Exists = true
						telm.Data = elm
						tbl.Elms[id] = telm
					} else if dt != telm.DT || telm.Data == nil {
						if idxnext == indexData {
							idxnext++
						}
						telm.DT = dt
						telm.Index = idxnext
						telm.Exists = true
						telm.Data = elm
					} else {
						telm.Exists = true
					}
				}
			}
			for _,telm := range tbl.Elms {
				if !telm.Exists && telm.Data != nil {
					if idxnext == indexData {
						idxnext++
					}
					telm.Index = idxnext
					telm.Data = nil
				}
			}
		}
	}
	if indexData != idxnext {
		if base.DebugLevel > 0 {
			lik.SayInfo(fmt.Sprintf("Changes №: %d", idxnext))
		}
		indexData = idxnext
	}
	syncData.Unlock()
}

func Answer(index int) lik.Seter {
	syncData.Lock()
	answer := lik.BuildSet()
	answer.SetItem(indexData, "index")
	for table,tbl := range registerData {
		anstab := lik.BuildSet()
		answer.SetItem(anstab, table)
		for id,telm := range tbl.Elms {
			if telm.Index > index || index > indexData {
				if telm.Data != nil {
					anstab.SetItem(telm.Data, lik.IDBToStr(id))
				} else {
					anstab.SetItem(lik.BuildSet(), lik.IDBToStr(id))
				}
			}
		}
	}
	syncData.Unlock()
	return answer
}

