package dnstun

import (
	"io/ioutil"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

func init() { plugin.Register("dnstun", setup) }

func setup(c *caddy.Controller) error {
	opts, err := parseOptions(c)
	if err != nil {
		return plugin.Error("dnstun", err)
	}

	b, err := ioutil.ReadFile(opts.Graph)
	if err != nil {
		return err
	}

	graph := tf.NewGraph()
	if err := graph.Import(b, ""); err != nil {
		return err
	}

	p := NewDnstun(graph)
	dnsserver.GetConfig(c).AddPlugin(newChainHandler(p))
	return nil
}

func parseOptions(c *caddy.Controller) (opts Options, err error) {
	c.Next() // directive name

	for c.NextBlock() {
		switch c.Val() {
		case "graph":
			if !c.Args(&opts.Graph) {
				return opts, c.ArgErr()
			}
		default:
			return opts, c.Errf("unknown property %q", c.Val())
		}
	}
	return
}
