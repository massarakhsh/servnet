#!/usr/bin/perl -CS

package LikSrv;
$BaseDir = "/var/lik/liksrv";
require($BaseDir."/liksrv.pm");

$Debug = (scalar(@ARGV)>0) ? $ARGV[0] : 0;

my ($host,$user,$pass,$base) = kukareku_rptp();
$MyB = lb_open("Host"=>$host,"User"=>$user,"Pass"=>$pass,"Base"=>$base);
if ($MyB) {
    Addr_Syn_Fun(0);
    Addr_Syn_Fun(1);
    Addr_Syn_Fun(2);
    Addr_Syn_Fun(3);
    lb_close($MyB);
}

sub Addr_Syn_Fun {
	my $fun = shift(@_);
	my $filename = "";
	if ($fun==0) { $filename = "/etc/dhcp/dhcpd.conf"; }
	elsif ($fun==1) { $filename = "/etc/bind/rptp.org.zone"; }
	elsif ($fun==2) { $filename = "/etc/bind/192.168.rev"; }
	elsif ($fun==3) { $filename = "/etc/iptables/gatelist.sh"; }
	my $dataold = "";
	if (open(my $fh, '<:encoding(UTF-8)', $filename)) {
		while (my $row = <$fh>) {
			$dataold .= $row;
		}
		close($fh);
	}
	my $datanew = "";
	if ($fun==0) { $datanew = Addr_Gen_DHCP(); }
	elsif ($fun==1) { $datanew = Addr_Gen_Direct(); }
	elsif ($fun==2) { $datanew = Addr_Gen_Reverse(); }
	elsif ($fun==3) { $datanew = Addr_Gen_IPTables(); }
	if ($datanew ne $dataold) {
		if (open(my $fh, '>', $filename)) {
			print $fh $datanew;
			close($fh);
			my $diag = "";
			if ($fun==0) {
				my @ans = `/etc/init.d/isc-dhcp-server restart`;
				$diag = "DHCP upgraded";
			}
			elsif ($fun==1) {
				my @ans = `/etc/init.d/bind9 reload`;
				$diag = "DNS forward upgraded";
			}
			elsif ($fun==2) {
				my @ans = `/etc/init.d/bind9 reload`;
				$diag = "DNS reverse upgraded";
			}
			elsif ($fun==3) {
				my @ans = `/etc/iptables/iptables.sh`;
				$diag = "IPTables upgraded";
			}
			if ($diag) {
				if ($Debug>=1) { printf("$diag\n"); }
			}
		}
	}
}

sub Addr_Gen_DHCP {
	my $list_ip = lb_query($MyB,"SELECT IP.IP,IP.MAC,Unit.Namely FROM IP LEFT JOIN Unit ON IP.SysUnit=Unit.SysNum".
							" WHERE IP.IP<>'' AND IP.MAC<>'' AND (IP.Roles&0x200000)=0");
	my $data = "";
	my %used = ();
	$data .= "max-lease-time 40000;\n";
	$data .= "default-lease-time 40000;\n";
	$data .= "authoritative;\n";
	$data .= "update-static-leases on;\n";
	$data .= "use-host-decl-names on;\n";
	$data .= "option domain-name \"rptp.org\";\n";
	$data .= "option static-route-rfc code 121 = string;\n";
	$data .= "option static-route-win code 249 = string;\n";
	$data .= "option wpad code 252 = text;\n";
	$data .= "\n";
	$data .= "shared-network RPTP {\n";
	my $hosts = "";
	my $dias = lb_query($MyB, "SELECT * FROM IPZone WHERE (Roles&0x4)=0 ORDER BY IP");
	foreach my $dia (@$dias) {
		my $ipdia = $$dia{'IP'};
		if ($ipdia =~ /192168(\d\d\d)000/) {
			my $ip3 = int($1)+0;
			my $pre3 = "192.168.".$ip3;
			if ($ip3!=200) {
				foreach my $elm (@$list_ip) {
					my $ip = IPToShow($$elm{'IP'});
					my $mac = MACToShow($$elm{'MAC'});
					if ($mac && $ip =~ /^192\.168\.$ip3\.(\d+)/) {
						my $ip4 = int($1)+0;
						if (!$used{$ip} && !$used{$mac}) {
							$used{$ip} = 1;
							$used{$mac} = 1;
							my $name = NameToNormal($$elm{'Namely'});
							if (!$name) { $name = "IP-$ip3-$ip4"; }
							for ($nix=0; ; $nix++) {
								my $namix = lc(($nix>0) ? $name."-".$nix : $name);
								if (!$used{$namix}) {
									$used{$namix} = 1;
									$name = $namix;
									last;
								}
							}
							if ($ip && $name && $mac) {
								$hosts .= "host $name {\n";
								$hosts .= "\thardware ethernet $mac;\n";
								$hosts .= "\tfixed-address $ip;\n";
								$hosts .= "}\n";
							}
						}
					}
				}
			}

			$data .= "\tsubnet $pre3.0 netmask 255.255.255.0 {\n";
			$data .= "\t\toption ntp-servers 192.168.234.62;\n";
			$data .= "\t\toption time-servers 192.168.234.62;\n";
			$data .= "\t\toption domain-name-servers 192.168.234.62;\n";
			$data .= "\t\toption netbios-name-servers 192.168.234.62;\n";
			$data .= "\t\toption broadcast-address $pre3.255;\n";
			$data .= "\t\toption subnet-mask 255.255.255.0;\n";
			$data .= "\t\toption wpad \"http://192.168.234.62/wpad.dat\";\n";
			$data .= "\t\toption netbios-node-type 4;\n";
			$data .= "\t\toption routers $pre3.3;\n";
			if ($ip3==200) {
				$data .= "\t\trange $pre3.16 $pre3.62;\n";
			}
			if ($ip3==229) {
				my $routes = {
					#'0.0.0.0/0'     => '10.10.124.10', # default route
					'10.62.155.0/24'  => '192.168.229.3',
					'192.168.0.0/16'  => '192.168.229.3',
				};

				my $option_value = make_classless_option($routes);
				$data .= "\t\toption static-route-rfc $option_value;\n";
				$data .= "\t\toption static-route-win $option_value;\n";
			}
			$data .= "\t}\n";
		}
	}
	$data .= "}\n";
	$data .= $hosts;
	return $data;
}

# see RFC 3442
sub make_classless_option {
	my $routes = shift;

	my @bytes = ();

	foreach my $destination (keys %{$routes}) {
		my ($net, $mask) = split '/', $destination;
		die "Bad netmask in $destination" unless $mask =~ /^\d\d?$/ && $mask >= 0 && $mask <= 32;
		push @bytes, $mask;

		my $significant_octets = int($mask / 8);
		my @octets = split /\./, $net;
		push @bytes, @octets[0 .. $significant_octets - 1];

		my @gw = split /\./, $routes->{$destination};
		die "Bad gateway " . $routes->{$destination} unless scalar @gw == 4;
		push @bytes, @gw;
	}

	return join(':', map { octet_to_hex($_) } @bytes);
}

sub octet_to_hex {
	my $octet = shift;
	die "Bad octet $octet" unless $octet =~ /^\d{1,3}$/ && $octet >= 0 && $octet <= 255;
	return sprintf('%02x', $octet);
}

sub Addr_Gen_Direct {
	my $list_ip = lb_query($MyB,"SELECT IP.IP,IP.Namely,Unit.Namely AS Host".
				" FROM IP LEFT JOIN Unit ON IP.SysUnit=Unit.SysNum".
				" WHERE IP.IP<>'' AND Unit.Namely<>'' AND Unit.Namely IS NOT NULL AND (IP.Roles&0x200)=0".
				" ORDER BY IP.IP,Unit.Namely");
	my $data = "";
	$data .= "\$TTL 38400	; 10 hours 40 minutes\n";
	$data .= "rptp.org.	IN	SOA	root.rptp.org. master.rptp.org. ( 1428401303 10800 3600 604800 38400 )\n";
	$data .= "rptp.org.	IN	NS	root.rptp.org.\n";
	$data .= "rptp.org.	IN	A	192.168.234.62\n";
	$data .= ";\n";
	$data .= "root.rptp.org.	IN	A	192.168.234.62\n";
	$data .= ";\n";
	foreach my $elm (@$list_ip) {
		my $ip = IPToShow($$elm{"IP"});
		if ($ip =~ /^\d+\.\d+\.\d+\.\d+$/) {
			my $host = lc($$elm{"Host"});
			my @names = ($host);
			my $namely = lc($$elm{"Namely"});
			if ($namely=~/^=(.+)$/) { push(@names,(split(/[\s,;]/,$1))); }
			foreach my $name (@names) {
				my $pref = NameToNormal($name);
				if ($name=~/^[\w\d\-\_]+$/ && $name!~/^IP-\d+-\d+/i && $data!~/^$pref\./m) {
					$data .= "$pref.rptp.org. IN A $ip\n";
				}
			}
		}
	}
	return $data;
}

sub Addr_Gen_Reverse {
	my $list_ip = lb_query($MyB,"SELECT IP.IP,Unit.Namely".
				" FROM IP LEFT JOIN Unit ON IP.SysUnit=Unit.SysNum".
				" WHERE IP.IP<>'' AND Unit.Namely<>'' AND Unit.Namely IS NOT NULL AND (IP.Roles&0x200)=0".
				" ORDER BY IP.IP,Unit.Namely");
	my $data = "";
	$data .= "\$TTL 38400	; 10 hours 40 minutes\n";
	$data .= "168.192.in-addr.arpa.	IN	SOA	root.rptp.org. master.rptp.org. ( 1330323963 10800 3600 604800 38400 )\n";
	$data .= "168.192.in-addr.arpa.	IN	NS	root.rptp.org.\n";
	$data .= ";\n";
	$data .= "1.0.168.192.in-addr.arpa.	IN	PTR	root.rptp.org.\n";
	$data .= "62.234.168.192.in-addr.arpa.	IN	PTR	root.rptp.org.\n";
	$data .= ";\n";
	foreach my $elm (@$list_ip) {
		my $ip = IPToShow($$elm{"IP"});
		if ($ip =~ /^192\.168\.(\d+)\.(\d+)$/) {
			my $ip3 = $1;
			my $ip4 = $2;
			my $namely = NameToNormal($$elm{"Namely"});
			if ($namely && $namely !~ /^IP-\d+-\d+/i) {
				if ($data !~ /^$ip4\.$ip3\.168\.192./m) {
					$data .= "$ip4.$ip3.168.192.in-addr.arpa. IN PTR $namely.rptp.org.\n";
				}
			}
		}
	}
	return $data;
}

sub Addr_Gen_IPTables {
	my $list_ip = lb_query($MyB,"SELECT IP.IP,IP.Roles FROM IP WHERE IP.IP<>'' ORDER BY IP.IP");
	my $data = "#!/bin/bash\n";
	foreach my $elm (@$list_ip) {
		my $roles = $$elm{"Roles"};
		if ($roles&0x8) {
			my $ip = IPToShow($$elm{"IP"});
			$data .= "/sbin/iptables -A FORWARD -s $ip -j ACCEPT\n";
			$data .= "/sbin/iptables -A FORWARD -d $ip -j ACCEPT\n";
			$data .= "/sbin/iptables -t nat -A POSTROUTING -s $ip".
					" -o eth1 -j SNAT --to-source 172.16.199.1\n";
		}
	}
	return $data;
}

sub NameToNormal {
	my $name = lc(shift(@_));
	if ($name =~ /^IP-\d+-\d+/i) { $name = ""; }
	$name =~ s/[^a-z0-9\-]//g; 
	return $name;
}

