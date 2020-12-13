package baser

import (
	"fmt"
	"github.com/go-ping/ping"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"os/exec"
	"time"
)

type Pinger struct {
	task.Task
}

func StartPinger() {
	go func() {
		pinger := &Pinger{}
		pinger.Initialize("Pinger", pinger)
	}()
}

func (it *Pinger) DoStep() {
	if base.AskPingDelay() <= 0 {
		if pit := base.AskPingPop(); pit != nil {
			//if runtime.GOOS == "windows" {
			go it.pingICMP(pit)
			dura := base.AskPingDelay()
			if dura < time.Millisecond {
				dura = time.Millisecond
			} else if dura > time.Second {
				dura = time.Second
			}
			it.SetPause(dura)
		}
	}
}

func (it *Pinger) pingExec(pit *base.ElmAsk) {
	if out, err := exec.Command("ping", "-n", "2", "-w", "250", base.IPToShow(pit.IP)).Output(); err != nil {
		fmt.Println(err)
	} else {
		outs := string(out)
		it.pingSetOnline(pit, lik.RegExCompare(outs, "TTL="))
	}
	base.AskPingPush(pit)
}

func (it *Pinger) pingICMP(pit *base.ElmAsk) {
	pinger, err := ping.NewPinger(base.IPToShow(pit.IP))
	if err != nil {
		return
	}
	pinger.Count = 3
	pinger.Timeout = time.Second * 1
	pinger.SetPrivileged(true)
	if err := pinger.Run(); err != nil {
		return
	}
	stats := pinger.Statistics()
	it.pingSetOnline(pit, stats.PacketsRecv > 0)
	if base.DebugLevel > 1 {
		text := fmt.Sprintf("Ping %s", base.IPToShow(pit.IP))
		diff := int(pit.At.Sub(time.Now()).Seconds())
		if diff < -2 || diff > 2 {
			text += fmt.Sprintf(" [%d]", diff)
		}
		if !pit.Online {
			text += " OFF"
		}
		lik.SayInfo(text)
	}
	base.AskPingPush(pit)
}

func (it *Pinger) pingSetOnline(pit *base.ElmAsk, on bool) {
	base.LockDB()
	if on {
		pit.Online = true
		base.SetIPOnline(pit.IP)
	} else {
		pit.Online = false
		base.SetIPOffline(pit.IP)
	}
	base.UnlockDB()
}
