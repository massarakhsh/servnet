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

type hostAddress struct {
	IP		string
	MAC		string
	Name	string
}

var hostList []*hostAddress
var hostMapIP map[string]*hostAddress
var hostListIP []*hostAddress
var hostMapMAC map[string]*hostAddress
var hostListMAC []*hostAddress
var hostMapName map[string]*hostAddress
var hostListName []*hostAddress

func Configurate() {
	confLoad()
	if confDirect("/etc/lik/rptp.org.zone") ||
		confReverse("/etc/lik/192.168.rev") {
		if host,_ := os.Hostname(); strings.ToLower(host) == "root2" {
			if cmd := exec.Command("/etc/init.d/bind9 restart"); cmd != nil {
				cmd.Run()
			}
		}
	}
	if confDHCP("/etc/lik/dhcpd.conf") {
		if host,_ := os.Hostname(); strings.ToLower(host) == "root2" {
			if cmd := exec.Command("/etc/init.d/isc-dhcp-server restart"); cmd != nil {
				cmd.Run()
			}
		}
	}
	if confGate("/etc/lik/gatelist.sh") {
		if host,_ := os.Hostname(); strings.ToLower(host) == "root2" {
			if cmd := exec.Command("/etc/iptables/iptables.sh"); cmd != nil {
				cmd.Run()
			}
		}
	}
}

func confLoad() {
	hostList = []*hostAddress{}
	hostMapIP = make(map[string]*hostAddress)
	hostMapMAC = make(map[string]*hostAddress)
	hostMapName = make(map[string]*hostAddress)
	list := confListIP()
	for _,elm := range list {
		elm.Host = ""
		if elm.SysNum > 0 && elm.IP > "" && (elm.Roles & 0x200) == 0 {	//	Первичный адрес
			if unit,_ := UnitMapSys[elm.SysUnit]; unit != nil {
				if name := confNameSymbols(unit.Namely); name != ""  {
					confLoadAdd(elm.IP, elm.MAC, name)
					elm.Host = name
				}
			}
			names := strings.Split(confNameSymbols(elm.Namely), ",")
			for _,name := range names {
				if name != "" {
					confLoadAdd(elm.IP, elm.MAC, name)
					if elm.Host == "" {
						elm.Host = name
					}
				}
			}
			if elm.Host == "" {
				confLoadAdd(elm.IP, elm.MAC, "")
			}
		}
	}
	hostListIP = []*hostAddress{}
	for _,host := range hostMapIP {
		hostListIP = append(hostListIP, host)
	}
	sort.SliceStable(hostListIP, func(i, j int) bool {
		return hostListIP[i].IP < hostListIP[j].IP
	})
	hostListMAC = []*hostAddress{}
	for _,host := range hostMapMAC {
		hostListMAC = append(hostListMAC, host)
	}
	sort.SliceStable(hostListMAC, func(i, j int) bool {
		return hostListMAC[i].MAC < hostListMAC[j].MAC
	})
	hostListName = []*hostAddress{}
	for _,host := range hostMapName {
		hostListName = append(hostListName, host)
	}
	sort.SliceStable(hostListName, func(i, j int) bool {
		return hostListName[i].Name < hostListName[j].Name
	})
}

func confLoadAdd(ip string, mac string, name string) {
	host := &hostAddress{IP: ip, MAC: mac, Name: name }
	hostList = append(hostList, host)
	if ip != "" {
		hostMapIP[ip] = host
	}
	if mac != "" {
		hostMapMAC[mac] = host
	}
	if ip != "" {
		hostMapName[name] = host
	}
}

func confDirect(namefile string) bool {
	code := "$TTL	38400\n"
	code += "rptp.org.	IN	SOA	root.rptp.org.	master.rptp.org. ( 1428401303 10800 3600 604800 38400 )\n"
	code += "rptp.org.	IN	NS	root.rptp.org.\n"
	code += "rptp.org.	IN	A	192.168.234.62\n"
	code += ";\n"
	for _,host := range hostListName {
		if host.Name != "" && host.IP != "" {
			code += fmt.Sprintf("%s.rptp.org.	IN	A	%s\n", host.Name, IPToShow(host.IP))
		}
	}
	return confWrite(namefile, code)
}

func confReverse(namefile string) bool {
	code := "$TTL	38400\n"
	code += "168.192.in-addr.arpa.	IN	NS	root.rptp.org.\n"
	code += ";\n"
	for _,host := range hostListIP {
		if match := lik.RegExParse(host.IP, "(192)(168)(\\d\\d\\d)(\\d\\d\\d)"); match != nil && host.Name != "" {
			ip3 := lik.StrToInt(match[3])
			ip4 := lik.StrToInt(match[4])
			code += fmt.Sprintf("%d.%d.168.192.in-addr.arpa.	IN	PTR	%s.rptp.org.\n", ip4, ip3, host.Name)
		}
	}
	return confWrite(namefile, code)
}

func confDHCP(namefile string) bool {
	code := "max-lease-time	40000;\n"
	code += "default-lease-time	40000;\n"
	code += "authoritative;\n"
	code += "update-static-leases on;\n"
	code += "use-host-decl-names on;\n"
	code += "option domain-name \"rptp.org\";\n"
	code += "option static-route-rfc code 121 = string;\n"
	code += "option static-route-win code 249 = string;\n"
	code += "option wpad code 252 = text;\n"
	code += "\n"
	code += "shared-network RPTP {\n"
	hosts := ""
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
				for _,host := range hostMapMAC {
					if match := lik.RegExParse(host.IP, ipp + "(\\d\\d\\d)"); match != nil {
						ip4 := lik.StrToInt(match[1])
						if ip4 > 0 && host.IP != "" && host.MAC != "" {
							name := host.Name
							if name == "" {
								name = "ip" + host.IP
							}
							hosts += fmt.Sprintf("host %s {\n", name)
							hosts += fmt.Sprintf("	hardware ethernet %s;\n", MACToShow(host.MAC))
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
	return confWrite("/etc/lik/dhcpd.conf", code)
}

func confClassLess() string {
	//my $routes = {
	//	#'0.0.0.0/0'     => '10.10.124.10', # default route
	//'10.62.155.0/24'  => '192.168.229.3',
	//'192.168.0.0/16'  => '192.168.229.3',
	//};
	return "10:c0:a8:c0:a8:e5:03:18:0a:3e:9b:c0:a8:e5:03"
}

func confGate(namefile string) bool {
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
	return confWrite(namefile, code)
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