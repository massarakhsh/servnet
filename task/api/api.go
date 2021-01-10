package api

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likapi"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"log"
	"net/http"
	"sync"
	"time"
)

const ID = "SysNum"

type APIer struct {
	task.Task
}

type regTable struct {
	Table	string
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
		apier.Initialize("APIer", apier)
		go apier.runHttp()
	}()
}

func (it *APIer) runHttp() {
	http.HandleFunc("/", it.routerHttp)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", base.ConfPort), nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (it *APIer) DoStep() {
	it.updateAll()
	it.SetPause(time.Second * 15)
}

func (it *APIer) updateAll() {
	syncData.Lock()
	for table,_ := range registerData {
		it.updateTable(table)
	}
	syncData.Unlock()
}

func (it *APIer) updateTable(table string) {
	idxnext := indexData
	tbl := registerData[table]
	if tbl == nil {
		tbl = &regTable{ Table: table }
		tbl.Elms = make(map[lik.IDB]*regElm)
		registerData[table] = tbl
	}
	if list := base.GetList(table); list != nil {
		for _,val := range tbl.Elms {
			val.Exists = false
		}
		for ne := 0; ne < list.Count(); ne++ {
			if elm := list.GetSet(ne); elm != nil {
				id := elm.GetIDB(ID)
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
	if indexData != idxnext {
		if base.DebugLevel > 0 {
			lik.SayInfo(fmt.Sprintf("Changes №: %d", idxnext))
		}
		indexData = idxnext
	}
}

func (it *APIer) routerHttp(w http.ResponseWriter, r *http.Request) {
	lik.SayInfo("Request: " + r.RequestURI)
	names := lik.PathToNames(r.RequestURI)
	if len(names) > 0 && names[0] == "api" {
		it.routeAPI(w, names[1:])
	}
}

func (it *APIer) routeAPI(w http.ResponseWriter, names []string) {
	if len(names) > 0 && names[0] != "" {
		table := names[0]
		names = names[1:]
		index := 0
		if len(names) > 0 {
			index = lik.StrToInt(names[0])
			names = names[1:]
		}
		if answer := it.getTable(table, index); answer != nil {
			likapi.RouteJson(w, answer, "relay=*")
		}
	}
}

func (it *APIer) getTable(table string, index int) lik.Seter {
	syncData.Lock()
	defer syncData.Unlock()
	tbl := registerData[table]
	if tbl == nil {
		it.updateTable(table)
		tbl = registerData[table]
	}
	answer := it.Answer(tbl, index)
	return answer
}

func (it *APIer) Answer(tbl *regTable, index int) lik.Seter {
	answer := lik.BuildSet()
	answer.SetItem(indexData, "index")
	if tbl != nil {
		anstab := lik.BuildSet()
		answer.SetItem(anstab, tbl.Table)
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
	return answer
}

