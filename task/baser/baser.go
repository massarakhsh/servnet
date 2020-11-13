package baser

import (
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"time"
)

type Baser struct {
	task.Task
	OnAlarm   bool
	OffAlarm  bool
	TimeAlarm time.Time
}

func StartBaser() {
	go func() {
		iper := &Baser{}
		iper.Initialize("Baser", iper)
	}()
}

func (it *Baser) DoStep() {
	on := lik.StrToInt(base.GetParm("LikSrvAlarm")) > 0
	if on {
		it.OnAlarm = true
		it.OffAlarm = false
	} else if it.OnAlarm {
		it.OnAlarm = false
		it.OffAlarm = true
		it.TimeAlarm = time.Now()
	} else if it.OffAlarm && time.Now().Sub(it.TimeAlarm) > time.Second*5 {
		it.OffAlarm = false
		lik.SayInfo("ALARM!")
		//base.LoadTables()
	}
	it.SetPause(time.Millisecond * 500)
}
