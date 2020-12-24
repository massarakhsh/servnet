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
	confDHCP()
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
		elm.Canonic = ""
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
				if elm.Canonic == "" {
					elm.Canonic = name
				}
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
	code += "option static-route-rfc code 121 = string;\n"
	code += "option static-route-win code 249 = string;\n"
	code += "option wpad code 252 = text;\n"
	code += "\n"
	code += "shared-network RPTP {\n"
	hosts := ""
	list_ip := confListIP()
	list_zone := DB.GetListElm("*", "IPZone", "(Roles&0x4)=0", "IP")
	for nz := 0; nz < list_zone.Count(); nz++ {
		if zone := list_zone.GetSet(nz); zone != nil {
			pic := "(192)(168)(\\d\\d\\d)(\\d\\d\\d)"
			if match := lik.RegExParse(zone.GetString("IP"), pic); match != nil {
				ip1 := lik.StrToInt(match[1])
				ip2 := lik.StrToInt(match[2])
				ip3 := lik.StrToInt(match[3])
				ipp := match[1] + match[2] + match[3]
				ip13 := fmt.Sprintf("%d.%d.%d", ip1, ip2, ip3)
				code += fmt.Sprintf("	subnet %s.0 netmask 255.255.255.0 {\n", ip13)
				code += fmt.Sprintf("		option ntp-servers 192.168.234.62;\n")
				code += fmt.Sprintf("		option time-servers 192.168.234.62;\n")
				code += fmt.Sprintf("		option domain-name-servers 192.168.234.62;\n")
				code += fmt.Sprintf("		option netbios-name-servers 192.168.234.62;\n")
				code += fmt.Sprintf("		option broadcast-address %s.255;\n", ip13)
				code += fmt.Sprintf("		option subnet-mask 255.255.255.0;\n")
				code += fmt.Sprintf("		option wpad \"http://192.168.234.62/wpad.dat\";\n")
				code += fmt.Sprintf("		option netbios-node-type 4;\n")
				code += fmt.Sprintf("		option routers %s.3;\n", ip13)
				if ip3 == 200 {
					code += fmt.Sprintf("		range %s.16 %s.62;\n", ip13, ip13)
				}
				if ip3 == 229 {
					option := confClassLess()
					code += fmt.Sprintf("		option static-route-rfc %s;\n", option)
					code += fmt.Sprintf("		option static-route-win %s;\n", option)
				}
				code += fmt.Sprintf("	}\n")
				for _,elm := range list_ip {
					if match := lik.RegExParse(elm.IP, ipp + "(\\d\\d\\d)"); match != nil {
						ip4 := lik.StrToInt(match[1])
						if ip4 > 0 && elm.MAC != "" && elm.Canonic != "" {
							hosts += fmt.Sprintf("host %s {\n", elm.Canonic)
							hosts += fmt.Sprintf("	hardware ethernet %s;\n", MACToShow(elm.MAC))
							hosts += fmt.Sprintf("	fixed-address %s.%d;\n", ip13, ip4)
							hosts += fmt.Sprintf("}\n")
						}
					}
				}
			}
		}
	}
	code += "}\n"
	code += hosts
	if confWrite("/etc/lik/dhcpd.conf", code) {
		if host,_ := os.Hostname(); strings.ToLower(host) == "root2" {
			if cmd := exec.Command("/etc/init.d/bind9 restart"); cmd != nil {
				cmd.Run()
			}
		}
	}
}

func confClassLess() string {
	//my $routes = {
	//	#'0.0.0.0/0'     => '10.10.124.10', # default route
	//'10.62.155.0/24'  => '192.168.229.3',
	//'192.168.0.0/16'  => '192.168.229.3',
	//};
	return "10:c0:a8:c0:a8:e5:03:18:0a:3e:9b:c0:a8:e5:03"
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