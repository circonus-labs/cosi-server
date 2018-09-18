// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGet(t *testing.T) {
	t.Log("Testing get")

	cfg := &Config{
		OSType:    "test",
		OSDistro:  "test",
		OSVersion: "test",
		SysArch:   "test",
	}

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error (%s)", err)
	}

	t.Log("invalid (nil request url)")
	{
		_, err := c.get(nil, nil)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "invalid request url (nil)" {
			t.Fatalf("unexpected error (%s)", err)
		}
	}

	t.Log("invalid (empty request url)")
	{
		u := url.URL{}
		_, err := c.get(&u, nil)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "invalid request url (empty)" {
			t.Fatalf("unexpected error (%s)", err)
		}
	}

	t.Log("invalid (invalid url)")
	{
		u, err := url.Parse("http://not_a_host.foo/")
		if err != nil {
			t.Fatalf("unexpected error (%s)", err)
		}

		if _, err := c.get(u, nil); err == nil {
			t.Fatal("expected error")
		} else if err.Error() != "cosi-server request: Get http://not_a_host.foo/: dial tcp: lookup not_a_host.foo: no such host" {
			t.Fatalf("unexpected error (%s)", err)
		}
	}

	t.Log("invalid (not found)")
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}))
		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("unexpected error (%s)", err)
		}

		if _, err := c.get(u, nil); err == nil {
			t.Fatal("expected error")
		} else if err.Error() != "404 Not Found - "+ts.URL+" - Not Found" {
			t.Fatalf("unexpected error (%s)", err)
		}
		ts.Close()
	}

	t.Log("valid")
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("valid"))
		}))
		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("unexpected error (%s)", err)
		}

		if _, err := c.get(u, nil); err != nil {
			t.Fatalf("expected no error got (%s)", err)
		}
		ts.Close()
	}
}
