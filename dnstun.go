package dnstun

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/pkg/errors"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	tfop "github.com/tensorflow/tensorflow/tensorflow/go/op"
)

type Options struct {
	Graph string
}

// Dnstun is a plugin to block DNS tunneling queries.
type Dnstun struct {
	predictGraph execGraph
	argmaxGraph  execGraph
	tokenizer    Tokenizer
}

// NewDnstun creates a new instance of the DNS tunneling detector plugin.
func NewDnstun(predictGraph *tf.Graph) *Dnstun {
	return &Dnstun{
		predictGraph: newExecGraph(predictGraph),
		argmaxGraph:  newExecGraph(newArgmax([]int64{1, 2}, tf.Float, 1)),
		tokenizer:    NewTokenizer(enUS, 256),
	}
}

func (d *Dnstun) Name() string {
	return "dnstun"
}

func (d *Dnstun) predict(name string) (int64, error) {
	input, err := tf.NewTensor([][]int64{d.tokenizer.TextToSeq(name)})
	if err != nil {
		return -1, err
	}

	output, err := d.predictGraph.Exec(input)
	if err != nil {
		return -1, err
	}
	if len(output) == 0 {
		return -1, errors.New("prediction graph returned empty tensor")
	}

	// Select max argument position from the response vector.
	output, err = d.argmaxGraph.Exec(output[0])
	if err != nil {
		return -1, err
	}
	if len(output) == 0 {
		return -1, errors.New("argmax returned empty tensor")
	}
	index, ok := output[0].Value().([]int64)
	if !ok {
		return -1, errors.Errorf("unexpected output type %T", output[0].Value())
	}
	if len(index) == 0 {
		return -1, errors.New("argmax return empty result")
	}
	return index[0], nil
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

func newArgmax(shape []int64, dtype tf.DataType, dim int64) (graph *tf.Graph) {
	inShape := tf.MakeShape(shape...)
	root := tfop.NewScope()

	input := tfop.Placeholder(root, dtype, tfop.PlaceholderShape(inShape))
	tfop.ArgMax(root, input, tfop.Const(root, dim))

	graph, err := root.Finalize()
	if err != nil {
		panic(err)
	}
	return graph
}

type execGraph struct {
	graphInput  tf.Output
	graphOutput tf.Output
	graph       *tf.Graph
}

func newExecGraph(graph *tf.Graph) execGraph {
	var (
		ops   = graph.Operations()
		input tf.Output
	)

	for _, o := range ops {
		if o.Type() == "Placeholder" {
			input = o.Output(0)
			break
		}
	}

	if input == (tf.Output{}) {
		panic("graph without input")
	}
	return execGraph{
		graphInput:  input,
		graphOutput: ops[len(ops)-1].Output(0),
		graph:       graph,
	}
}

func (e execGraph) Exec(in *tf.Tensor) (output []*tf.Tensor, err error) {
	sess, err := tf.NewSession(e.graph, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		e := sess.Close()
		if e != nil {
			if err != nil {
				err = errors.WithMessage(err, e.Error())
			} else {
				err = e
			}
		}
	}()

	return sess.Run(
		map[tf.Output]*tf.Tensor{e.graphInput: in},
		[]tf.Output{e.graphOutput},
		nil,
	)
}
