package dnstun

import (
	"strings"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("dnstun", setup) }

func setup(c *caddy.Controller) error {
	opts, err := parseOptions(c)
	if err != nil {
		return plugin.Error("dnstun", err)
	}

	p := NewDnstun(opts)
	dnsserver.GetConfig(c).AddPlugin(newChainHandler(p))
	return nil
}

func parseOptions(c *caddy.Controller) (opts Options, err error) {
	c.Next() // directive name

	for c.NextBlock() {
		switch c.Val() {
		case "runtime":
			if !c.Args(&opts.Runtime) {
				return opts, c.ArgErr()
			}
		case "detector":
			var mapping, detector string
			if !c.Args(&mapping, &detector) {
				return opts, c.ArgErr()
			}

			if _, ok := mappings[mapping]; !ok {
				return opts, c.Errf("unknown mapping %q", mapping)
			}

			tuple := strings.SplitN(detector, ":", 2)
			if len(tuple) != 2 {
				return opts, c.Errf("unknown detector name %q", detector)
			}

			opts.Mapping = mapping
			opts.Model, opts.Version = tuple[0], tuple[1]
		default:
			return opts, c.Errf("unknown property %q", c.Val())
		}
	}
	return
}
