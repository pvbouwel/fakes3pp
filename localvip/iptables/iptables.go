package iptables

import (
	"errors"

	ipt "github.com/coreos/go-iptables/iptables"
)

var ipTables *ipt.IPTables

func GetIpTables() (*ipt.IPTables, error) {
	if ipTables == nil {
		ipt, err := ipt.New()
		if err != nil {
			return nil, err
		}
		ipTables = ipt
	}
	return ipTables, nil
}

func CreateVip(vipAddress, vipPort, targetAddress, targetPort string) error {
	if vipAddress == "" {
		return errors.New("vipAddress cannot be empty when creating VIP")
	}

	if vipPort == "" {
		return errors.New("vipPort cannot be empty when creating VIP")
	}

	if targetAddress == "" {
		return errors.New("targetAddress cannot be empty when creating VIP")
	}

	if targetPort == "" {
		return errors.New("targetPort cannot be empty when creating VIP")
	}

	ipTables, err := GetIpTables()
	if err != nil {
		return err
	}

	return ipTables.InsertUnique(
		"nat", "PREROUTING", 1, "-p", "tcp", "-d", vipAddress, "--dport", vipPort,
		"-j", "DNAT", "--to-destination", targetAddress+":"+targetPort,
	)
}
