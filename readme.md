# krouter

A lightweight router implementation using CoreDNS and CoreDHCP to provide DNS and DHCP services while handling NAT and network interface configuration.

## Features

- IPv4/IPv6 Interface Configuration 
- DHCP Server (v4/v6) - for hijacking hosts networking
- DNS Server - for hijacking dns 
- NAT/Forwarding - for LAN to WAN 
- MiTM Cert Deployment - for hijacking ssl/tls
- Logging - for multiple output formats
- Pcap - for post-mortem analysis

## Project Structure

```
krouter/
├── main.go
├── pkg/
│   ├── cmd/
│   │   └── cmd.go          # CLI implementation using cobra
│   ├── config/
│   │   └── config.go       # YAML configuration handling
│   ├── dhcp/
│   │   └── server.go       # CoreDHCP server implementation
│   ├── dns/
│   │   └── server.go       # CoreDNS server implementation
│   └── network/
│       ├── network.go      # Handles network setup 
│       ├── interface.go    # Network interface configuration using netlink
│       └── forwarding.go   # NAT and forwarding rules using iptables
└── go.mod
```

## Configuration

Example `config.yaml`:
```yaml
interfaces:
  lan: "enx000ec677b5a6"
  wan: "wlp0s20f3"
configs:
  coredhcp: "./coredhcp.config"
  coredns: "./coredns.config"
```

## Dependencies

- github.com/spf13/cobra - CLI framework
- github.com/coredns/coredns - DNS server
- github.com/coredhcp/coredhcp - DHCP server
- github.com/vishvananda/netlink - Network interface configuration
- github.com/coreos/go-iptables - Firewall/NAT rules
- gopkg.in/yaml.v3 - Configuration parsing

## Usage

Basic usage:
```bash
# Start the router
./krouter

# Enable verbose logging
./krouter -v

# Use specific config file
./krouter --config custom-config.yaml
```

## Features Implemented

### Network Configuration
- Interface configuration (IPv4/IPv6)
- NAT/Forwarding setup
- IP forwarding enablement

### DHCP Server
- IPv4/IPv6 address assignment
- Network configuration distribution
- DNS server information distribution

### DNS Server
- Local zone hosting
- DNS forwarding
- Caching

## TODO

### Required
- [ ] Config fixes
    - [ ] Create config if !exist
- [ ] Add IPv6 NAT configuration
- [ ] Implement router advertisements
- [ ] Add interface validation
- [ ] Add firewall configuration

### Feature Request
- [ ] Create krouter config wrapper for coredhcp and coredns
- [ ] Pcap capture
- [ ] Better logging and output formats
- [ ] Add more configuration options
- [ ] Add metric collection
- [ ] diskless running? i.e. hold leases in mem?

## Building

```bash
go build -o krouter
```

## Requirements

- Linux (uses netlink and iptables)
- Root privileges (for network configuration)
- Go 1.22 or later

## License

