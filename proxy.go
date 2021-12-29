package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewSimpleProxy(host string) *Proxy {
	return &Proxy{config: &Config{Host: host}}
}

func NewProxy(c *Config) (*Proxy, error) {
	if !strings.HasPrefix(c.Host, "http://") && !strings.HasPrefix(c.Host, "https://") {
		return nil, errors.New("invalid proxy host")
	}
	p := &Proxy{config: c, keepRequestHeaders: []string{}, keepResponseHeaders: []string{}}
	p.keepRequestHeaders = append(p.keepRequestHeaders, c.KeepRequestHeaders...)
	p.keepResponseHeaders = append(p.keepResponseHeaders, c.KeepResponseHeaders...)
	return p, nil
}

type Config struct {
	Host                string            `json:"host" yaml:"host"`
	Timeout             int64             `json:"timeout" yaml:"timeout"`
	Mapping             map[string]string `json:"mapping" yaml:"mapping"`
	KeepRequestHeaders  []string          `json:"keepRequestHeaders" yaml:"keepRequestHeaders"`
	KeepResponseHeaders []string          `json:"keepResponseHeaders" yaml:"keepResponseHeaders"`
}

type PreRequest func(req *http.Request)
type PreResponse func(resp *http.Response) error

type Proxy struct {
	config              *Config
	keepRequestHeaders  []string
	keepResponseHeaders []string
	PreRequests         []PreRequest
	PreResponses        []PreResponse
}

func (p *Proxy) Before(fns ...PreRequest) *Proxy {
	if len(fns) == 0 {
		return p
	}
	np := *p
	if np.PreRequests == nil {
		np.PreRequests = make([]PreRequest, 0)
	}
	np.PreRequests = append(np.PreRequests, fns...)
	return &np
}

func (p *Proxy) After(fns ...PreResponse) *Proxy {
	if len(fns) == 0 {
		return p
	}
	np := *p
	if np.PreResponses == nil {
		np.PreResponses = make([]PreResponse, 0)
	}
	np.PreResponses = append(np.PreResponses, fns...)
	return &np
}

func (p *Proxy) url(uri string) string {
	if p.config.Host == "" {
		return uri
	}
	uri = strings.TrimLeft(uri, "/")
	if _uri, ok := p.config.Mapping[uri]; ok {
		if strings.HasPrefix(_uri, "http://") || strings.HasPrefix(_uri, "https://") {
			return _uri
		}
		uri = strings.TrimLeft(_uri, "/")
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(p.config.Host, "/"), uri)
}

func (p *Proxy) Proxy(resp http.ResponseWriter, req *http.Request) {
	uri := p.url(req.RequestURI)
	u, e := url.Parse(uri)
	if e != nil || u.Host == "" {
		resp.WriteHeader(http.StatusNotFound)
		_, _ = resp.Write([]byte(`parse url error`))
		return
	}

	pxy := httputil.NewSingleHostReverseProxy(u)
	pxy.Director = func(req *http.Request) {
		req.URL = u
		req.RequestURI = u.RequestURI()
		req.Host = u.Host

		if len(p.keepRequestHeaders) > 0 {
			for k, _ := range req.Header {
				if !inArray(p.keepRequestHeaders, k) {
					req.Header.Del(k)
				}
			}
		}

		if len(p.PreRequests) > 0 {
			for _, fn := range p.PreRequests {
				fn(req)
			}
		}
	}

	pxy.ModifyResponse = func(resp *http.Response) error {
		if len(p.keepResponseHeaders) > 0 {
			for k, _ := range resp.Header {
				if !inArray(p.keepResponseHeaders, k) {
					resp.Header.Del(k)
				}
			}
		}

		if len(p.PreResponses) > 0 {
			for _, fn := range p.PreResponses {
				if e := fn(resp); e != nil {
					return e
				}
			}
		}
		return nil
	}

	pxy.ServeHTTP(resp, req)
}

func inArray(data []string, v string) bool {
	for _, datum := range data {
		if strings.ToLower(datum) == strings.ToLower(v) {
			return true
		}
	}
	return false
}
