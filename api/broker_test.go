// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchBroker(t *testing.T) {
	t.Log("Testing FetchBroker")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "agent_mode=httptrap") {
			w.Write([]byte(`{"broker_id":"httptrap"}`))
		} else if strings.Contains(r.URL.String(), "agent_mode=push") {
			w.Write([]byte(`{"broker_id":"httptrap"}`))
		} else if strings.Contains(r.URL.String(), "agent_mode=trap") {
			w.Write([]byte(`{"broker_id":"httptrap"}`))
		} else if strings.Contains(r.URL.String(), "agent_mode=json") {
			w.Write([]byte(`{"broker_id":"json"}`))
		} else if strings.Contains(r.URL.String(), "agent_mode=pull") {
			w.Write([]byte(`{"broker_id":"json"}`))
		} else if strings.Contains(r.URL.String(), "agent_mode=reverse") {
			w.Write([]byte(`{"broker_id":"json"}`))
		} else if strings.Contains(r.URL.String(), "agent_mode=error") {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			w.Write([]byte("invalid"))
		}
	}))

	tests := []struct {
		name        string
		ctype       string
		shouldError bool
		errorExpect error
	}{
		{"invalid (empty)", "", true, errors.New("invalid check type (empty)")},
		{"invalid (error)", "error", true, errors.New("fetching broker: 500 Internal Server Error - " + ts.URL + "/broker/?agent_mode=error - Internal Server Error")},
		{"valid (json)", "json", false, nil},
		{"valid (pull)", "pull", false, nil},
		{"valid (reverse)", "reverse", false, nil},
		{"valid (trap)", "httptrap", false, nil},
		{"valid (trap)", "trap", false, nil},
		{"valid (push)", "push", false, nil},
	}

	cfg := &Config{
		OSType:    "Linux",
		OSDistro:  "CentOS",
		OSVersion: "7.1.1408",
		SysArch:   "x86_64",
		CosiURL:   ts.URL,
	}

	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			c, err := New(cfg)
			if err != nil {
				t.Fatalf("unexpected error (%s)", err)
			}

			_, err = c.FetchBroker(tst.ctype)
			if tst.shouldError {
				if err == nil {
					t.Fatal("expected error")
				}
				if err.Error() != tst.errorExpect.Error() {
					t.Fatalf("unexpected error (%s) [%s]", err, tst.errorExpect)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error (%s)", err)
				}
			}
		})
	}
}
