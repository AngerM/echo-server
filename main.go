package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/http2/config"
	"github.com/hertz-contrib/http2/factory"
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

func ServeHTTP(ctx context.Context, c *app.RequestContext) {
	body := c.Request.Body()
	headers := http.Header{}
	c.Request.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = []string{string(value)}
	})
	u,_ := url.Parse(string(c.Request.URI().FullURI()))
	resp := req{
		Method:  string(c.Request.Method()),
		Headers: headers,
		Body:    string(body),
		URL:     u,
		Query:   u.Query(),
	}
	// Parse body if json
	json.Unmarshal(body, &resp.ParsedBody)
	c.Header("Content-Type", "application/json")
	c.Header("Cache-Control", "no-cache")
	buf, _ := json.Marshal(resp)
	c.Status(200)
	c.Write(buf)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	h := server.New(server.WithHostPorts(":" + port))
	h.AddProtocol("h2", factory.NewServerFactory(
		config.WithReadTimeout(time.Minute),
		config.WithDisableKeepAlive(false)))
	h.NoRoute(ServeHTTP)
	h.Spin()
}
