package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"encoding/json/v2"
)

type req struct {
	Method     string
	Headers    http.Header
	Body       string
	ParsedBody map[string]any
	URL        *url.URL
	Query      url.Values
}

// ServeHTTP handles all incoming HTTP requests and echoes back information about the request.
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Reconstruct the full URL
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	u, _ := url.Parse(scheme + "://" + r.Host + r.URL.RequestURI())

	resp := req{
		Method:  r.Method,
		Headers: r.Header,
		Body:    string(body),
		URL:     u,
		Query:   u.Query(),
	}

	// Try to parse the request body as JSON
	_ = json.Unmarshal(body, &resp.ParsedBody)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	buf, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(buf)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      http.HandlerFunc(ServeHTTP),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
