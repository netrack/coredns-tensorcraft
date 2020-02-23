package dnstun

import (
	"context"
	"io/ioutil"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	tfop "github.com/tensorflow/tensorflow/tensorflow/go/op"
)

type Options struct {
	Input  string
	Output string
	Graph  string
}

// Dnstun is a plugin to block DNS tunneling queries.
type Dnstun struct {
	inputOp   string
	outputOp  string
	graph     *tf.Graph
	tokenizer Tokenizer
}

// NewDnstun creates a new instance of the DNS tunneling detector plugin.
func NewDnstun(opts Options) (*Dnstun, error) {
	b, err := ioutil.ReadFile(opts.Graph)
	if err != nil {
		return nil, err
	}

	graph := tf.NewGraph()
	if err := graph.Import(b, ""); err != nil {
		return nil, err
	}

	return &Dnstun{
		inputOp:   opts.Input,
		outputOp:  opts.Output,
		graph:     graph,
		tokenizer: NewTokenizer(enUS, 256),
	}, nil
}

func (d *Dnstun) Name() string {
	return "dnstun"
}

func (d *Dnstun) argmax(in *tf.Tensor, dim int64) (int64, error) {
	inShape := tf.MakeShape(in.Shape()...)
	root := tfop.NewScope()

	input := tfop.Placeholder(root, tf.Float, tfop.PlaceholderShape(inShape))
	argmax := tfop.ArgMax(root, input, tfop.Const(root, dim))

	graph, err := root.Finalize()
	if err != nil {
		return -1, err
	}

	sess, err := tf.NewSession(graph, nil)
	if err != nil {
		return -1, err
	}

	output, err := sess.Run(
		map[tf.Output]*tf.Tensor{input: in},
		[]tf.Output{argmax},
		nil,
	)
	if err != nil {
		return -1, err
	}

	index, _ := output[0].Value().([]int64)
	return index[0], nil
}

func (d *Dnstun) predict(name string) (int64, error) {
	input, err := tf.NewTensor([][]int64{d.tokenizer.TextToSeq(name)})
	if err != nil {
		return -1, err
	}

	sess, err := tf.NewSession(d.graph, nil)
	if err != nil {
		return -1, err
	}

	defer sess.Close()

	output, err := sess.Run(
		map[tf.Output]*tf.Tensor{
			d.graph.Operation(d.inputOp).Output(0): input,
		},
		[]tf.Output{
			d.graph.Operation(d.outputOp).Output(0),
		},
		nil,
	)
	if err != nil {
		return -1, err
	}

	// Select max argument position from the response vector.
	return d.argmax(output[0], 1)
}

func (d *Dnstun) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	category, err := d.predict(state.QName())
	if err != nil {
		return dns.RcodeServerFailure, plugin.Error(d.Name(), err)
	}

	// The first position of the prediction vector corresponds to the DNS
	// tunneling class, therefore such requests should be rejected.
	if category == 0 {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeRefused)
		w.WriteMsg(m)
		return dns.RcodeRefused, nil
	}

	// Pass control to the next plugin.
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
