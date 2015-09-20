// Package ipcc provides IP to country code mapping
package ipcc

import (
	"encoding/binary"
	"net"
	"sort"
)

//go:generate go run gen.go

func lookup4(ip4 net.IP) string {
	ip := binary.BigEndian.Uint32(ip4)
	i := sort.Search(len(ipv4blocks), func(n int) bool {
		return ipv4blocks[n].e >= ip
	})
	if i < len(ipv4blocks) && ipv4blocks[i].s <= ip && ip <= ipv4blocks[i].e {
		return ccs[ipv4blocks[i].cc&0xff]
	}
	return ""
}

func lookup6(ip net.IP) string {
	prefix := binary.BigEndian.Uint64(ip)
	i := sort.Search(len(ipv6blocks), func(n int) bool {
		v := ipv6blocks[n]
		e := v.p + ^uint64(0)>>(v.l)
		return e >= prefix
	})
	if i < len(ipv6blocks) {
		v := ipv6blocks[i]
		e := v.p + ^uint64(0)>>(v.l)
		if v.p <= prefix && prefix <= e {
			return ccs[v.cc&0xff]
		}
	}
	return ""
}

// Lookup returns a 2 letter country code for the given IP or "" if the IP
// isn't found in the database.
func Lookup(ip net.IP) string {
	if ip4 := ip.To4(); ip4 != nil {
		return lookup4(ip4)
	}
	return lookup6(ip)
}
