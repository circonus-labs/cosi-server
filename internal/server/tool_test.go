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

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/circonus-labs/cosi-server/internal/config/defaults"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func TestTool(t *testing.T) {
	t.Log("Testing tool handler")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	// TODO: update tests to include now required query args for
	//       type, dist, vers, and arch

	viper.Set(config.KeyParamTypeRx, defaults.ParamTypeRx)
	viper.Set(config.KeyParamDistroRx, defaults.ParamDistroRx)
	viper.Set(config.KeyParamVersionRx, defaults.ParamVersionRx)
	viper.Set(config.KeyParamVersionCleanerRx, defaults.ParamVersionCleanerRx)
	viper.Set(config.KeyParamArchRx, defaults.ParamArchRx)
	viper.Set(config.KeyCosiToolVersion, "v0.0.0")
	s, err := New()
	if err != nil {
		t.Fatalf("expected NO error, got %v", err)
	}
	handler := s.tool()

	tt := []struct {
		method  string
		headers map[string]string
		path    string
		status  int
		msg     string
	}{
		{"GET", map[string]string{}, "/tool", http.StatusNotFound, "Not Found"},
		{"POST", map[string]string{}, "/tool/", http.StatusMethodNotAllowed, "Method Not Allowed"},
		{"GET", map[string]string{}, "/tool/", http.StatusBadRequest, "invalid system 'type' specified"},
		{"GET", map[string]string{}, "/tool/?type=Linux", http.StatusBadRequest, "invalid system 'dist' specified"},
		{"GET", map[string]string{}, "/tool/?type=Linux&dist=Ubuntu", http.StatusBadRequest, "invalid system 'vers' specified"},
		{"GET", map[string]string{}, "/tool/?type=Linux&dist=Ubuntu&vers=16.04", http.StatusBadRequest, "invalid system 'arch' specified"},
		{"GET", map[string]string{}, "/tool/?type=Linux&dist=Ubuntu&vers=16.04&arch=x86_64", http.StatusTemporaryRedirect, `/v0.0.0/cosi-tool_0.0.0_linux_x86_64.tar.gz`},
		{"GET", map[string]string{"Accept": "application/json"}, "/tool/?type=Linux&dist=Ubuntu&vers=16.04&arch=x86_64", http.StatusTemporaryRedirect, `/v0.0.0/cosi-tool_0.0.0_linux_x86_64.tar.gz`},
		{"GET", map[string]string{"Accept": "*/*"}, "/tool/?type=Linux&dist=Ubuntu&vers=16.04&arch=x86_64", http.StatusTemporaryRedirect, `/v0.0.0/cosi-tool_0.0.0_linux_x86_64.tar.gz`},
	}

	for _, tst := range tt {
		t.Logf("\t%s %s", tst.method, tst.path)

		req := httptest.NewRequest(tst.method, "http://cosi"+tst.path, nil)
		for key, val := range tst.headers {
			req.Header.Set(key, val)
		}
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
