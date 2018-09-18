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

func TestPackage(t *testing.T) {
	t.Log("Testing Package")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "CentOS") {
			w.Write([]byte(`{"package_url":"http://updates.circonus.net/node-agent/packages/","package_file":"nad-omnibus-2.6.0-1.el7.x86_64.rpm"}`))
		} else if strings.Contains(r.URL.String(), "OmniOS") {
			w.Write([]byte(`{"publisher_url":"http://updates.circonus.net/omnios/r151014/","publisher_name":"circonus","package_name":"field/nad"}`))
		} else if strings.Contains(r.URL.String(), "FreeBSD") {
			w.Write([]byte(`{"package_url":"http://updates.circonus.net/node-agent/packages/","package_file":"nad-omnibus-2.6.0-freebsd.11.0-amd64.tar.gz"}`))
		} else {
			w.Write([]byte("invalid"))
		}
	}))

	tests := []struct {
		name        string
		format      string
		cfg         Config
		shouldError bool
		errorExpect error
	}{
		{"invalid (format)", "foo", Config{OSType: "test", OSDistro: "test", OSVersion: "test", SysArch: "test"}, true, errors.New("invalid format (foo)")},
		{"valid (centos)", "", Config{OSType: "Linux", OSDistro: "CentOS", OSVersion: "7.1.1408", SysArch: "x86_64"}, false, nil},
		{"valid (omnios)", "json", Config{OSType: "Solaris", OSDistro: "OmniOS", OSVersion: "r151014", SysArch: "amd64"}, false, nil},
		{"valid (freebsd)", "text", Config{OSType: "BSD", OSDistro: "FreeBSD", OSVersion: "11.0", SysArch: "amd64"}, false, nil},
	}

	for _, test := range tests {
		t.Logf("\t%s", test.name)

		test.cfg.CosiURL = ts.URL
		c, err := New(&test.cfg)
		if err != nil {
			t.Fatalf("unexpected error (%s)", err)
		}

		_, err = c.FetchPackage(test.format)
		if test.shouldError {
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != test.errorExpect.Error() {
				t.Fatalf("unexpected error (%s)", err)
			}
			continue
		}

		if err != nil {
			t.Fatalf("unexpected error (%s)", err)
		}
	}
}
