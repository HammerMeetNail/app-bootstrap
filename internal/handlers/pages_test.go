package handlers

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPageHandler_Index(t *testing.T) {
	handler, err := NewPageHandler("../../web/templates")
	if err != nil {
		t.Fatalf("failed to create page handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.Index(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "__TEMPLATE_PROJECT_NAME__") {
		t.Fatalf("expected title to be present")
	}
	if !strings.Contains(body, "/static/css/") && !strings.Contains(body, "styles.css") {
		t.Fatalf("expected css link to be present")
	}
}

func TestPageHandler_NewPageHandler_InvalidDir(t *testing.T) {
	_, err := NewPageHandler(filepath.Join(os.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveBaseURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		TLS     bool
		headers map[string]string
		want    string
	}{
		{
			name: "direct request uses host + http",
			url:  "http://example.com/",
			want: "http://example.com",
		},
		{
			name: "tls sets https",
			url:  "http://example.com/",
			TLS:  true,
			want: "https://example.com",
		},
		{
			name: "x-forwarded-proto overrides scheme",
			url:  "http://example.com/",
			headers: map[string]string{
				"X-Forwarded-Proto": "https",
			},
			want: "https://example.com",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, c.url, nil)
			if c.TLS {
				req.TLS = &tls.ConnectionState{}
			}
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}

			got := resolveBaseURL(req)
			if got != c.want {
				t.Fatalf("expected %q, got %q", c.want, got)
			}
		})
	}
}
