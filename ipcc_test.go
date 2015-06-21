package ipcc

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/abh/geoip"
)

var g4 *geoip.GeoIP
var g6 *geoip.GeoIP

func init() {
	var err error

	g4, err = geoip.Open("/usr/share/GeoIP/GeoIP.dat")
	if err != nil {
		fmt.Println("can't open GeoIP.dat")
	}

	g6, err = geoip.Open("/usr/share/GeoIP/GeoIPv6.dat")
	if err != nil {
		fmt.Println("can't open GeoIPv6.dat")
	}

	rand.Seed(time.Now().UnixNano() * int64(os.Getpid()))
}

const N = 10000

func TestMatchGeoIP(t *testing.T) {
	if g4 == nil {
		t.Skip("no geoip data")
	}
	matches := 0
	hits := 0
	for i := 0; i < N; i++ {
		ips := fmt.Sprintf("%d.%d.%d.%d", rand.Int31n(256), rand.Int31n(256), rand.Int31n(256), rand.Int31n(256))
		ip := net.ParseIP(ips)
		if ip == nil {
			panic("can't parse ip")
		}

		ccn := Lookup(ip)
		ccg, _ := g4.GetCountry(ips)
		ccg = strings.ToLower(ccg)

		if ccn == ccg {
			matches++
			if ccn != "" {
				hits++
			}
		}
	}
	r :=float32(matches)/N 
	t.Log(r, matches, hits)
	if  r < 0.9 {
		t.Error("< 90% matches with geoip db")
	}
}

func TestMatchGeoIPv6(t *testing.T) {
	if g6 == nil {
		t.Skip("no geoip data")
	}
	matches := 0
	hits := 0
	for i := 0; i < N; i++ {
		p := []int32{0x2001, 0x2607, 0x2610, 0x2620, 0x2800, 0x2801, 0x2804, 0x2806}
		ips := fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x", p[rand.Int31n(int32(len(p)))], rand.Int31n(0x10000), rand.Int31n(0x10000), rand.Int31n(0x10000), rand.Int31n(0x10000), rand.Int31n(0x10000), rand.Int31n(0x10000), rand.Int31n(0x10000))

		ip := net.ParseIP(ips)
		if ip == nil {
			panic("can't parse ip")
		}

		ccn := Lookup(ip)
		ccg, _ := g6.GetCountry_v6(ips)
		ccg = strings.ToLower(ccg)

		if ccn == ccg {
			matches++
			if ccn != "" {
				hits++
			}
		}
	}
	r :=float32(matches)/N 
	t.Log(r, matches, hits)
	if  r < 0.9 {
		t.Error("< 90% matches with geoip db")
	}
}
