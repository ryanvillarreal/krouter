package router

import (
    "fmt"
    "log"
    "net"

    "github.com/vishvananda/netlink"
    k_config "github.com/ryanvillarreal/krouter/pkg/config"
)

type InterfaceManager struct {
    lanIface string
    wanIface string
    cfg      *k_config.Config
}

func NewInterfaceManager(cfg *k_config.Config) *InterfaceManager {
    im := &InterfaceManager{
        lanIface: cfg.Interfaces.LAN.Iface,
        wanIface: cfg.Interfaces.WAN,
        cfg:      cfg,
    }
    if err := im.validateInterfaces(); err != nil {
        log.Printf("Interface validation failed: %v", err)
        return nil
    }

    if err := im.configureLAN(); err != nil {
        log.Printf("LAN configuration failed: %v", err)
        return nil
    }
    return im
}

func (im *InterfaceManager) configureLAN() error {
    lanLink, err := netlink.LinkByName(im.lanIface)

    if err != nil {
        return fmt.Errorf("failed to get LAN interface: %w", err)
    }

    if err := netlink.LinkSetDown(lanLink); err != nil {
        return fmt.Errorf("failed to set LAN interface down: %w", err)
    }
 
    // Add IPv4 address
    addr4, err := netlink.ParseAddr(im.cfg.Interfaces.LAN.IPv4)
    if err != nil {
        return fmt.Errorf("failed to parse IPv4 address: %w", err)
    }

    if addr4.String() != im.cfg.Interfaces.LAN.IPv4 {
      err := netlink.AddrAdd(lanLink, addr4)
      if err != nil {
        fmt.Println("PANIC, err := netlink.AddrAdd(lanLink, addr4)")
      }
    }

    if err := netlink.LinkSetUp(lanLink); err != nil {
        return fmt.Errorf("failed to set LAN interface up: %w", err)
    }
    return nil
}

func (im *InterfaceManager) validateInterfaces() error {
    interfaces, err := net.Interfaces()
    if err != nil {
        return fmt.Errorf("failed to get network interfaces: %w", err)
    }

    foundLAN := false
    foundWAN := false

    for _, iface := range interfaces {
        if iface.Name == im.lanIface {
            foundLAN = true
            fmt.Println("found lan iface")
        }
        if iface.Name == im.wanIface {
            foundWAN = true
            fmt.Println("found lan iface")
        }
    }

    if !foundLAN {
        return fmt.Errorf("LAN interface %s not found", im.lanIface)
    }
    if !foundWAN {
        return fmt.Errorf("WAN interface %s not found", im.wanIface)
    }

    return nil
}
