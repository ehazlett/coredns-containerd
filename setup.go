package containerd

import (
	"context"
	"fmt"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/pkg/errors"
)

func init() { plugin.Register("containerd", setup) }

func setup(c *caddy.Controller) error {
	socketPath := ""
	namespace := ""
	label := ""

	for c.Next() {
		args := c.RemainingArgs()
		if len(args) != 3 {
			return fmt.Errorf("invalid config %q; expected <socket-path> <namespace>\n", args)
		}
		socketPath = args[0]
		namespace = args[1]
		label = args[2]
	}

	ctx := context.Background()

	ctrd, err := NewContainerd(ctx, socketPath, namespace, label)
	if err != nil {
		return errors.Wrap(err, "error connecting to containerd")
	}

	c.OnStartup(func() error {
		once.Do(func() { metrics.MustRegister(c, requestCount) })
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		ctrd.Next = next
		return ctrd
	})

	return nil
}
