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
			graph dns-cnn.onnx
		}`, Options{Graph: "dns-cnn.onnx"}},
		{`dnstun {
			graph sequential_1:0.0.0+build1
		}`, Options{Graph: "sequential_1:0.0.0+build1"}},
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
