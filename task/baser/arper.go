package baser

import (
	"bytes"
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/base"
	"github.com/massarakhsh/servnet/task"
	"github.com/mostlygeek/arp"
	"golang.org/x/crypto/ssh"
	"log"
	"time"
)

type ARPer struct {
	task.Task
}

func StartARP() {
	go func() {
		arper := &ARPer{}
		arper.Initialize("ARPer", arper)
	}()
}

func (it *ARPer) DoStep() {
	//CallRouter()
	if table := arp.Table(); table != nil {
		base.LockDB()
		for ip, ipa := range table {
			mac := ""
			if match := lik.RegExParse(ipa, "(\\w\\w:\\w\\w:\\w\\w:\\w\\w:\\w\\w:\\w\\w)"); match != nil {
				mac = base.MACFromShow(match[1])
				if mac != "000000000000" && mac != "ffffffffffff" {
					//fmt.Printf("%s : %s\n", ip, base.MACToShow(mac))
					base.SetPingOnline(base.IPFromShow(ip), mac)
				}
			}
		}
		base.UnlockDB()
	}
	it.SetPause(time.Second * 15)
}

func CallRouter() {
	lik.SayInfo("SSH")
	var hostKey ssh.PublicKey
	config := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			ssh.Password("gamilto!&"),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	client, err := ssh.Dial("tcp", "192.168.234.3:22", config)
	if err != nil {
		fmt.Print("Failed to dial: ", err)
		return
	}
	lik.SayInfo("Ok")
	defer client.Close()
	if true {
		return
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("/usr/bin/whoami"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())
}
