package containerd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

func (c *Containerd) lookup(ctx context.Context, query string, qtype uint16) ([]dns.RR, error) {
	var (
		records []dns.RR
		err     error
	)

	host := strings.TrimSuffix(query, ".")
	containers, err := c.client.Containers(context.Background())
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		labels, err := container.Labels(ctx)
		if err != nil {
			return nil, err
		}

		var r []*Record
		for k, v := range labels {
			switch k {
			case c.label:
				if err := json.Unmarshal([]byte(v), &r); err != nil {
					return nil, err
				}
			}
		}

		for _, record := range r {
			rt, err := recordTypeToQtype(record)
			if err != nil {
				return nil, err
			}
			if record.Name != host || rt != qtype {
				continue
			}

			hdr := dns.RR_Header{
				Name:   query,
				Ttl:    0,
				Class:  dns.ClassINET,
				Rrtype: rt,
			}

			switch rt {
			case dns.TypeA:
				ip := net.ParseIP(record.Value)
				records = append(records, &dns.A{
					Hdr: hdr,
					A:   ip,
				})
			case dns.TypeAAAA:
				ip := net.ParseIP(record.Value)
				records = append(records, &dns.AAAA{
					Hdr:  hdr,
					AAAA: ip,
				})
			case dns.TypeTXT:
				records = append(records, &dns.TXT{
					Hdr: hdr,
					Txt: []string{record.Value},
				})
			}
		}
	}

	return records, nil
}

func recordTypeToQtype(r *Record) (uint16, error) {
	switch r.Type {
	case A:
		return dns.TypeA, nil
	case AAAA:
		return dns.TypeAAAA, nil
	case TXT:
		return dns.TypeTXT, nil
	}

	return 0, fmt.Errorf("unknown record type %s", r.Type)
}
