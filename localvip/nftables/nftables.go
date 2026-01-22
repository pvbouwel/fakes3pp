package nftables

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

func CheckNFTablesSupport(c *nftables.Conn) error {

	tables, err := c.ListTables()
	if err != nil {
		return fmt.Errorf("nftables netlink unavailable: %w", err)
	}

	for _, t := range tables {
		fmt.Println(t.Name)
		if t.Name == "nat" {
			return nil
		}
	}

	return fmt.Errorf("nat table not found")
}

func parseDecimalPort(port string) (uint16, error) {
	var base = 10
	var size = 16
	v, err := strconv.ParseUint(port, base, size)
	if err != nil {
		return 0, nil
	}
	return uint16(v), nil
}

func CreateVip(vipAddress, vipPort, targetAddress, targetPort string) error {
	c := &nftables.Conn{}

	err := CheckNFTablesSupport(c)
	if err != nil {
		return err
	}

	// NAT table
	table := &nftables.Table{
		Family: nftables.TableFamilyIPv4,
		Name:   "nat",
	}
	c.AddTable(table)

	// PREROUTING chain
	chain := &nftables.Chain{
		Name:     "prerouting",
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityNATDest,
	}
	c.AddChain(chain)

	destIP := net.ParseIP(vipAddress).To4()
	toIP := net.ParseIP(targetAddress).To4()

	vipPortBytes := make([]byte, 2)
	typedVipPort, err := parseDecimalPort(vipPort)
	if err != nil {
		return fmt.Errorf("encountered issue parsing vipPort: %w", err)
	}
	binary.BigEndian.PutUint16(vipPortBytes, typedVipPort)

	targetPortBytes := make([]byte, 2)
	typedTargetPort, err := parseDecimalPort(targetPort)
	if err != nil {
		return fmt.Errorf("encountered issue parsing targetPort: %w", err)
	}
	binary.BigEndian.PutUint16(targetPortBytes, typedTargetPort)

	rule := &nftables.Rule{
		Table: table,
		Chain: chain,
		Exprs: []expr.Any{
			// Load destination IP
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       16,
				Len:          4,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     destIP,
			},

			// Load TCP dport
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2, // TCP dest port
				Len:          2,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     vipPortBytes,
			},

			// DNAT
			&expr.Immediate{
				Register: 1,
				Data:     toIP,
			},
			&expr.Immediate{
				Register: 2,
				Data:     targetPortBytes,
			},
			&expr.NAT{
				Type:        expr.NATTypeDestNAT,
				Family:      unix.NFPROTO_IPV4,
				RegAddrMin:  1,
				RegProtoMin: 2,
			},
		},
	}

	c.AddRule(rule)

	if err := c.Flush(); err != nil {
		return err
	}
	return nil
}
