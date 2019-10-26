package dnstun

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

var (
	// DefaultTransport is a default configuration of the Transport.
	DefaultTransport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// DefaultClient is a default instance of the HTTP client.
	DefaultClient = &http.Client{
		Transport: DefaultTransport,
	}
)

const (
	// MappingForward means that first element in the prediction tuple
	// is a probability of associating DNS query to the "good" domain
	// names. The second element is a probability of "bad" domain.
	MappingForward = "forward"

	// MappingReverse is reversed representation of probabilities in
	// the prediction tuple returned by the model.
	MappingReverse = "reverse"
)

// mappings lists all available mapping types.
var mappings = map[string]struct{}{
	MappingForward: struct{}{},
	MappingReverse: struct{}{},
}

type Options struct {
	Mapping string
	Model   string
	Version string
	Runtime string
}

// Dnstun is a plugin to block DNS tunneling queries.
type Dnstun struct {
	opts      Options
	client    *http.Client
	tokenizer Tokenizer
}

// NewDnstun creates a new instance of the DNS tunneling detector plugin.
func NewDnstun(opts Options) *Dnstun {
	return &Dnstun{
		opts:      opts,
		client:    DefaultClient,
		tokenizer: NewTokenizer(enUS, 256),
	}
}

func (d *Dnstun) Name() string {
	return "dnstun"
}

func (d *Dnstun) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	var (
		state = request.Request{W: w, Req: r}
		resp  PredictResponse
	)

	req := PredictRequest{
		X: [][]int{d.tokenizer.TextToSeq(state.QName())},
	}

	p := path.Join("/models", d.opts.Model, d.opts.Version, "predict")

	u := url.URL{Scheme: "http", Host: d.opts.Runtime, Path: p}
	err := d.do(ctx, "POST", &u, req, &resp)
	if err != nil {
		return dns.RcodeServerFailure, plugin.Error(d.Name(), err)
	}

	if len(resp.Y) != 1 || len(resp.Y[0]) == 0 {
		err = errors.Errorf("invalid predict response: %#v", resp)
		return dns.RcodeServerFailure, plugin.Error(d.Name(), err)
	}

	// Select max argument position from the response vector.
	var (
		yPos int     = 0
		yMax float64 = resp.Y[0][yPos]
	)
	for i := yPos + 1; i < len(resp.Y[0]); i++ {
		if resp.Y[0][i] > yMax {
			yPos = i
			yMax = resp.Y[0][i]
		}
	}

	// The first position of the prediction vector corresponds to the DNS
	// tunneling class, therefore such requests should be rejected.
	if (d.opts.Mapping == MappingForward && yPos == 1) ||
		(d.opts.Mapping == MappingReverse && yPos == 0) {

		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeRefused)
		w.WriteMsg(m)
		return dns.RcodeRefused, nil
	}

	// Pass control to the next plugin.
	return dns.RcodeSuccess, nil
}

// PredictRequest is a request to get predictions for the given attribute vectors.
type PredictRequest struct {
	X [][]int `json:"x"`
}

// PredictResponse lists probabilities for each attribute vector.
type PredictResponse struct {
	Y [][]float64 `json:"y"`
}

func (d *Dnstun) do(ctx context.Context, method string, u *url.URL, in, out interface{}) error {
	var (
		b   []byte
		err error
	)

	if in != nil {
		b, err = json.Marshal(in)
		if err != nil {
			return errors.Wrapf(err, "failed to encode request")
		}
	}
	req, err := http.NewRequest(method, u.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp, err := d.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	// Decode the list of nodes from the body of the response.
	defer resp.Body.Close()

	// If server returned non-zero status, the response body is treated
	// as a error message, which will be returned to the user.
	if resp.StatusCode != http.StatusOK {
		// Server could return a response error within a header.
		errorCode := resp.Header.Get(http.CanonicalHeaderKey("Error-Code"))
		if errorCode != "" {
			return errors.New(errorCode)
		}
		return errors.Errorf("unexpected response from server: %d", resp.StatusCode)
	}

	if out == nil {
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	return errors.Wrapf(decoder.Decode(out), "failed to decode response")
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
