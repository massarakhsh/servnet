package main

import (
	"bytes"
	"fmt"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task/api"
	"github.com/massarakhsh/servnet/task/baser"
	"io/ioutil"
	"os"
	"os/exec"
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
		base.ConfServ = "192.168.234.62"
		//base.ConfVirtual = true
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
		if prc := findActiveProcess(pidgo); prc != nil {
			if cmd == "S" {
				lik.SayWarning("Send stop process")
				prc.Signal(syscall.SIGTERM)
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
	if !base.OpenDB(base.ConfServ, base.ConfBase, base.ConfUser, base.ConfPass) {
		return
	}
	base.WaitDB()
	baser.StartBaser()
	baser.StartPinger()
	baser.StartARP()
	if base.ConfPort > 0 {
		api.StartAPI()
	}
	go func() {
		time.Sleep(time.Second * 10)
		//base.IsStoping = true
	}()

	for !base.IsStoping {
		time.Sleep(time.Second * 1)
	}
	time.Sleep(time.Second * 3)
	base.CloseDB()
	if getActiveProcess() == base.HostPid {
		setActiveProcess(0)
	}
	lik.SayError("Done on " + base.HostName)
}

func getArgs() bool {
	args, ok := lik.GetArgs(os.Args[1:])
	if val := args.GetString("serv"); val != "" {
		base.ConfServ = val
	}
	if val := args.GetString("base"); val != "" {
		base.ConfBase = val
	}
	if val := args.GetString("user"); val != "" {
		base.ConfUser = val
	}
	if val := args.GetString("pass"); val != "" {
		base.ConfPass = val
	}
	if val := args.GetInt("virtual"); val > 0 {
		base.ConfVirtual = val > 0
	}
	if val := args.GetInt("port"); val > 0 {
		base.ConfPort = val
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
		if match := lik.RegExParse(string(data), "(\\d+)"); match != nil {
			pid = lik.StrToInt(match[1])
		}
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

func findActiveProcess(pid int) *os.Process {
	var prc *os.Process
	spid := lik.IntToStr(pid)
	if exe := exec.Command("ps", "--pid=" + spid); exe != nil {
		var out bytes.Buffer
		exe.Stdout = &out
		exe.Run()
		if answer := out.String(); strings.Contains(answer, spid) {
			if pc, err := os.FindProcess(pid); pc != nil && err == nil {
				prc = pc
			}
		}
	}
	return prc
}

//	Процесс ожидания и обработки сигналов
func waitSignal() {
	for {
		signal := <-base.HostChan
		if signal == syscall.Signal(23) {
			//base.IsStoping = true
		} else if signal == syscall.Signal(25) {
			//repo.ToPause = false
		} else if signal == syscall.Signal(3) || signal == syscall.Signal(9) || signal == syscall.Signal(15)  {
			lik.SayWarning(fmt.Sprintf("Signal %d", signal))
			base.IsStoping = true
			break
		}
	}
}

