// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/circonus-labs/cosi-server/internal/config/defaults"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func TestAgentPackage(t *testing.T) {
	t.Log("Testing agentPackage")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	viper.Set(config.KeyParamTypeRx, defaults.ParamTypeRx)
	viper.Set(config.KeyParamDistroRx, defaults.ParamDistroRx)
	viper.Set(config.KeyParamVersionRx, defaults.ParamVersionRx)
	viper.Set(config.KeyParamVersionCleanerRx, defaults.ParamVersionCleanerRx)
	viper.Set(config.KeyParamArchRx, defaults.ParamArchRx)
	viper.Set(config.KeyPackageBaseURL, defaults.BasePackageURL)
	viper.Set(config.KeyIsRHELDistroRx, defaults.IsRHELDistroRx)
	viper.Set(config.KeyIsSolarisDistroRx, defaults.IsSolarisDistroRx)
	viper.Set(config.KeyPackageConfigFile, "../packages/testdata/valid.yaml")
	viper.Set(config.KeyContentPath, "../templates/testdata")
	s, err := New()
	if err != nil {
		t.Fatalf("expected NO error, got %v", err)
	}

	handler := s.agentPackage()

	tt := []struct {
		method  string
		headers map[string]string
		path    string
		status  int
		msg     string
	}{
		{"GET", map[string]string{}, "/package", http.StatusNotFound, "Not Found"},
		{"POST", map[string]string{}, "/package/", http.StatusMethodNotAllowed, "Method Not Allowed"},
		{"GET", map[string]string{}, "/package/", http.StatusBadRequest, "invalid system 'type' specified"},
		{"GET", map[string]string{}, "/package/?type=Linux", http.StatusBadRequest, "invalid system 'dist' specified"},
		{"GET", map[string]string{}, "/package/?type=Linux&dist=Ubuntu", http.StatusBadRequest, "invalid system 'vers' specified"},
		{"GET", map[string]string{}, "/package/?type=Linux&dist=Ubuntu&vers=16.04", http.StatusBadRequest, "invalid system 'arch' specified"},
		{"GET", map[string]string{}, "/package/?type=Linux&dist=Ubuntu&vers=16.04&arch=x86_64", http.StatusOK, `/%%nad-omnibus`},
		{"GET", map[string]string{}, "/package/?type=Linux&dist=Ubuntu&vers=16.04&arch=x86_64&redirect", http.StatusTemporaryRedirect, `node-agent/packages/nad-omnibus`},
		{"GET", map[string]string{"Accept": "application/json"}, "/package/?type=Linux&dist=Ubuntu&vers=16.04&arch=x86_64", http.StatusOK, ""},
		{"GET", map[string]string{"Accept": "*/*"}, "/package/?type=Linux&dist=Ubuntu&vers=16.04&arch=x86_64", http.StatusOK, ""},
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

		if resp.Header.Get("Accept") == "application/json" {
			var x map[string]interface{}
			err := json.Unmarshal(body, &x)
			if err != nil {
				t.Fatalf("expected NO error, got %v", err)
			}
		} else {
			if !bytes.Contains(body, []byte(tst.msg)) {
				t.Fatalf("body missing '%s' (%s)", tst.msg, string(body))
			}
		}
	}
}
