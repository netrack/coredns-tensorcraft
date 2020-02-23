package dnstun

import (
	"context"
	"testing"

	plugintest "github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

type TestResponseWriter struct {
	plugintest.ResponseWriter
	m *dns.Msg
}

func (rw *TestResponseWriter) WriteMsg(m *dns.Msg) error {
	rw.m = m
	return rw.ResponseWriter.WriteMsg(m)
}

func TestDnstunServeDNS(t *testing.T) {
	tests := []struct {
		qname string
		rcode int
		err   bool
	}{
		{"tunnel.example.org", dns.RcodeSuccess, false},
		{"r17788.tunnel.tuns.org", dns.RcodeRefused, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d, err := NewDnstun(Options{
				Input:  "sent_input_2",
				Output: "dense_18_2/Softmax",
				Graph:  "dnscnn.pb",
			})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

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
