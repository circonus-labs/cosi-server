// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInfo(t *testing.T) {
	t.Log("invalid (json/parse)")
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("invalid"))
		}))
		cfg := &Config{
			OSType:    "test",
			OSDistro:  "test",
			OSVersion: "test",
			SysArch:   "test",
			CosiURL:   ts.URL,
		}

		c, err := New(cfg)
		if err != nil {
			t.Fatalf("unexpected error (%s)", err)
		}

		if _, err := c.FetchInfo(); err == nil {
			t.Fatal("expected error")
		} else if err.Error() != "parsing server info: invalid character 'i' looking for beginning of value" {
			t.Fatalf("expected no error got (%s)", err)
		}
		ts.Close()
	}

	t.Log("valid")
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"description":"test","version":"test","supported":["foo","bar"]}`))
		}))
		cfg := &Config{
			OSType:    "test",
			OSDistro:  "test",
			OSVersion: "test",
			SysArch:   "test",
			CosiURL:   ts.URL,
		}

		c, err := New(cfg)
		if err != nil {
			t.Fatalf("unexpected error (%s)", err)
		}

		if _, err := c.FetchInfo(); err != nil {
			t.Fatalf("expected no error got (%s)", err)
		}
		ts.Close()
	}
}
