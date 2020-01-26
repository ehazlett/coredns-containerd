package containerd

import (
	"context"
	"io"
	"os"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("containerd")

// Containerd is a containerd DNS resolver
type Containerd struct {
	SocketPath string
	Next       plugin.Handler
}

// ServeDNS implements the plugin.Handler interface
func (e Containerd) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	log.Debug("Received response")
	pw := NewResponsePrinter(w)
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	return plugin.NextOrFailure(e.Name(), e.Next, ctx, pw, r)
}

// Name returns the name of the plugin
func (e Containerd) Name() string { return "containerd" }

type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	// TODO: lookup label in containerd
	return r.ResponseWriter.WriteMsg(res)
}

var out io.Writer = os.Stdout
