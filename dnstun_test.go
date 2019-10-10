package dnstun

import (
	"context"
	"errors"
	"testing"

	plugintest "github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"google.golang.org/grpc"
)

type TestResolver struct {
	name string
	resp *ResolveResponse
	err  error
}

func (r TestResolver) Resolve(
	ctx context.Context, in *ResolveRequest, opts ...grpc.CallOption,
) (*ResolveResponse, error) {
	r.name = in.Name
	return r.resp, r.err
}

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
		resolver TestResolver
		rcode    int
	}{
		{TestResolver{resp: &ResolveResponse{Action: Action_REJECT}}, dns.RcodeRefused},
		{TestResolver{resp: &ResolveResponse{Action: Action_ACCEPT}}, dns.RcodeSuccess},
		{TestResolver{err: errors.New("err")}, dns.RcodeSuccess},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d := Dnstun{resolver: tt.resolver}
			req := plugintest.Case{Qname: "tunnel.example.org", Qtype: dns.TypeCNAME}

			rw := new(TestResponseWriter)
			rcode, err := d.ServeDNS(context.TODO(), rw, req.Msg())
			if rcode != tt.rcode {
				t.Errorf("rcode is wrong: %v != %v", rcode, tt.rcode)
			}
			if err != nil {
				t.Errorf("error returned: %v", err)
			}

			if tt.rcode != dns.RcodeSuccess && rw.m == nil {
				t.Fatalf("message is not written")
			}
			if tt.rcode != dns.RcodeSuccess && rw.m.Rcode != tt.rcode {
				t.Errorf("wrong rcode in response %v != %v", rw.m.Rcode, tt.rcode)
			}
		})
	}
}
