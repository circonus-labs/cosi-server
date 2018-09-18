// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexcesaro/statsd"
	"github.com/rs/zerolog"
)

func TestIndex(t *testing.T) {
	t.Log("Testing / handler")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	c, _ := statsd.New()
	s := &Server{info: "foobar", stats: c}
	handler := s.index()

	tt := []struct {
		method string
		path   string
		status int
		msg    string
	}{
		{"GET", "/foo", http.StatusNotFound, "Not Found"},
		{"POST", "/", http.StatusMethodNotAllowed, "Method Not Allowed"},
		{"GET", "/", http.StatusOK, "foobar"},
	}

	for _, tst := range tt {
		t.Logf("\t%s %s", tst.method, tst.path)

		req := httptest.NewRequest(tst.method, "http://cosi"+tst.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tst.status {
			t.Fatalf("expected %d, got %d %s", tst.status, resp.StatusCode, http.StatusText(resp.StatusCode))
		}

		if !bytes.Contains(body, []byte(tst.msg)) {
			t.Fatalf("body missing '%s' (%s)", tst.msg, string(body))
		}
	}
}
