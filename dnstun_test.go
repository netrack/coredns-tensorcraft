package dnstun

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	plugintest "github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

type TestPredictor struct {
	resp PredictResponse
	err  error
}

func (p *TestPredictor) Handle(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(p.resp)
	if err != nil || p.err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
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
		predictor TestPredictor
		rcode     int
		err       bool
	}{
		{TestPredictor{resp: PredictResponse{Y: [][]float64{{1.0, 0.2}}}}, dns.RcodeRefused, false},
		{TestPredictor{resp: PredictResponse{Y: [][]float64{{0.1, 0.7}}}}, dns.RcodeSuccess, false},
		{TestPredictor{err: errors.New("err")}, dns.RcodeServerFailure, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(tt.predictor.Handle))
			defer s.Close()
			defer s.CloseClientConnections()

			d := NewDnstun(Options{Host: strings.TrimLeft(s.URL, "http://")})
			req := plugintest.Case{Qname: "tunnel.example.org", Qtype: dns.TypeCNAME}

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
