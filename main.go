package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
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

func ServeHTTP(c *gin.Context) {
	defer c.Request.Body.Close()
	body, _ := ioutil.ReadAll(c.Request.Body)
	resp := req{
		Method:  c.Request.Method,
		Headers: c.Request.Header,
		Body:    string(body),
		URL:     c.Request.URL,
		Query:   c.Request.URL.Query(),
	}
	// Parse body if json
	json.Unmarshal(body, &resp.ParsedBody)
	c.Header("Content-Type", "application/json")
	c.Header("Cache-Control", "no-cache")
	buf, _ := json.Marshal(resp)
	c.Writer.Write(buf)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	s := gin.Default()
	s.Any("/*any", ServeHTTP)
	s.Run(":" + port)
}
