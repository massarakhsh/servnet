package baser

import (
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"time"
)

type Baser struct {
	task.Task
	LastTimeAlarm time.Time
}

func StartBaser() {
	go func() {
		iper := &Baser{ LastTimeAlarm: time.Now() }
		iper.Initialize("Baser", iper)
	}()
}

func (it *Baser) DoStep() {
	if lik.StrToInt(base.GetParm("LikSrvAlarm")) > 0 {
		base.SetParm("LikSrvAlarm", "0")
		base.DBNetUpdated = true
	} else if time.Now().Sub(it.LastTimeAlarm) > base.TimeoutFull {
		it.doAlarm()
	} else if base.DBNetUpdated && time.Now().Sub(it.LastTimeAlarm) > base.TimeoutAlarm {
		it.doAlarm()
	}
	it.SetPause(time.Millisecond * 500)
}

func (it *Baser) doAlarm() {
	it.LastTimeAlarm = time.Now()
	base.LoadTables()
	it.LastTimeAlarm = time.Now()
}

