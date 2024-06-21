package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/marmotedu/errors"
)

type Request struct {
	host    string
	path    string
	body    []byte
	method  string
	headers map[string]string
}

type HttpClient struct {
	Request
	client *http.Client
}

var (
	client *HttpClient
	once   sync.Once
)

func NewHttpClient() *HttpClient {
	once.Do(func() {
		client = &HttpClient{
			client: &http.Client{},
			Request: Request{
				headers: make(map[string]string),
			},
		}
	})
	return client
}

type Option interface {
	apply(*HttpClient)
}

type optionFunc func(*HttpClient)

func (f optionFunc) apply(c *HttpClient) {
	f(c)
}

func WithHost(host string) Option {
	return optionFunc(func(c *HttpClient) {
		c.host = host
	})
}

func WithToHermes() Option {
	return optionFunc(func(c *HttpClient) {
		c.host = GetEnvDefault("HERMESURL", "http://121.41.31.123:31447")
	})
}

func WithPath(path string) Option {
	return optionFunc(func(c *HttpClient) {
		c.path = path
	})
}

func WithBody(body []byte) Option {
	return optionFunc(func(c *HttpClient) {
		c.body = body
	})
}

func WithMethod(method string) Option {
	return optionFunc(func(c *HttpClient) {
		c.method = method
	})
}

func WithUseGet() Option {
	return optionFunc(func(c *HttpClient) {
		c.method = http.MethodGet
	})
}

func WithUsePost() Option {
	return optionFunc(func(c *HttpClient) {
		c.method = http.MethodPost
	})
}

func WithUsePut() Option {
	return optionFunc(func(c *HttpClient) {
		c.method = http.MethodPut
	})
}

func WithUseDelete() Option {
	return optionFunc(func(c *HttpClient) {
		c.method = http.MethodDelete
	})
}

func WithUsePatch() Option {
	return optionFunc(func(c *HttpClient) {
		c.method = http.MethodPatch
	})
}

func WithHeader(key, value string) Option {
	return optionFunc(func(c *HttpClient) {
		c.headers[key] = value
	})
}

func (c *HttpClient) Call(opts ...Option) (*http.Response, error) {
	for _, o := range opts {
		o.apply(c)
	}

	url := fmt.Sprintf("%s%s", c.host, c.path)
	req, err := http.NewRequest(c.method, url, bytes.NewBuffer(c.body))
	if err != nil {
		return nil, errors.WrapC(err, ErrNetwork, "create http request: %s %s failed", c.method, url)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.WrapC(err, ErrNetwork, "send http request: %v failed", req)
	}

	if resp.StatusCode != http.StatusOK {
		var rpcErr RpcError
		if err = json.NewDecoder(resp.Body).Decode(&rpcErr); err != nil {
			return nil, errors.WrapC(err, ErrDecodingJSON, "decode http response failed")
		}
		rpcErr.HTTP = resp.StatusCode
		return nil, errors.Wrapf(rpcErr, "call downstream: %s error", url)
	}

	return resp, nil
}
