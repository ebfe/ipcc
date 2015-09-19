package main

import (
	"fmt"
	"github.com/ebfe/ipcc"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <ip>\n", os.Args[0])
		os.Exit(1)
	}
	ip := net.ParseIP(os.Args[1])
	if ip == nil {
		fmt.Fprintf(os.Stderr, "ipcc: can't parse ip %q\n", os.Args[1])
		os.Exit(1)
	}
	if cc := ipcc.Lookup(ip); cc != "" {
		fmt.Println(cc)
		os.Exit(0)
	}
	os.Exit(1)
}
