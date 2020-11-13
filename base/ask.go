package base

import (
	"math/rand"
	"sync"
	"time"
)

type ElmAsk struct {
	IP      string
	Online  bool
	At      time.Time
	ElmNext *ElmAsk
	ElmPred *ElmAsk
}

var AskSync sync.Mutex
var AskFirst *ElmAsk
var AskNext *ElmAsk
var AskLast *ElmAsk
var AskMap map[string]*ElmAsk

func InitAsk() {
	AskFirst = nil
	AskNext = nil
	AskLast = nil
	AskMap = make(map[string]*ElmAsk)
}

func AddAsk(ip string, on bool) *ElmAsk {
	AskSync.Lock()
	pit, _ := AskMap[ip]
	if pit == nil {
		pit = &ElmAsk{IP: ip, Online: on}
		AskMap[ip] = pit
		askInsert(pit, 5)
	} else if on && !pit.Online {
		pit.Online = true
		askMove(pit, 5)
	}
	AskSync.Unlock()
	return pit
}

func AskPingDelay() time.Duration {
	AskSync.Lock()
	delay := time.Hour
	if AskNext != nil {
		delay = AskNext.At.Sub(time.Now())
	}
	AskSync.Unlock()
	return delay
}

func AskPingPop() *ElmAsk {
	AskSync.Lock()
	pit := AskNext
	if AskNext != nil {
		AskNext = AskNext.ElmNext
	}
	AskSync.Unlock()
	return pit
}

func AskPingPush(pit *ElmAsk) {
	AskSync.Lock()
	sec := 15 + rand.Intn(10)
	if !pit.Online {
		sec += 180 + rand.Intn(60)
	}
	askMove(pit, sec)
	AskSync.Unlock()
}

func askMove(pit *ElmAsk, sec int) {
	askExtract(pit)
	askInsert(pit, sec)
}

func askExtract(pit *ElmAsk) {
	if AskNext == pit {
		AskNext = pit.ElmNext
	}
	if pit.ElmPred != nil {
		pit.ElmPred.ElmNext = pit.ElmNext
	} else {
		AskFirst = pit.ElmNext
	}
	if pit.ElmNext != nil {
		pit.ElmNext.ElmPred = pit.ElmPred
	} else {
		AskLast = pit.ElmPred
	}
}

func askInsert(pit *ElmAsk, sec int) {
	pit.At = time.Now().Add(time.Second * time.Duration(sec))
	pred := AskLast
	next := AskNext
	if next != nil {
		pred = next.ElmPred
	}
	for next != nil && !next.At.After(pit.At) {
		pred = next
		next = next.ElmNext
	}
	pit.ElmPred = pred
	if pred != nil {
		pred.ElmNext = pit
	} else {
		AskFirst = pit
	}
	pit.ElmNext = next
	if next != nil {
		next.ElmPred = pit
	} else {
		AskLast = pit
	}
	if AskNext == next {
		AskNext = pit
	}
}
