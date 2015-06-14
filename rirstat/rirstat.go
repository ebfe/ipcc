// Package rirstat provides a parser for the RIR statistic exchange format.
package rirstat

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

const timeFmt = "20060102"

type Header struct {
	Version   int
	Registry  string
	Serial    int
	Records   int
	StartDate time.Time
	EndDate   time.Time
	UTCOffset int
}

type Record struct {
	Registry   string
	CC         string
	Type       string
	Start      string
	Value      string
	Date       time.Time
	Status     string
	Extensions []string
}

func Parse(r io.Reader) (*Header, []Record, error) {
	var hdr *Header
	var records []Record

	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		cols := strings.Split(line, "|")
		if hdr == nil {
			h, err := parseHeader(cols)
			if err != nil {
				return nil, nil, err
			}
			hdr = h
			continue
		}

		if len(cols) < 6 {
			return nil, nil, errors.New("rirstat: format error")
		}

		// skip summary lines
		if cols[1] == "*" {
			continue
		}

		rec, err := parseRecord(cols)
		if err != nil {
			return nil, nil, err
		}
		records = append(records, *rec)
	}

	if err := s.Err(); err != nil {
		return nil, nil, err
	}

	return hdr, records, nil
}

func parseTime(s string) (time.Time, error) {
	if s == "00000000" {
		return time.Time{}, nil
	}
	return time.Parse(timeFmt, s)
}

func parseHeader(cols []string) (*Header, error) {
	var hdr Header
	var err error

	if len(cols) < 7 {
		return nil, errors.New("rirstat: header too short")
	}

	i, err := strconv.ParseInt(cols[0], 10, 16)
	if err != nil {
		return nil, err
	}
	hdr.Version = int(i)
	hdr.Registry = cols[1]
	i, err = strconv.ParseInt(cols[2], 10, 32)
	if err != nil {
		return nil, err
	}
	hdr.Serial = int(i)
	i, err = strconv.ParseInt(cols[3], 10, 32)
	if err != nil {
		return nil, err
	}
	hdr.Records = int(i)
	t, err := parseTime(cols[4])
	if err != nil {
		return nil, err
	}
	hdr.StartDate = t
	t, err = parseTime(cols[5])
	if err != nil {
		return nil, err
	}
	hdr.EndDate = t
	i, err = strconv.ParseInt(cols[6], 10, 32)
	if err != nil {
		return nil, err
	}
	hdr.UTCOffset = int(i)

	return &hdr, nil
}

func parseRecord(cols []string) (*Record, error) {
	var rec Record

	if len(cols) < 7 {
		return nil, errors.New("rirstat: record too short")
	}
	rec.Registry = cols[0]
	rec.CC = cols[1]
	rec.Type = cols[2]
	rec.Start = cols[3]
	rec.Value = cols[4]
	t, err := parseTime(cols[5])
	if err != nil {
		return nil, err
	}
	rec.Date = t
	rec.Status = cols[6]
	if len(cols) > 7 {
		rec.Extensions = cols[7:]
	}

	return &rec, nil

}
