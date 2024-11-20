package dns

import (
	"fmt"
	"log"

	"github.com/coredns/caddy"

	// Default CoreDNS plugins
	_ "github.com/coredns/coredns/plugin/cache"
	_ "github.com/coredns/coredns/plugin/forward"
	_ "github.com/coredns/coredns/plugin/log"
)

type Server struct {
	instance *caddy.Instance
}

func New(iface string) (*Server, error) {
	// Basic Corefile configuration
	conf := fmt.Sprintf(`.:53 {
		forward . 8.8.8.8 8.8.4.4
		cache
		log
	}`)

	// Load configuration
	instance, err := caddy.Start(caddy.CaddyfileInput{
		Contents:       []byte(conf),
		ServerTypeName: "dns",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start DNS server: %w", err)
	}

	return &Server{
		instance: instance,
	}, nil
}

func (s *Server) Start() error {
	log.Printf("DNS server running")
	return nil
}

func (s *Server) Stop() error {
	if s.instance != nil {
		log.Printf("Stopping DNS server")
		s.instance.Stop()
	}
	return nil
}
