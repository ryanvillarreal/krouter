package dns

import (
    "context"
    "fmt"
    "log"
    "net"
    "strings"
    "sync"
    "time"
    
    "github.com/miekg/dns"
    "github.com/ryanvillarreal/krouter/pkg/config"
)

type DNSProxy struct {
    cfg      *config.Config
    ctx      context.Context
    cancel   context.CancelFunc
    wg       sync.WaitGroup
    errChan  chan error
    domains  map[string][]net.IP // domain name -> IP addresses
}

func NewDNSProxy(cfg *config.Config) *DNSProxy {
    ctx, cancel := context.WithCancel(context.Background())
    proxy := &DNSProxy{
        cfg:     cfg,
        ctx:     ctx,
        cancel:  cancel,
        errChan: make(chan error, 1),
        domains: make(map[string][]net.IP),
    }
    
    // Initialize domain mappings
    proxy.initializeDomains()
    
    return proxy
}

func (p *DNSProxy) initializeDomains() {
    for _, domain := range p.cfg.DNS.LocalDomains {
        var ips []net.IP
        
        // Add IPv4 addresses
        for _, ipv4Str := range domain.IPv4 {
            if ip := net.ParseIP(ipv4Str); ip != nil {
                ips = append(ips, ip)
            } else {
                log.Printf("Warning: invalid IPv4 address for domain %s: %s", domain.Name, ipv4Str)
            }
        }
        
        // Add IPv6 addresses
        for _, ipv6Str := range domain.IPv6 {
            if ip := net.ParseIP(ipv6Str); ip != nil {
                ips = append(ips, ip)
            } else {
                log.Printf("Warning: invalid IPv6 address for domain %s: %s", domain.Name, ipv6Str)
            }
        }
        
        // Ensure domain name ends with a dot
        name := domain.Name
        if !strings.HasSuffix(name, ".") {
            name = name + "."
        }
        
        p.domains[name] = ips
    }
}

func (p *DNSProxy) Start() error {
    server := &dns.Server{
        Addr: ":53",
        Net:  "udp",
        Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
            p.handleDNSRequest(w, r)
        }),
    }

    p.wg.Add(1)
    go func() {
        defer p.wg.Done()
        if err := server.ListenAndServe(); err != nil {
            select {
            case p.errChan <- fmt.Errorf("DNS server error: %w", err):
            default:
            }
        }
    }()
    
    log.Println("DNS Proxy started on :53")
    return nil
}

func (p *DNSProxy) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
    if len(r.Question) == 0 {
        return
    }

    question := r.Question[0]
    qname := question.Name
    qtype := question.Qtype

    // Check if it's one of our local domains
    if ips, exists := p.domains[qname]; exists {
        m := new(dns.Msg)
        m.SetReply(r)
        m.Authoritative = true

        switch qtype {
        case dns.TypeA:
            // Add all IPv4 addresses
            for _, ip := range ips {
                if ipv4 := ip.To4(); ipv4 != nil {
                    rr := &dns.A{
                        Hdr: dns.RR_Header{
                            Name:   qname,
                            Rrtype: dns.TypeA,
                            Class:  dns.ClassINET,
                            Ttl:    300,
                        },
                        A: ipv4,
                    }
                    m.Answer = append(m.Answer, rr)
                }
            }
        case dns.TypeAAAA:
            // Add all IPv6 addresses
            for _, ip := range ips {
                if ip.To4() == nil { // Is IPv6
                    rr := &dns.AAAA{
                        Hdr: dns.RR_Header{
                            Name:   qname,
                            Rrtype: dns.TypeAAAA,
                            Class:  dns.ClassINET,
                            Ttl:    300,
                        },
                        AAAA: ip,
                    }
                    m.Answer = append(m.Answer, rr)
                }
            }
        }

        if len(m.Answer) > 0 {
            w.WriteMsg(m)
            return
        }
    }

    // Forward request using WAN interface
    c := &dns.Client{
        Timeout: 5 * time.Second,
        Dialer: &net.Dialer{
            LocalAddr: &net.UDPAddr{
                IP:   net.ParseIP("0.0.0.0"),
                Port: 0,
            },
        },
    }

    // Try each upstream DNS server until one responds
    var lastErr error
    for _, upstream := range p.cfg.DNS.Upstream.IPv4 {
        m, _, err := c.Exchange(r, upstream+":53")
        if err == nil && m != nil {
            w.WriteMsg(m)
            return
        }
        lastErr = err
    }

    log.Printf("Failed to forward DNS request: %v", lastErr)
    
    // Return SERVFAIL if all upstream servers fail
    m := new(dns.Msg)
    m.SetRcode(r, dns.RcodeServerFailure)
    w.WriteMsg(m)
}

func (p *DNSProxy) Stop() {
    p.cancel()
    p.wg.Wait()
    log.Println("DNS Proxy stopped")
}

func (p *DNSProxy) Errors() <-chan error {
    return p.errChan
}
