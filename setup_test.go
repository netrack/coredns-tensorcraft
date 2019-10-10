package dnstun

import (
	"testing"

	"github.com/caddyserver/caddy"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		input string
		want  Options
	}{
		{`dnstun tcp://localhost:4545`, Options{"tcp://localhost:4545"}},
		{`dnstun unix:///var/run/dnstun.sock`, Options{"unix:///var/run/dnstun.sock"}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			c := caddy.NewTestController("dns", tt.input)
			opts, err := parseOptions(c)

			if err != nil {
				t.Fatalf("failed to parse options: %v", err)
			}
			if opts != tt.want {
				t.Errorf("wrong options %#v != %#v", opts, tt.want)
			}
		})
	}
}
