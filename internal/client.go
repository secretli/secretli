package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL = "https://patrickscheid.de/s/"
	userAgent      = "secretli-cli"
)

type Client struct {
	client    *http.Client
	baseUrl   *url.URL
	userAgent string
}

func NewClient(options ...ClientOptionFunc) (*Client, error) {
	client := &Client{
		client:    &http.Client{},
		userAgent: userAgent,
	}

	if err := client.setBaseUrl(defaultBaseURL); err != nil {
		return nil, err
	}

	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *Client) BaseUrl() *url.URL {
	u := *c.baseUrl
	return &u
}

func (c *Client) setBaseUrl(urlStr string) error {
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	c.baseUrl = baseURL
	return nil
}

func (c *Client) NewRequest(method string, path string, body interface{}) (*http.Request, error) {
	unescaped, err := url.PathUnescape(path)
	if err != nil {
		return nil, err
	}

	u := c.BaseUrl()
	u.RawPath = c.baseUrl.Path + path
	u.Path = c.baseUrl.Path + unescaped

	reqHeaders := make(http.Header)
	reqHeaders.Set("Accept", "application/json")
	reqHeaders.Set("User-Agent", c.userAgent)

	var bodyReader io.Reader = nil
	if body != nil {
		reqHeaders.Set("Content-Type", "application/json")
		requestBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(requestBytes)
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range reqHeaders {
		req.Header[k] = v
	}

	return req, nil
}

func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}

	return resp, nil
}

type ErrorResponse struct {
	Body     []byte
	Response *http.Response
}

func (e *ErrorResponse) Error() string {
	path, _ := url.QueryUnescape(e.Response.Request.URL.Path)
	u := fmt.Sprintf("%s://%s%s", e.Response.Request.URL.Scheme, e.Response.Request.URL.Host, path)
	return fmt.Sprintf("%s %s: %d", e.Response.Request.Method, u, e.Response.StatusCode)
}

func checkResponse(r *http.Response) error {
	switch r.StatusCode {
	case 200, 201, 202, 204, 304:
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorResponse.Body = data
	}
	return errorResponse
}
