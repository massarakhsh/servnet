package main

import (
	"fmt"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task/baser"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/massarakhsh/lik"
)

const pidFile = "var/servnet.pid"

func main() {
	lik.SetLevelInf()

	pidgo := getActiveProcess()
	base.HostPid = syscall.Getpid()

	if host,err := os.Hostname(); err == nil {
		base.HostName = strings.ToLower(host)
	}
	if base.HostName == "shaman" {
		base.HostServ = "192.168.234.62"
		base.HostVirtual = true
	}
	lik.SayError("System started on " + base.HostName)

	if !getArgs() {
		return
	}
	cmd := ""
	if base.HostSignal != "" {
		cmd = strings.ToUpper(base.HostSignal[:1])
	}
	if pidgo > 0 {
		if prc, err := os.FindProcess(pidgo); prc != nil && err == nil {
			if cmd == "S" {
				lik.SayWarning("Send stop process")
				prc.Signal(syscall.Signal(23))
			} else if cmd == "C" {
				lik.SayWarning("Send continue process")
				prc.Signal(syscall.Signal(25))
			} else {
				lik.SayWarning("Send term process")
				prc.Signal(syscall.SIGTERM)
			}
		} else {
			lik.SayWarning("Proceess not found")
		}
		setActiveProcess(0)
	}
	if cmd != "" {
		return
	}

	setActiveProcess(base.HostPid)
	base.HostChan = make(chan os.Signal, 1)
	signal.Notify(base.HostChan, syscall.SIGKILL, syscall.SIGTERM, syscall.Signal(23), syscall.Signal(25))

	go waitSignal()
	if !base.OpenDB(base.HostServ, base.HostBase, base.HostUser, base.HostPass) {
		return
	}
	base.WaitDB()
	baser.StartBaser()
	baser.StartPinger()
	baser.StartARP()

	for !base.IsStoping {
		time.Sleep(time.Second * 1)
	}
	time.Sleep(time.Second * 3)
	base.CloseDB()
	if getActiveProcess() == base.HostPid {
		setActiveProcess(0)
	}
}

func getArgs() bool {
	args, ok := lik.GetArgs(os.Args[1:])
	if val := args.GetString("serv"); val != "" {
		base.HostServ = val
	}
	if val := args.GetString("base"); val != "" {
		base.HostBase = val
	}
	if val := args.GetString("user"); val != "" {
		base.HostUser = val
	}
	if val := args.GetString("pass"); val != "" {
		base.HostPass = val
	}
	if val := args.GetInt("virtual"); val > 0 {
		base.HostVirtual = val > 0
	}
	if val := args.GetString("signal"); val != "" {
		base.HostSignal = val
	}
	if val := args.GetString("debug"); val != "" {
		base.DebugLevel = lik.StrToInt(val)
	}
	if !ok {
		fmt.Println("Usage: servnet [-key val | --key=val]...")
		fmt.Println("serv    - Database server")
		fmt.Println("base    - Database name")
		fmt.Println("user    - Database user")
		fmt.Println("pass    - Database pass")
	}
	return ok
}

//	Получить код активного процесса из PID-файла
func getActiveProcess() int {
	pid := 0
	if data, err := ioutil.ReadFile(pidFile); err == nil {
		pid = lik.StrToInt(string(data))
	}
	return pid
}

//	Записать код процесса в PID-файл
func setActiveProcess(pid int) {
	var data []byte
	if pid > 0 {
		data = []byte(lik.IntToStr(pid))
	}
	ioutil.WriteFile(pidFile, data, 0777)
}

//	Процесс ожидания и обработки сигналов
func waitSignal() {
	for {
		signal := <-base.HostChan
		if signal == syscall.Signal(23) {
			//repo.ToPause = true
		} else if signal == syscall.Signal(25) {
			//repo.ToPause = false
		} else {
			base.IsStoping = true
			break
		}
	}
}

