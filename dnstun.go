package dnstun

//go:generate protoc --proto_path=. --go_out=plugins=grpc:. resolver.proto

import (
	"context"
	"io"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Options struct {
	Host string
}

// Dnstun is a plugin to block DNS tunneling queries.
type Dnstun struct {
	opts      Options
	conn      io.Closer
	resolver  ResolverClient
	tokenizer Tokenizer
}

func NewDnstun(opts Options) *Dnstun {
	return &Dnstun{opts: opts, tokenizer: NewTokenizer(enUS, 256)}
}

func (dt *Dnstun) Dial() error {
	conn, err := grpc.Dial(dt.opts.Host)
	if err != nil {
		return plugin.Error(dt.Name(), errors.Wrapf(err, "failed to dial %q", dt.opts.Host))
	}
	dt.conn = conn
	dt.resolver = NewResolverClient(conn)
	return nil
}

// Close closes connection to the DNS tunneling server.
func (dt *Dnstun) Close() error {
	err := dt.conn.Close()
	if err != nil {
		switch errors.Cause(err) {
		case grpc.ErrClientConnClosing:
			return nil
		default:
			return plugin.Error(dt.Name(), err)
		}
	}
	return nil
}

func (dt *Dnstun) Name() string {
	return "dnstun"
}

func (dt *Dnstun) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	//seq := dt.tokenizer.TextToSeq(state.QName())
	req := ResolveRequest{Name: state.QName()}

	// When the remote resolver is not available, simply forward request
	// processing to the next handler as there is no error.
	resp, err := dt.resolver.Resolve(ctx, &req)
	if err != nil {
		return dns.RcodeSuccess, nil
	}

	switch resp.Action {
	case Action_REJECT:
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeRefused)
		w.WriteMsg(m)

		return dns.RcodeRefused, nil
	}

	// Pass the control to the next plugin.
	return dns.RcodeSuccess, nil
}

type chainHandler struct {
	plugin.Handler
	next plugin.Handler
}

func newChainHandler(h plugin.Handler) plugin.Plugin {
	return func(next plugin.Handler) plugin.Handler {
		return chainHandler{h, next}
	}
}

func (p chainHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	rcode, err := p.Handler.ServeDNS(ctx, w, r)
	if rcode != dns.RcodeSuccess {
		return rcode, err
	}

	state := request.Request{W: w, Req: r}
	return plugin.NextOrFailure(state.Name(), p.next, ctx, w, r)
}
