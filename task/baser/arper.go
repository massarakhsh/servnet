package baser

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/task"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
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
	CallRouter()
	/*if table := arp.Table(); table != nil {
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
	}*/
	it.SetPause(time.Second * 15)
}

func CallRouter() {
	lik.SayInfo("SSH")
	key, err := ioutil.ReadFile("root.opn")
	if err != nil {
		fmt.Println("BadRead: ", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Println("BadParse: ", err)
	}
	//var hostKey ssh.PublicKey
	config := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	client, err := ssh.Dial("tcp", "192.168.0.3:22", config)
	if err != nil {
		fmt.Println("Failed to dial: ", err)
		return
	}
	defer client.Close()
	lik.SayInfo("Ok")
}

