package base

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/lik/likbase"
	"sync"
	"time"
)

const (
	Version = "0.1"
)

var dbLock sync.Mutex
var dbOk bool
var DB likbase.DBaser

func OpenDB(serv string, name string, user string, pass string) bool {
	likbase.FId = "SysNum"
	logon := user + ":" + pass
	addr := "tcp(" + serv + ":3306)"
	if DB = likbase.OpenDBase("mysql", logon, addr, name); DB == nil {
		lik.SayError(fmt.Sprint("DB not opened"))
		return false
	}
	LoadTables()
	return true
}

func LockDB() {
	dbLock.Lock()
}

func UnlockDB() {
	dbLock.Unlock()
}

func WaitDB() bool {
	for tw := 0; tw < 30; tw++ {
		if dbOk {
			return true
		}
		time.Sleep(time.Millisecond * 100)
	}
	return false
}

func LoadTables() {
	LockDB()
	dbOk = false
	InitAsk()
	LoadUnit()
	LoadLink()
	LoadIP()
	LoadPing()
	NetLink()
	dbOk = true
	UnlockDB()
}

func GetElm(part string, id lik.IDB) lik.Seter {
	return DB.GetOneById(part, id)
}

func InsertElm(part string, sets lik.Seter) lik.IDB {
	if (HostModes & MODE_REAL) == 0 { return 0 }
	return DB.InsertElm(part, sets)
}

func UpdateElm(part string, id lik.IDB, sets lik.Seter) bool {
	if (HostModes & MODE_REAL) == 0 { return false }
	return DB.UpdateElm(part, id, sets)
}

func DeleteElm(part string, id lik.IDB) bool {
	if (HostModes & MODE_REAL) == 0 { return false }
	return DB.DeleteElm(part, id)
}

func GetList(part string) lik.Lister {
	return DB.GetListElm("*", part, "", "SysNum")
}

func GetLastId(part string) lik.IDB {
	id, _ := DB.CalculeIDB(DB.PrepareSql("MAX(SysNum)", part, "", ""))
	return id
}

func CalculateString(sql string) string {
	val := ""
	if one := DB.GetOneBySql(sql); one != nil {
		for _, set := range one.Values() {
			if set.Val != nil {
				val = set.Val.ToString()
				break
			}
		}
	}
	return val
}

func GetParm(key string) string {
	return CalculateString(fmt.Sprintf("SELECT Value FROM LikParam WHERE Namely='%s'", key))
}

func SetParm(key string, val string) {
	if one := DB.GetOneBySql(fmt.Sprintf("SELECT SysNum,Value FROM LikParam WHERE Namely='%s'", key)); one != nil && val == "" {
		DeleteElm("LikParam", one.GetIDB("SysNum"))
	} else if one != nil && val != one.GetString("Value") {
		UpdateElm("LikParam", one.GetIDB("SysNum"), lik.BuildSet("Value", val))
	} else if val != "" {
		InsertElm("LikParam", lik.BuildSet("Namely", key, "Value", val))
	}
}

func AddEvent(ip string, mac string, namely string, formula string) {
	at := int(time.Now().Unix())
	set := lik.BuildSet()
	set.SetItem(at, "TimeAt")
	set.SetItem(ip, "IP")
	set.SetItem(mac, "MAC")
	set.SetItem(namely, "Namely")
	set.SetItem(formula, "Formula")
	InsertElm("Eventage", set)
	old := int(time.Now().Add(-time.Hour * 24 * 35).Unix())
	DB.Execute(fmt.Sprintf("DELETE FROM Eventage WHERE TimeAt<%d", old))
	if DebugLevel > 0 {
		who := IPToShow(ip)
		if mac != "" {
			who += " " + MACToShow(mac)
		}
		lik.SayInfo(fmt.Sprintf("%s %s: %s", formula, namely, who))
	}
}
