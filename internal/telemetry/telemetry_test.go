package telemetry

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPClient_setUAHeader(t *testing.T) {
	tests := []struct {
		name   string
		c      *HTTPClient
		wantUA string
	}{
		{
			name:   "provider Foo@1.0.0, tf@1.5.7",
			c:      NewHTTPClient("Foo", "1.0.0", "1.5.7"),
			wantUA: "tfProvider-Foo@1.0.0 (terraform@1.5.7)",
		},
		{
			name:   "provider Bar@0.0.1, tf@1.6.2",
			c:      NewHTTPClient("Bar", "0.0.1", "1.6.2"),
			wantUA: "tfProvider-Bar@0.0.1 (terraform@1.6.2)",
		},
		{
			name:   "tf version is unknown",
			c:      NewHTTPClient("Bar", "0.0.1", ""),
			wantUA: "tfProvider-Bar@0.0.1 (terraform@)",
		},
		{
			name:   "provider version is unknown",
			c:      NewHTTPClient("Bar", "", "1.5.7"),
			wantUA: "",
		},
		{
			name:   "provider name is unknown",
			c:      NewHTTPClient("Bar", "", "1.5.7"),
			wantUA: "",
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.c)
			var r = &http.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "foo.com",
				},
			}
			tt.c.setUAHeader(r)
			assert.Equal(t, tt.wantUA, r.Header.Get("User-Agent"))
		})
	}
}
