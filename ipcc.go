package ipcc

import (
	"encoding/binary"
	"net"
)

func lookup4(ip4 net.IP) string {
	ip := binary.BigEndian.Uint32(ip4)
	for _, v := range ipv4blocks {
		if v.s <= ip && ip <= v.e {
			return ccs[v.cc&0xff]
		}
	}
	return ""
}

func lookup6(ip net.IP) string {
	prefix := binary.BigEndian.Uint64(ip)
	for _, v := range ipv6blocks {
		e := v.p + ^uint64(0)>>(v.l)
		if v.p <= prefix && prefix <= e {
			return ccs[v.cc&0xff]
		}
	}
	return ""
}

func Lookup(ip net.IP) string {
	if ip4 := ip.To4(); ip4 != nil {
		return lookup4(ip4)
	}
	return lookup6(ip)
}
