package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestProxy_Proxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("x-foo"); v != "foo" {
			t.Errorf("got x-foo %q; expected %q", v, "foo")
		}
		if v := r.Header.Get("x-bar"); v != "" {
			t.Errorf("got x-bar %q; expected %q", v, "nil")
		}
		w.Header().Set("x-rt-foo", "foo")
		w.Header().Set("x-rt-bar", "bar")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer backend.Close()
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	proxy, _ := NewProxy(&Config{
		Host:                fmt.Sprintf("%s://%s", backendURL.Scheme, backendURL.Host),
		KeepRequestHeaders:  []string{"x-foo"},
		KeepResponseHeaders: []string{"x-rt-foo"},
	})
	frontend := httptest.NewServer(http.HandlerFunc(proxy.Proxy))
	defer frontend.Close()
	frontendClient := frontend.Client()

	getReq, _ := http.NewRequest("GET", frontend.URL, nil)
	getReq.Close = true
	getReq.Header.Set("x-foo", "foo")
	getReq.Header.Set("x-bar", "bar")
	res, err := frontendClient.Do(getReq)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if g, e := res.StatusCode, http.StatusOK; g != e {
		t.Errorf("got res.StatusCode %d; expected %d", g, e)
	}
	if g, e := res.Header.Get("x-rt-foo"), "foo"; g != e {
		t.Errorf("got x-rt-foo %q; expected %q", g, e)
	}
	if g, e := res.Header.Get("x-rt-bar"), ""; g != e {
		t.Errorf("got x-rt-bar %q; expected %q", g, e)
	}
	bodyBytes, _ := io.ReadAll(res.Body)
	if g, e := string(bodyBytes), `{"code":0,"msg":"ok"}`; g != e {
		t.Errorf("got body %q; expected %q", g, e)
	}
}

func TestProxy_Proxy_PreRequests(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("x-foo"); v != "foo" {
			t.Errorf("got x-foo %q; expected %q", v, "foo")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer backend.Close()
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	proxy, _ := NewProxy(&Config{
		Host: fmt.Sprintf("%s://%s", backendURL.Scheme, backendURL.Host),
	})
	frontend := httptest.NewServer(http.HandlerFunc(proxy.Before(func(req *http.Request) {
		req.Header.Set("x-foo", "foo")
	}).Proxy))
	defer frontend.Close()
	frontendClient := frontend.Client()

	getReq, _ := http.NewRequest("GET", frontend.URL, nil)
	getReq.Close = true
	res, err := frontendClient.Do(getReq)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if g, e := res.StatusCode, http.StatusOK; g != e {
		t.Errorf("got res.StatusCode %d; expected %d", g, e)
	}
	bodyBytes, _ := io.ReadAll(res.Body)
	if g, e := string(bodyBytes), `{"code":0,"msg":"ok"}`; g != e {
		t.Errorf("got body %q; expected %q", g, e)
	}
}

func TestProxy_Proxy_PreResponse(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer backend.Close()
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	proxy, _ := NewProxy(&Config{
		Host: fmt.Sprintf("%s://%s", backendURL.Scheme, backendURL.Host),
	})
	frontend := httptest.NewServer(http.HandlerFunc(proxy.After(func(resp *http.Response) error {
		resp.Header.Set("x-rt-foo", "foo")
		return nil
	}).Proxy))
	defer frontend.Close()
	frontendClient := frontend.Client()

	getReq, _ := http.NewRequest("GET", frontend.URL, nil)
	getReq.Close = true
	res, err := frontendClient.Do(getReq)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if g, e := res.StatusCode, http.StatusOK; g != e {
		t.Errorf("got res.StatusCode %d; expected %d", g, e)
	}
	if g, e := res.Header.Get("x-rt-foo"), "foo"; g != e {
		t.Errorf("got x-rt-foo %q; expected %q", g, e)
	}
	bodyBytes, _ := io.ReadAll(res.Body)
	if g, e := string(bodyBytes), `{"code":0,"msg":"ok"}`; g != e {
		t.Errorf("got body %q; expected %q", g, e)
	}
}
