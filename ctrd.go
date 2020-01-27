package containerd

import (
	"context"
	"io"
	"os"

	"github.com/containerd/containerd"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/plugin/pkg/fall"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

const (
	// A is a DNS A record
	A RecordType = "A"
	// AAAA is a DNS AAAA record
	AAAA RecordType = "AAAA"
	// TXT is a DNS TXT record
	TXT RecordType = "TXT"
	// DNSLabel is the container label used for record lookup
	DNSLabel = "io.containerd.ext.dns"
)

var log = clog.NewWithPlugin("containerd")

// Containerd is a containerd DNS resolver
type Containerd struct {
	client *containerd.Client
	zones  map[string][]Record

	Next plugin.Handler
	Fall fall.F
}

// RecordType is the type of DNS record
type RecordType string

// Record is a DNS record
type Record struct {
	Type  RecordType `json:"type,omitempty"`
	Name  string     `json:"name,omitempty"`
	Value string     `json:"value,omitempty"`
}

func NewContainerd(ctx context.Context, socketPath, namespace string) (*Containerd, error) {
	log.Debugf("connecting to containerd on %s", socketPath)
	client, err := getClient(socketPath, namespace)
	if err != nil {
		return nil, err
	}
	log.Debug("connected to containerd")
	ready, err := client.IsServing(ctx)
	if err != nil {
		return nil, err
	}
	log.Debugf("ready: %+v", ready)
	return &Containerd{
		client: client,
	}, nil
}

// ServeDNS implements the plugin.Handler interface
func (c *Containerd) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	query := state.Name()

	records, err := c.lookup(ctx, query, state.QType())
	if err != nil {
		// log the error or it will be swallowed by ServeDNS
		log.Error(err)
		return -1, err
	}
	// no records found; pass through
	if len(records) == 0 {
		return plugin.NextOrFailure(c.Name(), c.Next, ctx, w, r)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = true
	m.Answer = records

	log.Debugf("answering: query=%s", query)
	defer w.WriteMsg(m)

	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	return dns.RcodeSuccess, nil
}

// Name returns the name of the plugin
func (c Containerd) Name() string { return "containerd" }

func getClient(socketPath, namespace string) (*containerd.Client, error) {
	return containerd.New(socketPath, containerd.WithDefaultNamespace(namespace))
}

var out io.Writer = os.Stdout
