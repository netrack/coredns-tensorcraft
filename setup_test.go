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
		{`dnstun {
			runtime  localhost:4545
			detector dns_cnn:latest
			mapping  reverse
		}`, Options{"reverse", "dns_cnn", "latest", "localhost:4545"}},
		{`dnstun {
			runtime  1.1.1.1:2345
			detector sequential_1:0.0.0+build1
			mapping  forward
		}`, Options{"forward", "sequential_1", "0.0.0+build1", "1.1.1.1:2345"}},
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
