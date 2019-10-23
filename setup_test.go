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
		{`dnstun http://localhost:4545`, Options{"http://localhost:4545"}},
		{`dnstun http://1.1.1.1:2345`, Options{"http://1.1.1.1:2345"}},
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
