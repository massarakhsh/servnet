package baser

import (
	"bytes"
	"fmt"
	"github.com/massarakhsh/lik"
	"github.com/massarakhsh/servnet/task"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
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
	if touch := Open("192.168.0.3:22", "admin", "", "root.opn"); touch != nil {
	}
}

type LikSSH struct {
	Server		string
	ItClient	*ssh.Client
	ItSession	*ssh.Session
}

func Open(server string, user string, password string, keyfile string) *LikSSH {
	header := &LikSSH{ Server: server }
	var auth []ssh.AuthMethod
	if password != "" {
		auth = append(auth, ssh.Password(password))
	}
	if keyfile != "" {
		key, err := ioutil.ReadFile(keyfile)
		if err != nil {
			return nil
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", "192.168.0.3:22", config)
	if err != nil {
		fmt.Println("Failed to dial: ", err)
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	return header
	return nil
}

func CallRouter2() {
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("interface bridge host print without-paging"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())

	lik.SayInfo("Ok")
}

