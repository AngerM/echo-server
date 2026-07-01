package main

import (
	"crypto/tls"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestServeHTTP_Paths(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		target         string
		requestBody    string
		isTLS          bool
		expectedPath   string
		expectedQuery  url.Values
		expectedMethod string
		expectedScheme string
	}{
		{
			name:           "Root path",
			method:         "GET",
			target:         "/",
			expectedPath:   "/",
			expectedQuery:  url.Values{},
			expectedMethod: "GET",
		},
		{
			name:           "Simple path",
			method:         "GET",
			target:         "/hello",
			expectedPath:   "/hello",
			expectedQuery:  url.Values{},
			expectedMethod: "GET",
		},
		{
			name:           "Deep path",
			method:         "GET",
			target:         "/some/deep/path/to/resource",
			expectedPath:   "/some/deep/path/to/resource",
			expectedQuery:  url.Values{},
			expectedMethod: "GET",
		},
		{
			name:           "Path with trailing slash",
			method:         "GET",
			target:         "/hello/",
			expectedPath:   "/hello/",
			expectedQuery:  url.Values{},
			expectedMethod: "GET",
		},
		{
			name:           "Path with query parameters",
			method:         "GET",
			target:         "/search?q=golang&limit=10",
			expectedPath:   "/search",
			expectedQuery:  url.Values{"q": []string{"golang"}, "limit": []string{"10"}},
			expectedMethod: "GET",
		},
		{
			name:           "Path with special characters",
			method:         "POST",
			target:         "/api/v1/users/info@example.com/profile",
			requestBody:    `{"status":"active"}`,
			expectedPath:   "/api/v1/users/info@example.com/profile",
			expectedQuery:  url.Values{},
			expectedMethod: "POST",
		},
		{
			name:           "Path with spaces (URL encoded)",
			method:         "GET",
			target:         "/files/my%20document.pdf",
			expectedPath:   "/files/my document.pdf",
			expectedQuery:  url.Values{},
			expectedMethod: "GET",
		},
		{
			name:           "Path with unicode characters",
			method:         "GET",
			target:         "/path/%F0%9F%8C%9F/star", // URL encoded 🌟
			expectedPath:   "/path/🌟/star",
			expectedQuery:  url.Values{},
			expectedMethod: "GET",
		},
		{
			name:           "Relative paths in target",
			method:         "GET",
			target:         "/a/b/../c",
			expectedPath:   "/a/b/../c",
			expectedQuery:  url.Values{},
			expectedMethod: "GET",
		},
		{
			name:           "Consecutive slashes",
			method:         "GET",
			target:         "//foo//bar",
			expectedPath:   "//foo//bar",
			expectedQuery:  url.Values{},
			expectedMethod: "GET",
		},
		{
			name:           "TLS Request / HTTPS scheme",
			method:         "GET",
			target:         "/secure-path",
			isTLS:          true,
			expectedPath:   "/secure-path",
			expectedQuery:  url.Values{},
			expectedMethod: "GET",
			expectedScheme: "https",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.requestBody != "" {
				bodyReader = strings.NewReader(tc.requestBody)
			}
			req := httptest.NewRequest(tc.method, tc.target, bodyReader)
			if tc.isTLS {
				req.TLS = &tls.ConnectionState{}
			}
			rec := httptest.NewRecorder()

			ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status OK, got %d", rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}

			cacheControl := rec.Header().Get("Cache-Control")
			if cacheControl != "no-cache" {
				t.Errorf("expected Cache-Control no-cache, got %s", cacheControl)
			}

			var resp reqStruct
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if resp.Method != tc.expectedMethod {
				t.Errorf("expected method %q, got %q", tc.expectedMethod, resp.Method)
			}

			if resp.URL.Path != tc.expectedPath {
				t.Errorf("expected path %q, got %q (URL: %s)", tc.expectedPath, resp.URL.Path, resp.URL.String())
			}

			expectedScheme := tc.expectedScheme
			if expectedScheme == "" {
				expectedScheme = "http"
			}
			if resp.URL.Scheme != expectedScheme {
				t.Errorf("expected scheme %q, got %q", expectedScheme, resp.URL.Scheme)
			}

			// Compare queries
			for k, expectedVals := range tc.expectedQuery {
				actualVals, ok := resp.Query[k]
				if !ok {
					t.Errorf("expected query parameter %q not found", k)
					continue
				}
				if len(expectedVals) != len(actualVals) {
					t.Errorf("expected parameter %q to have %d values, got %d", k, len(expectedVals), len(actualVals))
					continue
				}
				for i, v := range expectedVals {
					if actualVals[i] != v {
						t.Errorf("expected parameter %q index %d to be %q, got %q", k, i, v, actualVals[i])
					}
				}
			}
		})
	}
}

// Make sure we have a type matching the response structure.
// Note: We need to adapt to jsonv2's unmarshaling of url.URL.
type reqStruct struct {
	Method     string
	Headers    http.Header
	Body       string
	ParsedBody map[string]any
	URL        *url.URL
	Query      url.Values
}
