// +build ignore

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"

	"github.com/ebfe/ipcc/rirstat"
)

var registries = []string{"afrinic", "apnic", "arin", "iana", "lacnic", "ripe-ncc"}
//var registries = []string{"afrinic"}
var mirror = "https://ftp.apnic.net/stats"

func filename(reg string) string {
	ffmt := "delegated-%s-extended-latest"
	if reg == "iana" {
		ffmt = "delegated-%s-latest"
	}
	return fmt.Sprintf(ffmt, strings.Replace(reg, "-", "", -1))
}

func fetch() error {
	for _, reg := range registries {
		fname := filename(reg)

		_, err := os.Stat(fname)
		if err == nil {
			fmt.Printf("%s: cached\n", reg)
			continue
		}
		if !os.IsNotExist(err) {
			return err
		}

		url := mirror + "/" + reg + "/" + fname
		fmt.Printf("%s: get %s\n", fname, url)
		rsp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer rsp.Body.Close()

		if rsp.StatusCode != http.StatusOK {
			return fmt.Errorf("gen: http error: %d %s", rsp.StatusCode, rsp.Status)
		}

		out, err := os.Create(fname)
		if err != nil {
			return err
		}
		defer out.Close()

		n, err := io.Copy(out, rsp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %d bytes\n", fname, n)
	}
	return nil
}

type ipv4block struct {
	First uint32
	Last  uint32
	Cc    [2]byte
}

type ipv6block struct {
	Prefix uint64
	Len    byte
	Cc     [2]byte
}

type v4ByAddr []ipv4block

func (ir v4ByAddr) Len() int           { return len(ir) }
func (ir v4ByAddr) Less(i, j int) bool { return ir[i].First < ir[j].First }
func (ir v4ByAddr) Swap(i, j int)      { ir[i], ir[j] = ir[j], ir[i] }

type v6ByAddr []ipv6block

func (ir v6ByAddr) Len() int           { return len(ir) }
func (ir v6ByAddr) Less(i, j int) bool { return ir[i].Prefix < ir[j].Prefix }
func (ir v6ByAddr) Swap(i, j int)      { ir[i], ir[j] = ir[j], ir[i] }

func merge4(ir []ipv4block) []ipv4block {
	sort.Sort(v4ByAddr(ir))

	w := 0
	for r := 0; r < len(ir); r++ {
		if ir[r].Cc != ir[w].Cc || ir[w].Last+1 <= ir[r].First {
			w++
			ir[w] = ir[r]
		} else {
			ir[w].Last = ir[r].Last
		}
	}
	return ir[:w+1]
}

func merge6(ir []ipv6block) []ipv6block {
	sort.Sort(v6ByAddr(ir))
	// TODO
	return ir
}

func main() {
	if err := fetch(); err != nil {
		fmt.Fprintf(os.Stderr, "gen: fetch error: %s", err)
		os.Exit(1)
	}

	ipv4 := []ipv4block{}
	ipv6 := []ipv6block{}

	for _, reg := range registries {
		fname := filename(reg)
		f, err := os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gen: open %s: %s\n", fname, err)
			os.Exit(1)
		}
		defer f.Close()

		hdr, recs, err := rirstat.Parse(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gen: parse %s: %s\n", fname, err)
			os.Exit(1)
		}
		fmt.Println(hdr.Registry, hdr.EndDate.Format("20060-01-02"), hdr.Records, len(recs), "records")

		for i := range recs {
			rec := &recs[i]
			if rec.CC == "ZZ" || rec.CC == "" {
				continue
			}
			switch rec.Type {
			case "ipv4":
				ip := net.ParseIP(rec.Start).To4()
				if ip == nil {
					fmt.Fprintf(os.Stderr, "gen: invalid ip %s: %s\n", rec.Start, err)
					continue
				}
				start := binary.BigEndian.Uint32([]byte(ip))
				end, err := strconv.ParseUint(rec.Value, 10, 32)
				if err != nil {
					fmt.Fprintf(os.Stderr, "gen: invalid length %s: %s\n", rec.Value, err)
					continue
				}
				ipv4 = append(ipv4, ipv4block{First: start, Last: start + uint32(end), Cc: [2]byte{rec.CC[0], rec.CC[1]}})
			case "ipv6":
				ip := net.ParseIP(rec.Start)
				if ip == nil {
					fmt.Fprintf(os.Stderr, "gen: invalid ip %s: %s\n", rec.Start, err)
					continue
				}
				plen, err := strconv.ParseUint(rec.Value, 10, 8)
				if err != nil {
					fmt.Fprintf(os.Stderr, "gen: invalid length %s: %s\n", rec.Value, err)
					continue
				}
				length := byte(plen)
				if plen > 64 {
					fmt.Fprintf(os.Stderr, "gen: prefix to long %d\n", plen)
					continue
				}
				prefix := binary.BigEndian.Uint64(ip)
				ipv6 = append(ipv6, ipv6block{Prefix: prefix, Len: length, Cc: [2]byte{rec.CC[0], rec.CC[1]}})
			}
		}
	}

	fmt.Println("ipv4:", len(ipv4))
	ipv4 = merge4(ipv4)
	fmt.Println(" -> :", len(ipv4))

	fmt.Println("ipv6:", len(ipv6))
	ipv6 = merge6(ipv6)
	fmt.Println(" -> :", len(ipv6))

	//pretty.Print(ipv4)
	//pretty.Print(ipv6)
	buf := bytes.NewBuffer(nil)
	buf.WriteString(
		`
	package main

	type ipv4block struct {
		First uint32
		Last  uint32
		Cc    [2]byte
	}

	type ipv6block struct {
		Prefix uint64
		Len    byte
		Cc     [2]byte
	}
	`)

	buf.WriteString("var ipv4blocks = [...]ipv4block{\n")
	for _, b := range ipv4 {
		fmt.Fprintf(buf, "%#v,\n", b)
	}
	buf.WriteString("}\n")
	buf.WriteString("var ipv6blocks = [...]ipv6block{\n")
	for _, b := range ipv6 {
		fmt.Fprintf(buf, "%#v,\n", b)
	}
	buf.WriteString("}\n")
	src := bytes.Replace(buf.Bytes(), []byte("main.ipv4block{"), []byte("{"), -1)
	src = bytes.Replace(src, []byte("main.ipv6block{"), []byte("{"), -1)
	src, err := imports.Process("db.go", src, nil)
	if err != nil {
		ioutil.WriteFile("db.go", buf.Bytes(), 0600)
		fmt.Fprintf(os.Stderr, "gen: imports: %s\n", err)
		os.Exit(1)
	}
	ioutil.WriteFile("db.go", src, 0600)
}
