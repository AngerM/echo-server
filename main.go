package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	json "github.com/json-iterator/go"
)

type req struct {
	Method     string
	Headers    http.Header
	Body       string
	ParsedBody map[string]interface{}
	URL        *url.URL
	Query      url.Values
}

type echoServer struct{}

func (e echoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	resp := req{
		Method:  r.Method,
		Headers: r.Header,
		Body:    string(body),
		URL:     r.URL,
		Query:   r.URL.Query(),
	}
	json.Unmarshal(body, &resp.ParsedBody)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	buf, _ := json.Marshal(resp)
	w.Write(buf)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	myHandler := &echoServer{}
	s := http.Server{
		Addr:    ":" + port,
		Handler: myHandler,
	}
	s.ListenAndServe()
}
