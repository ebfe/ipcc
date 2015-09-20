package ipcc_test

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/abh/geoip"
	"github.com/ebfe/ipcc"
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

		ccn := ipcc.Lookup(ip)
		ccg, _ := g4.GetCountry(ips)
		ccg = strings.ToLower(ccg)

		if ccn == ccg {
			matches++
			if ccn != "" {
				hits++
			}
		}
	}
	r := float32(matches) / N
	t.Log(r, matches, hits)
	if r < 0.9 {
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

		ccn := ipcc.Lookup(ip)
		ccg, _ := g6.GetCountry_v6(ips)
		ccg = strings.ToLower(ccg)

		if ccn == ccg {
			matches++
			if ccn != "" {
				hits++
			}
		}
	}
	r := float32(matches) / N
	t.Log(r, matches, hits)
	if r < 0.8 {
		t.Error("< 90% matches with geoip db")
	}
}

func BenchmarkLookupIPv4(b *testing.B) {
	b.ReportAllocs()
	ip := []byte{0, 0, 0, 0}
	for i := 0; i < b.N; i++ {
		r := rand.Uint32()
		ip[0] = byte(r >> 24)
		ip[1] = byte(r >> 16)
		ip[2] = byte(r >> 8)
		ip[3] = byte(r)
		ipcc.Lookup(ip)
	}
}

func BenchmarkLookupIPv6(b *testing.B) {
	b.ReportAllocs()
	prefixes := []uint16{
		0x0000, 0x0100, 0x0200, 0x0400, 0x0800, 0x1000, 0x2000, 0x2001,
		0x2002, 0x2003, 0x2400, 0x2401, 0x2402, 0x2403, 0x2404, 0x2405,
		0x2406, 0x2407, 0x2408, 0x2409, 0x240a, 0x240b, 0x240c, 0x240d,
		0x240e, 0x240f, 0x2600, 0x2601, 0x2602, 0x2603, 0x2604, 0x2605,
		0x2606, 0x2607, 0x2608, 0x2609, 0x260a, 0x260c, 0x260d, 0x260e,
		0x260f, 0x2610, 0x2620, 0x2800, 0x2801, 0x2802, 0x2803, 0x2804,
		0x2805, 0x2806, 0x2807, 0x2808, 0x2a00, 0x2a01, 0x2a02, 0x2a03,
		0x2a04, 0x2a05, 0x2a06, 0x2a07, 0x2a08, 0x2c00, 0x2c08, 0x2c0c,
		0x2c0e, 0x2c0f, 0x4000, 0x6000, 0x8000, 0xa000, 0xc000, 0xe000,
		0xf000, 0xf800, 0xfc00, 0xfe00, 0xfe80, 0xfec0, 0xff00}
	ip := make([]byte, 16)
	for i := 0; i < b.N; i++ {
		pre := prefixes[i%len(prefixes)]
		binary.LittleEndian.PutUint16(ip[0:], pre)
		binary.LittleEndian.PutUint32(ip[2:], rand.Uint32())
		binary.LittleEndian.PutUint32(ip[6:], rand.Uint32())
		binary.LittleEndian.PutUint32(ip[10:], rand.Uint32())
		binary.LittleEndian.PutUint16(ip[14:], uint16(rand.Uint32()))
		ipcc.Lookup(ip)
	}
}

func ExampleLookup() {
	ip := net.ParseIP("78.42.0.1")
	cc := ipcc.Lookup(ip)
	fmt.Println(cc)
	// Output:
	// de
}
