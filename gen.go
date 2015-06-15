// +build ignore
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ebfe/ipcc/rirstat"
)

var registries = []string{"afrinic", "apnic", "arin", "iana", "lacnic", "ripe-ncc"}
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

func main() {
	if err := fetch(); err != nil {
		fmt.Fprintf(os.Stderr, "gen: fetch error: %s", err)
		os.Exit(1)
	}

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
	}
}
