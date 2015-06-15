package benchmark

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	client    *http.Client
	server    string
	transport *http.Transport
}

func NewClient(server string, transport *http.Transport) *Client {
	c := &Client{
		server:    server,
		transport: transport,
	}
	c.ResetClient()
	return c
}

func (c *Client) ResetClient() {
	cookieJar, _ := cookiejar.New(nil)
	c.client = &http.Client{
		Jar:       cookieJar,
		Transport: c.transport,
	}
}

func (c *Client) Execute(call *Call, keepAlive bool) *CallStats {
	if call.Wait > 0 {
		time.Sleep(time.Millisecond * time.Duration(call.Wait))
	}
	s := &CallStats{
		Start: time.Now(),
	}
	//c.transport.DisableKeepAlives = true
	var uri string
	if strings.HasPrefix(call.URI, "/") {
		uri = c.server + call.URI
	} else {
		uri = call.URI
	}
	_, urlErr := url.Parse(uri)
	if urlErr != nil {
		panic(urlErr)
	}
	var resp *http.Response
	var err error
	if keepAlive == false {
		c.transport.DisableKeepAlives = true
	}
	switch call.Method {
	case "GET":
		resp, err = c.client.Get(uri)
	case "POST":
		resp, err = c.client.Post(uri, call.Mimetype, strings.NewReader(call.Data))
	default:
		panic(errors.New("unsupported call method:" + call.Method))
	}
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		panic(errors.New("that was a 500"))
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if call.Debug {
		log.Println(string(body))
	}
	resp.Body.Close()
	s.Duration = time.Now().Sub(s.Start)
	return s
}
