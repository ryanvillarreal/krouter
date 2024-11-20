package router

import (
        "context"
        "fmt"
        "log"
        "sync"
        "sync/atomic"
        "time"

        "github.com/ryanvillarreal/krouter/pkg/config"
        "github.com/ryanvillarreal/krouter/pkg/dhcp"
        "github.com/ryanvillarreal/krouter/pkg/dns"
)

type Status struct {
        healthy atomic.Bool
        lastCheck time.Time
}

type Service struct {
        // context objs
        cfg     *config.Config
        ctx     context.Context
        cancel  context.CancelFunc
        wg      sync.WaitGroup
        // info objs
        status  Status
        ticker  *time.Ticker
        // key objs
        ifManager *InterfaceManager
        dhcp      *dhcp.Service
        dns       *dns.DNSProxy
}

func New(cfg *config.Config) (*Service, error) {
        ctx, cancel := context.WithCancel(context.Background())
        s := &Service{
                cfg:    cfg,
                ctx:    ctx,
                cancel: cancel,
                ticker: time.NewTicker(5 * time.Second),
                ifManager: NewInterfaceManager(cfg),
                dhcp:   dhcp.NewDHCPService(cfg),
                dns:    dns.NewDNSProxy(cfg),
        }

        s.status.healthy.Store(true)
        return s, nil
}

func (s *Service) Start() error {
        log.Printf("Starting router service on interfaces \nLAN: %s \nWAN: %s",
                s.cfg.Interfaces.LAN.Iface, s.cfg.Interfaces.WAN)

        // Start interface configuration
        if s.ifManager == nil {
                return fmt.Errorf("interface manager initialization failed")
        }

        // Start DHCP service
        if err := s.dhcp.Start(); err != nil {
                return fmt.Errorf("failed to start DHCP service: %w", err)
        }

        // Start DNS service
        if err := s.dns.Start(); err != nil {
                s.dhcp.Stop() // cleanup on failure
                return fmt.Errorf("failed to start DNS service: %w", err)
        }

        s.wg.Add(1)
        go func() {
                defer s.wg.Done()
                s.run()
        }()

        // Monitor services for errors
        go s.monitorServices()

        return nil
}

func (s *Service) monitorServices() {
        for {
                select {
                case err := <-s.dhcp.Errors():
                        log.Printf("DHCP service error: %v", err)
                        s.status.healthy.Store(false)
                case err := <-s.dns.Errors():
                        log.Printf("DNS service error: %v", err)
                        s.status.healthy.Store(false)
                case <-s.ctx.Done():
                        return
                }
        }
}

func (s *Service) Stop() {
        s.ticker.Stop()
        s.dns.Stop()
        s.dhcp.Stop()
        s.cancel()
        s.wg.Wait()
        log.Println("Router service stopped")
}

func (s *Service) IsHealthy() bool {
        return s.status.healthy.Load()
}

func (s *Service) performHealthCheck() {
        s.status.lastCheck = time.Now()
        healthy := s.status.healthy.Load()
        log.Printf("Health Check - Status: %v, Last Check: %s",
          healthy,
          s.status.lastCheck.Format(time.RFC3339))
}

func (s *Service) run() {
        for {
                select {
                case <-s.ctx.Done():
                        log.Println("Router service shutting down...")
                        return
                case <-s.ticker.C:
                        s.performHealthCheck()
                }
        }
}
