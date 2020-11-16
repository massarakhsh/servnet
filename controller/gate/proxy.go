package gate

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likdom"
	"github.com/massarakhsh/lik/liktable"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/controller"
	"github.com/massarakhsh/servnet/ruler"
	"github.com/tealeg/xlsx"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

type Proxy struct {
	controller.DataControl
	Table *liktable.Table

	DataId   string
	DataSync sync.Mutex
	DataList []lik.Seter

	ServerIndex int
	//ServerLength int
	ServerTime time.Time

	ClientIndex  int
	ClientLength int
	ClientTime   time.Time
	ClientAt     string

	RunStarted bool
	FileDesc   *os.File
	FileNumber int
	FilePos    int64
	DozeData   []string
	DozeDepl   int

	GridActive     bool
	SearchActive   bool
	SearchStop     bool
	SearchDone     bool
	SearchAt       string
	SearchClient   string
	SearchServer   string
	SearchFrom     string
	SearchTo       string
	SearchClientIt string
	SearchFromIt   int64
	SearchToIt     int64
	SearchTime     time.Time
}

type Proxier interface {
	controller.Controller
}

var PathProxyLog = ""

const TimeoutMarshaling = 15 * time.Second
const DelayBeforeSearch = 1500 * time.Millisecond

var LimitLines = 9999
var LimitDoze = 100
var BufferSize = 1024 * 1024

var ProxyColumns = []controller.Column{
	{"Number", "#", "32px"},
	{"Time", "Время", "120px"},
	{"Host", "Компьютер", "60px"},
	{"Client", "Клиент", "80px"},
	{"Server", "Сервер", "128px"},
	{"Info", "Размер", "128px"},
}
var MapStations map[string]string

func BuildProxy(rule ruler.DataRuler, level int, path []string) Proxier {
	if host, _ := os.Hostname(); host == "Shaman" {
		PathProxyLog = "//root/root/var/log/squid3"
	} else if host == "Shamanus" {
		PathProxyLog = "var/squid3"
	} else {
		PathProxyLog = "/var/log/squid3"
	}
	InitializeMapStation()
	it := &Proxy{}
	it.Table = liktable.New("server=false", "page=15")
	for _, col := range ProxyColumns {
		it.Table.AddColumn("searchable=true",
			"data", col.Name, "title", col.Title, "width", col.Width)
	}
	rule.SetControl(level, it)
	it.Execute(rule, path)
	return it
}

func (it *Proxy) ShowMenu(rule ruler.DataRuler) likdom.Domer {
	return nil
}

func (it *Proxy) ShowInfo(rule ruler.DataRuler) likdom.Domer {
	it.GridActive = false
	id, table := it.ShowGrid("proxy")
	it.DataId = id
	div := likdom.BuildDivClass("grid")
	div.AppendItem(it.showFilter(rule))
	div.AppendItem(table)
	return div
}

func (it *Proxy) Execute(rule ruler.DataRuler, path []string) {
	if cmd := ruler.PopCommand(&path); cmd == "" {
	} else if cmd == "proxyinit" {
		it.execGridInit(rule)
	} else if cmd == "proxydata" {
		it.execGridData(rule, path)
	} else if cmd == "proxyrefresh" {
		it.execGridRefresh(rule, path)
	} else if cmd == "proxysearch" {
		it.execGridSearch(rule, path)
	} else if cmd == "clear" {
		it.execClear(rule)
	} else if cmd == "excel" {
		it.execExcel(rule)
	} else if cmd == "stop" {
		it.execStop(rule)
	} else {
		it.ExecuteController(rule, cmd)
	}
}

func (it *Proxy) Marshal(rule ruler.DataRuler) {
	it.serverControl()
	it.clientControl(rule)
}

func (it *Proxy) showFilter(rule ruler.DataRuler) likdom.Domer {
	tbl := likdom.BuildTable()
	row := tbl.BuildTr()
	row.BuildTd().BuildString("<b>Фильтр</b>")
	row.BuildTd().BuildString("&nbsp|")
	row.BuildTd().BuildString("&nbsp;От:")
	if td := row.BuildTd(); td != nil {
		date := td.BuildUnpairItem("input", "type=text",
			"class=tcal", "idgrid", it.DataId, "id=searchfrom")
		if it.SearchFrom == "" {
			it.SearchFrom = time.Now().Add(-60 * 24 * time.Hour).Format("02/01/2006")
		}
		date.SetAttr("value", it.SearchFrom)
	}
	row.BuildTd().BuildString("&nbsp;До:")
	if td := row.BuildTd(); td != nil {
		date := td.BuildUnpairItem("input", "type=text",
			"class=tcal", "idgrid", it.DataId, "id=searchto")
		if it.SearchTo == "" {
			it.SearchTo = time.Now().Format("02/01/2006")
		}
		date.SetAttr("value", it.SearchTo)
	}
	row.BuildTd().BuildString("&nbsp;Клиент:")
	if td := row.BuildTd(); td != nil {
		line := td.BuildUnpairItem("input", "type=text",
			"class=api", "idgrid", it.DataId, "id=searchclient")
		if it.SearchClient != "" {
			line.SetAttr("value", it.SearchClient)
		}
	}
	row.BuildTd().BuildString("&nbsp;Сервер:")
	if td := row.BuildTd(); td != nil {
		line := td.BuildUnpairItem("input", "type=text",
			"class=api", "idgrid", it.DataId, "id=searchserver")
		if it.SearchServer != "" {
			line.SetAttr("value", it.SearchServer)
		}
	}
	row.BuildTdClass("fill")
	row = tbl.BuildTr()
	row.BuildTd().BuildString("<b>Состояние</b>")
	row.BuildTd().BuildString("&nbsp|")
	if td := row.BuildTd("colspan=8"); td != nil {
		td.BuildItem("nobr").AppendItem(it.showStatus(rule))
	}
	row.BuildTdClass("fill")
	return tbl
}

func (it *Proxy) showStatus(rule ruler.DataRuler) likdom.Domer {
	status := likdom.BuildItem("B", "id=searchstatus")
	status.BuildString("&nbsp;")
	if it.SearchActive {
		status.BuildString(" ... поиск ... ")
		status.BuildString(it.SearchAt)
		path := it.BuildPart("stop")
		proc := fmt.Sprintf("front_get('%s')", path)
		status.AppendItem(it.LinkItemProc("СТОП", proc, "cmd"))
	} else {
		status.BuildString(it.SearchAt)
		if it.SearchStop || len(it.DataList) >= LimitLines {
			status.BuildString(" Поиск остановлен")
		} else if it.SearchDone {
			status.BuildString(" Поиск закончен")
		}
		if len(it.DataList) > 0 {
			status.BuildString(". Экспортировать в ")
			path := it.BuildPart("excel")
			proc := fmt.Sprintf("front_get('%s')", path)
			status.AppendItem(it.LinkItemProc("EXCEL", proc, "cmd"))
		}
	}
	it.ClientAt = it.SearchAt
	return status
}

func (it *Proxy) execGridInit(rule ruler.DataRuler) {
	grid := it.Table.Show()
	grid.SetItem(lik.BuildList(), "data")
	rule.SetResponse(grid, "grid")
	/*grid := lik.BuildSet()
	grid.SetItem(it.execInitLanguage(rule), "language")
	grid.SetItem(15, "pageLength")
	grid.SetItem(false, "searching")
	grid.SetItem(false, "sorting")
	grid.SetItem(false, "lengthChange")
	grid.SetItem("single", "select/style")
	columns := lik.BuildList()
	for _,col := range ProxyColumns {
		columns.AddItemSet("data", col.Name, "title", col.Title, "width", col.Width)
	}
	grid.SetItem(columns, "columns")
	grid.SetItem(lik.BuildList(), "data")
	rule.SetResponse(grid, "grid")*/
	it.GridActive = true
	it.searchStart()
}

func (it *Proxy) execGridData(rule ruler.DataRuler, path []string) {
	rule.SetResponse(lik.BuildList(), "data")
}

func (it *Proxy) execGridRefresh(rule ruler.DataRuler, path []string) {
	it.ClientTime = time.Now()
	it.ClientIndex = lik.StrToInt(ruler.PopCommand(&path))
	it.ClientLength = lik.StrToInt(ruler.PopCommand(&path))
	if it.ClientIndex != it.ServerIndex {
		it.ClientLength = 0
	}
	total := len(it.DataList)
	data := lik.BuildList()
	for k := 0; k < LimitDoze && it.ClientLength+k < total; k++ {
		if elm := it.DataList[it.ClientLength+k]; elm != nil {
			data.AddItems(elm)
		}
	}
	rule.SetResponse(it.ServerIndex, "index")
	rule.SetResponse(it.ClientLength, "start")
	rule.SetResponse(total, "total")
	rule.SetResponse(data, "data")
}

func (it *Proxy) execGridSearch(rule ruler.DataRuler, path []string) {
	it.SearchFrom = lik.StringFromXS(rule.GetContext("from"))
	it.SearchTo = lik.StringFromXS(rule.GetContext("to"))
	it.SearchClient = strings.ToLower(lik.StringFromXS(rule.GetContext("client")))
	it.SearchServer = strings.ToLower(lik.StringFromXS(rule.GetContext("server")))
	it.SearchClientIt = it.SearchClient
	if MapStations != nil {
		if !lik.RegExCompare(it.SearchClient, "\\d+\\.\\d+\\.\\d+\\.\\d+") {
			if ip, ok := MapStations[it.SearchClient]; ok {
				it.SearchClientIt = ip
			}
		}
	}
	if match := lik.RegExParse(it.SearchFrom, "(\\d\\d).(\\d\\d).(\\d\\d\\d\\d)"); match != nil {
		it.SearchFromIt = time.Date(lik.StrToInt(match[3]),
			time.Month(lik.StrToInt(match[2])), lik.StrToInt(match[1]),
			0, 0, 0, 0, time.Local).Unix()
	}
	if match := lik.RegExParse(it.SearchTo, "(\\d\\d).(\\d\\d).(\\d\\d\\d\\d)"); match != nil {
		it.SearchToIt = time.Date(lik.StrToInt(match[3]),
			time.Month(lik.StrToInt(match[2])), lik.StrToInt(match[1]),
			23, 59, 59, 0, time.Local).Unix()
	}
	it.searchStart()
}

func (it *Proxy) execClear(rule ruler.DataRuler) {
	it.ClientLength = 0
	rule.SetResponse(it.DataId, "_function_grid_clear")
}

func (it *Proxy) execExcel(rule ruler.DataRuler) {
	it.exportExcel(rule)
}

func (it *Proxy) execStop(rule ruler.DataRuler) {
	it.searchStop(false)
}

func (it *Proxy) clientControl(rule ruler.DataRuler) {
	if it.GridActive && time.Now().Sub(it.ClientTime) >= 3*time.Second {
		if it.ClientAt != it.SearchAt {
			rule.StoreItem(it.showStatus(rule))
		}
		if it.ClientIndex != it.ServerIndex || it.ClientLength != len(it.DataList) {
			it.ClientTime = time.Now()
			cmd := fmt.Sprintf("%s__%d__%d", it.DataId, it.ServerIndex, len(it.DataList))
			rule.SetResponse(cmd, "_function_grid_refresh")
		}
	}
}

////////////////////////////////////////////

func (it *Proxy) serverControl() {
	it.ServerTime = time.Now()
	if !it.RunStarted {
		go it.runStart()
	}
}

func (it *Proxy) runStart() {
	it.runInitialize()
	for time.Now().Sub(it.ServerTime) < TimeoutMarshaling {
		if time.Now().Sub(it.SearchTime) < DelayBeforeSearch {
			time.Sleep(time.Millisecond * 100)
		} else if it.SearchActive {
			it.searchContinue()
		} else {
			time.Sleep(time.Second * 1)
		}
	}
	it.runTerminate()
}

func (it *Proxy) runInitialize() {
	it.RunStarted = true
	it.searchStart()
}

func (it *Proxy) runTerminate() {
	it.RunStarted = false
}

func (it *Proxy) searchStart() {
	it.fileClose()
	it.SearchActive = true
	it.SearchStop = false
	it.SearchDone = false
	it.SearchTime = time.Now()
	it.SearchAt = ""
	it.FileNumber = 0
	it.ServerIndex++
	it.DataList = []lik.Seter{}
}

func (it *Proxy) searchStop(done bool) {
	it.fileClose()
	it.SearchActive = false
	it.SearchDone = done
	it.SearchStop = !done
	it.ClientAt = ""
}

func (it *Proxy) searchContinue() {
	if it.FileDesc == nil {
		it.fileOpen()
	} else if len(it.DataList) >= LimitLines {
		it.searchStop(false)
	} else if it.DozeDepl > 0 {
		it.DozeDepl--
		if it.DozeDepl < len(it.DozeData) {
			if data := it.DozeData[it.DozeDepl]; len(data) > 20 {
				it.searchProbe(data)
			}
		}
	} else if it.FilePos > 0 {
		time.Sleep(time.Millisecond * 1)
		pos := it.FilePos - int64(BufferSize)
		if pos < 0 {
			pos = 0
		}
		size := int(it.FilePos - pos)
		bts := make([]byte, size)
		if sz, err := it.FileDesc.ReadAt(bts, pos); err != nil || sz != size {
			it.searchStop(false)
		} else {
			if pos > 0 {
				for n := 1; n < size; n++ {
					if bts[n-1] == 0xD {
						pos += int64(n)
						size -= n
						bts = bts[n:]
						break
					}
				}
			}
			it.FilePos = pos
			strdata := string(bts)
			isit := true
			var tmu int64
			if match := lik.RegExParse(strdata, "^(\\d\\d\\d\\d\\d\\d\\d\\d\\d\\d)\\."); match != nil {
				tmu = int64(lik.StrToInt(match[1]))
			}
			if tmu > it.SearchToIt {
				isit = false
			}
			if it.SearchClientIt != "" && !strings.Contains(strdata, it.SearchClientIt) {
				isit = false
			} else if it.SearchServer != "" && !strings.Contains(strdata, it.SearchServer) {
				isit = false
			}
			if isit {
				it.DozeData = strings.Split(strdata, "\n")
				it.DozeDepl = len(it.DozeData)
			} else {
				tmg := time.Unix(tmu, 0)
				tmdate := tmg.Format("2006/01/02")
				if tmdate != it.SearchAt {
					it.SearchAt = tmdate
				}
				it.DozeData = []string{}
				it.DozeDepl = len(it.DozeData)
			}
		}
	} else {
		it.fileClose()
		it.FileNumber++
	}
}

func (it *Proxy) fileOpen() {
	if it.FileDesc == nil {
		path := PathProxyLog + "/access.log"
		if it.FileNumber > 0 {
			path += fmt.Sprintf(".%d", it.FileNumber)
		}
		fmt.Printf("Open %s\n", path)
		if fl, err := os.OpenFile(path, os.O_RDONLY, 0666); err != nil {
			it.searchStop(true)
		} else if fi, err := fl.Stat(); err != nil {
			fl.Close()
			it.searchStop(true)
		} else {
			it.FilePos = fi.Size()
			it.FileDesc = fl
			it.DozeData = nil
			it.DozeDepl = 0
		}
	}
}

func (it *Proxy) fileClose() {
	if it.FileDesc != nil {
		it.FileDesc.Close()
		it.FileDesc = nil
	}
}

func (it *Proxy) searchProbe(data string) {
	//1603956787.175 170405 192.168.230.10 TCP_MISS/200 4820 CONNECT fonts.googleapis.com:443 - HIER_DIRECT/74.125.205.95 -
	regex := "(\\d+)\\S*\\s+(\\S+)\\s+(\\S+)\\s+(\\S+)\\s+(\\S+)\\s+(\\S+)\\s+(\\S+)\\s+"
	if match := lik.RegExParse(data, regex); match != nil {
		tm := match[1]
		//tans := match[2]
		client := match[3]
		//status := match[4]
		size := match[5]
		//cmd := match[6]
		server := match[7]
		host := ""
		if MapStations != nil && client != "" {
			if namely, ok := MapStations[client]; ok {
				host = namely
			}
		}
		if len(tm) == 10 && size != "0" {
			ok := true
			tmu := int64(lik.StrToInt(tm))
			tmg := time.Unix(tmu, 0)
			tmdate := tmg.Format("2006/01/02")
			tmtime := tmg.Format("15:04:05")
			if tmdate != it.SearchAt {
				it.SearchAt = tmdate
			}
			if ok && tmu > it.SearchToIt {
				ok = false
			}
			if ok && tmu < it.SearchFromIt {
				ok = false
				it.searchStop(true)
			}
			if ok && it.SearchClientIt != "" {
				if client != it.SearchClientIt {
					ok = false
				}
			}
			if ok && it.SearchServer != "" {
				if srv := strings.ToLower(server); !strings.Contains(srv, it.SearchServer) {
					ok = false
				}
			}
			if ok {
				if len(server) > 50 {
					server = server[:50]
				}
				elm := lik.BuildSet("Number", len(it.DataList)+1)
				elm.SetItem(tmdate+" "+tmtime, "Time")
				elm.SetItem(host, "Host")
				elm.SetItem(client, "Client")
				elm.SetItem(server, "Server")
				elm.SetItem(size, "Info")
				it.DataList = append(it.DataList, elm)
			}
		}
	}
}

func InitializeMapStation() {
	if MapStations == nil {
		ms := make(map[string]string)
		if base.DB != nil {
			if list := base.DB.GetListElm("Unit.Namely,IP.IP",
				"Unit LEFT JOIN IP ON Unit.SysNum=IP.SysUnit", "(Unit.Roles&1)", ""); list != nil {
				for n := 0; n < list.Count(); n++ {
					if elm := list.GetSet(n); elm != nil {
						namely := strings.ToLower(elm.GetString("Namely"))
						ip := elm.GetString("IP")
						if match := lik.RegExParse(ip, "(\\d\\d\\d)(\\d\\d\\d)(\\d\\d\\d)(\\d\\d\\d)"); match != nil && namely != "" {
							ips := fmt.Sprintf("%d.%d.%d.%d",
								lik.StrToInt(match[1]),
								lik.StrToInt(match[2]),
								lik.StrToInt(match[3]),
								lik.StrToInt(match[4]))
							if _, ok := ms[ips]; !ok {
								ms[ips] = namely
							}
							if _, ok := ms[namely]; !ok {
								ms[namely] = ips
							}
						}
					}
				}
			}
		}
		MapStations = ms
	}
}

func (it *Proxy) exportExcel(rule ruler.DataRuler) {
	index := 100000000 + rand.Intn(900000000)
	f := xlsx.NewFile()
	sheet, _ := f.AddSheet("Отчет")
	row := sheet.AddRow()
	row.SetHeight(14)
	for _, col := range ProxyColumns {
		if col.Name != "Number" {
			cell := row.AddCell()
			cell.SetString(col.Title)
		}
	}
	for _, line := range it.DataList {
		row := sheet.AddRow()
		row.SetHeight(14)
		for _, col := range ProxyColumns {
			if col.Name != "Number" {
				cell := row.AddCell()
				cell.SetString(line.GetString(col.Name))
			}
		}
	}
	file := fmt.Sprintf("var/report/%d.xlsx", index)
	f.Save(file)
	rule.SetResponse("/"+file, "_function_lik_window_part")
}
