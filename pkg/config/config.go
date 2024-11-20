package config

import (
        "fmt"
        "github.com/spf13/viper"
)

// LocalDomain represents a single domain and its IP mappings
type LocalDomain struct {
    Name string   `yaml:"name"`
    IPv4 []string `yaml:"ipv4"`
    IPv6 []string `yaml:"ipv6"`
}

type Config struct {
    Interfaces struct {
        LAN struct {
            Iface string `yaml:"iface"`
            IPv4  string `yaml:"ipv4"`
            IPv6  string `yaml:"ipv6"`
        } `yaml:"lan"`
        WAN string `yaml:"wan"`
    } `yaml:"interfaces"`
    DNS struct {
        Upstream struct {
            IPv4 []string `yaml:"ipv4"`
            IPv6 []string `yaml:"ipv6"`
        } `yaml:"upstream"`
        LocalDomains []LocalDomain `yaml:"local_domains"`
    } `yaml:"dns"`
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
        v.SetDefault("interfaces.lan.iface", "eth0")
        v.SetDefault("interfaces.wan", "eth1")

        // Default upstream DNS servers
        v.SetDefault("dns.upstream.ipv4", []string{"1.1.1.1", "8.8.8.8"})
        v.SetDefault("dns.upstream.ipv6", []string{"2606:4700:4700::1111", "2001:4860:4860::8888"})

        // Default local domain example
        v.SetDefault("dns.local_domains", []map[string]interface{}{
                {
                        "name": "router.local",
                        "ipv4": []string{"192.168.1.1"},
                        "ipv6": []string{"fd00::1"},
                },
                {
                        "name": "nas.local",
                        "ipv4": []string{"192.168.1.2"},
                        "ipv6": []string{"fd00::2"},
                },
        })
}

func (c *Config) Display() {
        fmt.Println("\nInterfaces:")
        fmt.Printf("  LAN: %s\n", c.Interfaces.LAN.Iface)
        fmt.Printf("    IPv4: %s\n", c.Interfaces.LAN.IPv4)
        fmt.Printf("    IPv6: %s\n", c.Interfaces.LAN.IPv6)
        fmt.Printf("  WAN: %s\n", c.Interfaces.WAN)

        fmt.Println("\nDNS Configuration:")
        fmt.Println("  Upstream DNS:")
        fmt.Printf("    IPv4: %v\n", c.DNS.Upstream.IPv4)
        fmt.Printf("    IPv6: %v\n", c.DNS.Upstream.IPv6)

        fmt.Println("  Local Domains:")
        for _, domain := range c.DNS.LocalDomains {
                fmt.Printf("    %s:\n", domain.Name)
                fmt.Printf("      IPv4: %v\n", domain.IPv4)
                fmt.Printf("      IPv6: %v\n", domain.IPv6)
        }
}
