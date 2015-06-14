package rirstat

import (
	"strings"
	"testing"

	"github.com/kr/pretty"
)

func TestParseHeader(t *testing.T) {
	input := `
2|afrinic|20150612|7236|00000000|20150612|00000
afrinic|*|asn|*|2302|summary
afrinic|*|ipv4|*|2935|summary
afrinic|*|ipv6|*|1999|summary
afrinic|ZA|asn|1228|1|19910301|allocated|F36B9F4B
afrinic|ZA|asn|1229|1|19910301|allocated|F36B9F4B
afrinic|ZA|asn|1230|1|19910301|allocated|F36B9F4B
afrinic|ZA|asn|1231|1|19910301|allocated|F36B9F4B
`
	hdr, recs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	pretty.Print(hdr)
	pretty.Print(recs)
}
