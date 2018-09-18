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
	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func TestConfig(t *testing.T) {
	t.Log("Testing config handler")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	viper.Set(config.KeyContentPath, "../../content")
	c, _ := statsd.New()
	s := &Server{logger: log.With().Str("pkg", "server").Logger(), stats: c}
	handler := s.config()

	tt := []struct {
		method string
		path   string
		status int
		msg    string
	}{
		{"GET", "/install/config", http.StatusNotFound, "Not Found"},
		{"POST", "/install/config/", http.StatusMethodNotAllowed, "Method Not Allowed"},
		{"GET", "/install/config/", http.StatusOK, `#cosi_api_key=""`},
		{"GET", "/install/conf", http.StatusNotFound, "Not Found"},
		{"POST", "/install/conf/", http.StatusMethodNotAllowed, "Method Not Allowed"},
		{"GET", "/install/conf/", http.StatusOK, `#cosi_api_key=""`},
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
