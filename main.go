package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type req struct {
	Headers http.Header
	Body    string
	URL     *url.URL
}

type echoServer struct{}

func (e echoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	resp := req{
		Headers: r.Header,
		Body:    string(body),
		URL:     r.URL,
	}
	w.WriteHeader(200)
	w.Header().Add("content-type", "application/json")
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
