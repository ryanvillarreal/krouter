package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
    Interfaces struct {
        LAN struct {
            Iface string `yaml:"iface"`
            IPv4  string `yaml:"ipv4"`
            IPv6  string `yaml:"ipv6"`
        } `yaml:"lan"`
        WAN string `yaml:"wan"`
    } `yaml:"interfaces"`
    Configs struct {
        CDNS_CONFIG  string `yaml:"coredns_config"`
        CDHCP_CONFIG string `yaml:"coredhcp_config"`
    } `yaml:"configs"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()
	
	setDefaults(v)
	
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yml")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("interfaces.lan", "eth0")
	v.SetDefault("interfaces.wan", "eth1")
	
	v.SetDefault("dns.upstream.ipv4", []string{"1.1.1.1", "8.8.8.8"})
	v.SetDefault("dns.upstream.ipv6", []string{"2606:4700:4700::1111", "2001:4860:4860::8888"})
	
	v.SetDefault("dns.local_domains", []map[string]interface{}{
		{
			"name": "router.local",
			"ipv4": []string{"192.168.1.1"},
			"ipv6": []string{"fd00::1"},
		},
	})
	
	v.SetDefault("dhcp.ipv4.ranges", []map[string]interface{}{
		{
			"start": "192.168.1.100",
			"end":   "192.168.1.200",
		},
	})
	v.SetDefault("dhcp.ipv4.lease_time", "24h")
	v.SetDefault("dhcp.ipv4.gateway", "192.168.1.1")
	v.SetDefault("dhcp.ipv4.dns", []string{"192.168.1.1"})
	
	v.SetDefault("dhcp.ipv6.prefix", "fd00::")
	v.SetDefault("dhcp.ipv6.prefix_len", 64)
	v.SetDefault("dhcp.ipv6.lease_time", "24h")
	v.SetDefault("dhcp.ipv6.dns", []string{"fd00::1"})
}

func (c *Config) Display() {
	fmt.Println("\nInterfaces:")
	fmt.Printf("  LAN: %s\n  WAN: %s\n", c.Interfaces.LAN, c.Interfaces.WAN)

	fmt.Println("\nDNS Configuration:")
	fmt.Println("  Local Domains:")
	
  //for _, domain := range c.DNS.LocalDomains {
	//	fmt.Printf("    %s:\n", domain.Name)
	//	fmt.Printf("      IPv4: %v\n", domain.IPv4)
	//	fmt.Printf("      IPv6: %v\n", domain.IPv6)
	//}
	
	fmt.Println("  Upstream DNS:")
	//fmt.Printf("    IPv4: %v\n", c.DNS.Upstream.IPv4)
	//fmt.Printf("    IPv6: %v\n", c.DNS.Upstream.IPv6)

	fmt.Println("\nDHCP Configuration:")
	fmt.Println("  IPv4:")
	//fmt.Printf("    Gateway: %s\n", c.DHCP.IPv4.Gateway)
	//fmt.Printf("    DNS Servers: %v\n", c.DHCP.IPv4.DNS)
	//fmt.Printf("    Lease Time: %s\n", c.DHCP.IPv4.LeaseTime)
	//fmt.Println("    Ranges:")
	//for _, r := range c.DHCP.IPv4.Ranges {
	//	fmt.Printf("      %s - %s\n", r.Start, r.End)
	//}

	fmt.Println("  IPv6:")
	//fmt.Printf("    Prefix: %s/%d\n", c.DHCP.IPv6.Prefix, c.DHCP.IPv6.PrefixLen)
	//fmt.Printf("    DNS Servers: %v\n", c.DHCP.IPv6.DNS)
	//fmt.Printf("    Lease Time: %s\n", c.DHCP.IPv6.LeaseTime)
}

