package base

import (
	"fmt"
	"github.com/massarakhsh/lik"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

func Configurate() {
	confDirect()
	confReverse()
	//confDHCP()
	confGate()
}

func confDirect() {
	code := "$TTL	38400\n"
	code += "rptp.org.	IN	SOA	root.rptp.org.	master.rptp.org. ( 1428401303 10800 3600 604800 38400 )\n"
	code += "rptp.org.	IN	NS	root.rptp.org.\n"
	code += "rptp.org.	IN	A	192.168.234.62\n"
	code += ";\n"
	code += "root.rptp.org.	IN	A	192.168.234.62\n"
	code += ";\n"
	list := confListIP()
	used := make(map[string]bool)
	for _,elm := range list {
		if elm.SysNum > 0 && elm.IP > "" && (elm.Roles & 0x200) == 0 {	//	Первичный адрес
			ip := IPToShow(elm.IP)
			var hosts []string
			if unit,_ := UnitMapSys[elm.SysUnit]; unit != nil {
				if name := confNameSymbols(unit.Namely); name != ""  && !used[name] {
					hosts = append(hosts, name)
					used[name] = true
				}
			}
			names := strings.Split(confNameSymbols(elm.Namely), ",")
			for _,name := range names {
				if name != ""  && !used[name] {
					hosts = append(hosts, name)
					used[name] = true
				}
			}
			if len(hosts) == 0 {
				if name := strings.Replace(ip, ".", "-", -1); name != ""  && !used[name] {
					hosts = append(hosts, name)
					used[name] = true
				}
			}
			for _,name := range hosts {
				code += fmt.Sprintf("%s.rptp.org.	IN	A	%s\n", name, ip)
			}
		}
	}
	if confWrite("/etc/lik/rptp.org.zone", code) {
		if host,_ := os.Hostname(); strings.ToLower(host) == "root2" {
			if cmd := exec.Command("/etc/init.d/bind9 restart"); cmd != nil {
				cmd.Run()
			}
		}
	}
}

func confReverse() {
	code := "168.192.in-addr.arpa.	IN	NS	root.rptp.org.\n"
	code += ";\n"
	code += "62.234.168.192.in-addr.arpa.	IN	PTR	root.rptp.org\n"
	code += ";\n"
	list := confListIP()
	used := make(map[string]bool)
	for _,elm := range list {
		if match := lik.RegExParse(elm.IP, "192168(\\d\\d\\d)(\\d\\d\\d)"); elm.SysNum > 0 && match != nil {	//	Первичный адрес
			ip3 := lik.StrToInt(match[1])
			ip4 := lik.StrToInt(match[2])
			if unit,_ := UnitMapSys[elm.SysUnit]; unit != nil {
				ipi := fmt.Sprintf("%d.%d", ip4, ip3)
				if name := confNameSymbols(unit.Namely); name != ""  && !used[ipi] {
					code += fmt.Sprintf("%s.168.192.in-addr.arpa.	IN	PTR	%s.rptp.org.\n", ipi, name)
					used[ipi] = true
				}
			}
		}
	}
	if confWrite("/etc/lik/192.168.234.62.zone", code) {
		if host,_ := os.Hostname(); strings.ToLower(host) == "root2" {
			if cmd := exec.Command("/etc/init.d/bind9 restart"); cmd != nil {
				cmd.Run()
			}
		}
	}
}

func confDHCP() {
	code := "max-lease-time	40000;\n"
	code += "default-lease-time	40000;\n"
	code += "authoritative;\n"
	code += "update-status-leases on;\n"
	code += "use-host-decl-names on;\n"
	code += "option domain-name \"rptp.org\";\n"
	list := confListIP()
	used := make(map[string]bool)
	for _,elm := range list {
		if elm.SysNum > 0 && elm.IP > "" && (elm.Roles & 0x200) == 0 {	//	Первичный адрес
			ip := IPToShow(elm.IP)
			var hosts []string
			if unit,_ := UnitMapSys[elm.SysUnit]; unit != nil {
				if name := confNameSymbols(unit.Namely); name != ""  && !used[name] {
					hosts = append(hosts, name)
					used[name] = true
				}
			}
			names := strings.Split(confNameSymbols(elm.Namely), ",")
			for _,name := range names {
				if name != ""  && !used[name] {
					hosts = append(hosts, name)
					used[name] = true
				}
			}
			if len(hosts) == 0 {
				if name := strings.Replace(ip, ".", "-", -1); name != ""  && !used[name] {
					hosts = append(hosts, name)
					used[name] = true
				}
			}
			for _,name := range hosts {
				code += fmt.Sprintf("%s.rptp.org.	IN	A	%s\n", name, ip)
			}
		}
	}
	if confWrite("/etc/lik/rptp.org.zone", code) {
		if host,_ := os.Hostname(); strings.ToLower(host) == "root2" {
			if cmd := exec.Command("/etc/init.d/bind9 restart"); cmd != nil {
				cmd.Run()
			}
		}
	}
}

func confGate() {
	code := "#!/bin/bash\n"
	list := confListIP()
	for _,elm := range list {
		if elm.IP > "" && (elm.Roles & 0x8) != 0 {	//	Шлюз
			ip := IPToShow(elm.IP)
			code += fmt.Sprintf("/sbin/iptables -A FORWARD -s %s -j ACCEPT\n", ip)
			code += fmt.Sprintf("/sbin/iptables -A FORWARD -d %s -j ACCEPT\n", ip)
			code += fmt.Sprintf("/sbin/iptables -t nat -A POSTROUTING -s %s -o eth1 -j SNAT --to-source 172.16.199.1\n", ip)
		}
	}
	if confWrite("/etc/lik/gatelist.sh", code) {
		if host,_ := os.Hostname(); strings.ToLower(host) == "root2" {
			if cmd := exec.Command("/etc/iptables/iptables.sh"); cmd != nil {
				cmd.Run()
			}
		}
	}
}

func confNameSymbols(name string) string {
	name = strings.ToLower(name)
	name = lik.Transliterate(name)
	name = regexp.MustCompile("[^0-9a-z\\-\\_]").ReplaceAllString(name, "-")
	return name
}

func confListIP() []*ElmIP {
	var ips []*ElmIP
	for _,elm := range IPMapSys {
		ips = append(ips, elm)
	}
	sort.SliceStable(ips, func(i, j int) bool {
		return ips[i].IP < ips[j].IP
	})
	return ips
}

func confWrite(namefile string, code string) bool {
	if file, err := os.Open(namefile); err == nil {
		oldcode := ""
		buf := make([]byte, 4096)
		for {
			if n, err := file.Read(buf); err == nil {
				oldcode += string(buf[:n])
			} else {
				break
			}
		}
		file.Close()
		if oldcode == code {
			return false
		}
	}
	if file, err := os.Create(namefile); err == nil {
		file.WriteString(code)
		file.Close()
		return true
	}
	return false
}