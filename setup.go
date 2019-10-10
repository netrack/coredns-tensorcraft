package dnstun

import (
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

	c.OnStartup(p.Dial)
	c.OnRestart(p.Close)
	c.OnFinalShutdown(p.Close)

	dnsserver.GetConfig(c).AddPlugin(newChainHandler(p))
	return nil
}

func parseOptions(c *caddy.Controller) (opts Options, err error) {
	c.Next() // directive name

	if !c.Args(&opts.Host) {
		return opts, c.ArgErr()
	}
	return
}
