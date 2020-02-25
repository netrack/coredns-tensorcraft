package dnstun

import (
	"context"
	"testing"

	plugintest "github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	tfop "github.com/tensorflow/tensorflow/tensorflow/go/op"
)

type TestResponseWriter struct {
	plugintest.ResponseWriter
	m *dns.Msg
}

func (rw *TestResponseWriter) WriteMsg(m *dns.Msg) error {
	rw.m = m
	return rw.ResponseWriter.WriteMsg(m)
}

func newTestGraph(inputLen, outputLen int64, index int) *tf.Graph {
	inShape := tf.MakeShape(1, inputLen)
	root := tfop.NewScope()

	matrix := make([][]float32, inputLen)
	for i := range matrix {
		matrix[i] = make([]float32, outputLen)
		matrix[i][index] = 1
	}

	layer1 := tfop.Placeholder(root, tf.Int64, tfop.PlaceholderShape(inShape))
	layer2 := tfop.Cast(root, layer1, tf.Float)
	tfop.MatMul(root, layer2, tfop.Const(root, matrix))

	graph, err := root.Finalize()
	if err != nil {
		panic(err)
	}
	return graph
}

func TestDnstunServeDNS(t *testing.T) {
	tests := []struct {
		index int
		qname string
		rcode int
		err   bool
	}{
		{1, "tunnel.example.org", dns.RcodeSuccess, false},
		{0, "r17788.tunnel.tuns.org", dns.RcodeRefused, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d := NewDnstun(newTestGraph(256, 2, tt.index))

			req := plugintest.Case{Qname: tt.qname, Qtype: dns.TypeCNAME}

			rw := new(TestResponseWriter)
			rcode, err := d.ServeDNS(context.TODO(), rw, req.Msg())
			if rcode != tt.rcode {
				t.Errorf("rcode is wrong: %v != %v", rcode, tt.rcode)
			}
			if err != nil && !tt.err {
				t.Errorf("error returned: %v", err)
			}

			if tt.rcode == dns.RcodeRefused && rw.m == nil {
				t.Fatalf("message is not written")
			}
			if tt.rcode == dns.RcodeRefused && rw.m.Rcode != tt.rcode {
				t.Errorf("wrong rcode in response %v != %v", rw.m.Rcode, tt.rcode)
			}
		})
	}
}
