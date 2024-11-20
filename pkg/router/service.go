package router

import (
        "context"
        "log"
        "sync"
        "sync/atomic"
        "time"

        "github.com/ryanvillarreal/krouter/pkg/config"
        "github.com/ryanvillarreal/krouter/pkg/dhcp"
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
        dhcp     *dhcp.Service
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
        }

        s.status.healthy.Store(true)
        return s, nil
}

func (s *Service) IsHealthy() bool {
        return s.status.healthy.Load()
}

func (s *Service) performHealthCheck() {
        // For now, just update lastCheck time
        s.status.lastCheck = time.Now()
        healthy := s.status.healthy.Load()
        log.Printf("Health Check - Status: %v, Last Check: %s",
          healthy, 
          s.status.lastCheck.Format(time.RFC3339))
}

func (s *Service) Start() error {
        log.Printf("Starting router service on interfaces \nLAN: %s \nWAN: %s",
                s.cfg.Interfaces.LAN, s.cfg.Interfaces.WAN)

        s.wg.Add(1)
        go func() {
                defer s.wg.Done()
                s.run()
        }()

        return nil
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

func (s *Service) Stop() {
        s.ticker.Stop()
        s.cancel()
        s.wg.Wait()
        log.Println("Router service stopped")
}
