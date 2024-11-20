package router

import (
	"fmt"
	"github.com/coreos/go-iptables/iptables"
)

type ForwardingManager struct {
	lanIface string
	wanIface string
	ipt      *iptables.IPTables
}

func NewForwardingManager(lan, wan string) (*ForwardingManager, error) {
	ipt, err := iptables.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize iptables: %w", err)
	}

	return &ForwardingManager{
		lanIface: lan,
		wanIface: wan,
		ipt:      ipt,
	}, nil
}

func (fm *ForwardingManager) Configure() error {
	rules := []struct {
		table string
		chain string
		rule  []string
	}{
		// Allow forwarding
		{"filter", "FORWARD", []string{"-i", fm.lanIface, "-o", fm.wanIface, "-j", "ACCEPT"}},
		{"filter", "FORWARD", []string{"-i", fm.wanIface, "-o", fm.lanIface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"}},
		// Setup NAT
		{"nat", "POSTROUTING", []string{"-o", fm.wanIface, "-j", "MASQUERADE"}},
	}

	// Apply rules
	for _, r := range rules {
		exists, err := fm.ipt.Exists(r.table, r.chain, r.rule...)
		if err != nil {
			return fmt.Errorf("failed to check rule existence: %w", err)
		}

		if !exists {
			if err := fm.ipt.Append(r.table, r.chain, r.rule...); err != nil {
				return fmt.Errorf("failed to append rule: %w", err)
			}
		}
	}

	return nil
}

func (fm *ForwardingManager) Cleanup() error {
	rules := []struct {
		table string
		chain string
		rule  []string
	}{
		{"filter", "FORWARD", []string{"-i", fm.lanIface, "-o", fm.wanIface, "-j", "ACCEPT"}},
		{"filter", "FORWARD", []string{"-i", fm.wanIface, "-o", fm.lanIface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"}},
		{"nat", "POSTROUTING", []string{"-o", fm.wanIface, "-j", "MASQUERADE"}},
	}

	for _, r := range rules {
		exists, err := fm.ipt.Exists(r.table, r.chain, r.rule...)
		if err != nil {
			return fmt.Errorf("failed to check rule existence: %w", err)
		}

		if exists {
			if err := fm.ipt.Delete(r.table, r.chain, r.rule...); err != nil {
				return fmt.Errorf("failed to delete rule: %w", err)
			}
		}
	}

	return nil
}
