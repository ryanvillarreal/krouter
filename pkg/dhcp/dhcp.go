package dhcp

//TODO - overload the coredhcp logging to use our own
import (
    "context"
    "strings"
    "log"
    "fmt"
    "sync"
    "net"

    cd_config "github.com/coredhcp/coredhcp/config"
    cd_server"github.com/coredhcp/coredhcp/server"
    "github.com/coredhcp/coredhcp/plugins"
	  pl_autoconfigure "github.com/coredhcp/coredhcp/plugins/autoconfigure"
	  pl_dns "github.com/coredhcp/coredhcp/plugins/dns"
	  pl_file "github.com/coredhcp/coredhcp/plugins/file"
	  pl_ipv6only "github.com/coredhcp/coredhcp/plugins/ipv6only"
	  pl_leasetime "github.com/coredhcp/coredhcp/plugins/leasetime"
  	pl_mtu "github.com/coredhcp/coredhcp/plugins/mtu"
	  pl_nbp "github.com/coredhcp/coredhcp/plugins/nbp"
	  pl_netmask "github.com/coredhcp/coredhcp/plugins/netmask"
	  pl_prefix "github.com/coredhcp/coredhcp/plugins/prefix"
	  pl_range "github.com/coredhcp/coredhcp/plugins/range"
	  pl_router "github.com/coredhcp/coredhcp/plugins/router"
	  pl_searchdomains "github.com/coredhcp/coredhcp/plugins/searchdomains"
	  pl_serverid "github.com/coredhcp/coredhcp/plugins/serverid"
	  pl_sleep "github.com/coredhcp/coredhcp/plugins/sleep"
	  pl_staticroute "github.com/coredhcp/coredhcp/plugins/staticroute"


    krouter "github.com/ryanvillarreal/krouter/pkg/config"
)

var desiredPlugins = []*plugins.Plugin{
	&pl_autoconfigure.Plugin,
	&pl_dns.Plugin,
	&pl_file.Plugin,
	&pl_ipv6only.Plugin,
	&pl_leasetime.Plugin,
	&pl_mtu.Plugin,
	&pl_nbp.Plugin,
	&pl_netmask.Plugin,
	&pl_prefix.Plugin,
	&pl_range.Plugin,
	&pl_router.Plugin,
	&pl_searchdomains.Plugin,
	&pl_serverid.Plugin,
	&pl_sleep.Plugin,
	&pl_staticroute.Plugin,
}

type Service struct {
    ctx      context.Context
    cancel   context.CancelFunc
    wg       sync.WaitGroup
   // our config 
    cfg      *krouter.Config
    // coredhcp server
    servers  *cd_server.Servers
    errChan  chan error
}

func NewDHCPService(cfg *krouter.Config) *Service {
    ctx, cancel := context.WithCancel(context.Background())
    return &Service{
        ctx:     ctx,
        cancel:  cancel,
        cfg:     cfg,
        errChan: make(chan error, 1),
    }
}

// formatDNSAddress formats an IP address for DNS configuration
func formatDNSAddress(addr string) string {
    ip := stripNetmask(addr)
    return fmt.Sprintf("%s:53", ip)
}

// stripNetmask removes CIDR notation from IP address
func stripNetmask(addr string) string {
    if idx := strings.Index(addr, "/"); idx != -1 {
        return addr[:idx]
    }
    return addr
}


func getMACAddress(ifaceName string) (string, error) {
    iface, err := net.InterfaceByName(ifaceName)
    if err != nil {
        return "", fmt.Errorf("failed to get interface: %w", err)
    }
    return iface.HardwareAddr.String(), nil
}

func (s *Service) buildCoreDHCPConfig() *cd_config.Config {
    
    mac, err := getMACAddress(s.cfg.Interfaces.LAN.Iface)
    if err != nil {
        return nil
    }

    conf := cd_config.New()
    
    // Configure DHCPv6 Server
    fmt.Println("Configuring dhcpv6 server")

    // Format DNS addresses
    dnsv6 := formatDNSAddress(s.cfg.Interfaces.LAN.IPv6)
    dnsv4 := formatDNSAddress(s.cfg.Interfaces.LAN.IPv4)

    fmt.Println(dnsv4)
    // register plugins
	  for _, plugin := range desiredPlugins {
		  if err := plugins.RegisterPlugin(plugin); err != nil {
			  log.Fatalf("Failed to register plugin '%s': %v", plugin.Name, err)
		  }
	  }
    conf.Server6 = &cd_config.ServerConfig{
        Addresses: []net.UDPAddr{{
            IP:   net.IPv6unspecified,
            Port: 547,
            Zone: s.cfg.Interfaces.LAN.Iface,
        }},
        Plugins: []cd_config.PluginConfig{
            {
                Name: "server_id",
                Args: []string{"LL", mac}, // Proper DUID-LL format
            },
            {
                // dns: <resolver IP> <.. resolver IPs>
                Name: "dns",
                Args: []string{dnsv6},
            },
        },
    }
    
    fmt.Println("Configuring dhcpv4 server")
    // Configure DHCPv4 Server
    conf.Server4 = &cd_config.ServerConfig{
        Addresses: []net.UDPAddr{{
            IP:   net.IPv4zero,
            Port: 67,
            Zone: s.cfg.Interfaces.LAN.Iface,
        }},
        Plugins: []cd_config.PluginConfig{
            {
                Name: "lease_time",
                Args: []string{"3600s"},
            },
            {
                Name: "server_id",
                Args: []string{stripNetmask(s.cfg.Interfaces.LAN.IPv4)},
            },
            {
                Name: "dns",
                Args: []string{stripNetmask(s.cfg.Interfaces.LAN.IPv4)},
            },
            {
                Name: "router",
                Args: []string{stripNetmask(s.cfg.Interfaces.LAN.IPv4)},
            },
            {
                Name: "netmask",
                Args: []string{"255.255.255.0"},
            },
            {
                Name: "range",
                Args: []string{
                    "leases4.txt",
                    "192.168.52.10",
                    "192.168.52.250",
                    "3600s",
                },
            },
        },
    }
    return conf
}

func (s *Service) Start() error {
    dhcpConfig := s.buildCoreDHCPConfig()
    fmt.Println("launching dhcp server")
    servers, err := cd_server.Start(dhcpConfig)
    if err != nil {
        return fmt.Errorf("failed to start DHCP server: %w", err)
    }
    fmt.Println("successful")
    s.servers = servers

    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        if err := servers.Wait(); err != nil {
            select {
            case s.errChan <- err:
            default:
            }
        }
    }()

    return nil
}

func (s *Service) Stop() {
    if s.servers != nil {
        s.servers.Close()
    }
    s.cancel()
    s.wg.Wait()
}

func (s *Service) Errors() <-chan error {
    return s.errChan
}
