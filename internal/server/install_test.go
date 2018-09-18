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

func TestInstall(t *testing.T) {
	t.Log("Testing install handler")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	viper.Set(config.KeyContentPath, "../../content")
	c, _ := statsd.New()
	s := &Server{logger: log.With().Str("pkg", "server").Logger(), stats: c}
	handler := s.install()

	tt := []struct {
		method string
		path   string
		status int
		msg    string
	}{
		{"GET", "/install", http.StatusNotFound, "Not Found"},
		{"POST", "/install/", http.StatusMethodNotAllowed, "Method Not Allowed"},
		{"GET", "/install/", http.StatusOK, "cosi-install --key <apikey>"},
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
