package dhcp

import (
    "context"
    "fmt"
    "sync"
    "net"

    cd_config "github.com/coredhcp/coredhcp/config"
    "github.com/coredhcp/coredhcp/server"
    krouter "github.com/ryanvillarreal/krouter/pkg/config"
)

type Service struct {
    servers  *server.Servers
    ctx      context.Context
    cancel   context.CancelFunc
    wg       sync.WaitGroup
    cfg      *krouter.Config
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

func (s *Service) buildCoreDHCPConfig() *cd_config.Config {
    conf := cd_config.New()
    
    // Configure DHCPv6 Server
    conf.Server6 = &cd_config.ServerConfig{
        Addresses: []net.UDPAddr{{
            IP:   net.IPv6unspecified,
            Port: 547,
            Zone: s.cfg.Interfaces.LAN.Iface,
        }},
        Plugins: []cd_config.PluginConfig{
            {
                Name: "server_id",
                Args: []string{fmt.Sprintf("LL %s", s.cfg.Interfaces.LAN.Iface)},
            },
            {
                Name: "file",
                Args: []string{"leases6.txt"},
            },
            {
                Name: "dns",
                Args: []string{s.cfg.Interfaces.LAN.IPv6},
            },
        },
    }

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
                Args: []string{s.cfg.Interfaces.LAN.IPv4},
            },
            {
                Name: "dns",
                Args: []string{s.cfg.Interfaces.LAN.IPv4},
            },
            {
                Name: "router",
                Args: []string{s.cfg.Interfaces.LAN.IPv4},
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
    servers, err := server.Start(dhcpConfig)
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
