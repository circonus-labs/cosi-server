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
	"regexp"
	"testing"

	"github.com/alexcesaro/statsd"
	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/circonus-labs/cosi-server/internal/config/defaults"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func TestBroker(t *testing.T) {
	t.Log("Testing broker handler")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	pullList := defaults.BrokerPullList
	pullIdx := defaults.BrokerPullDefault
	pullID := pullList[pullIdx]

	pushList := defaults.BrokerPushList
	pushIdx := defaults.BrokerPushDefault
	pushID := pushList[pushIdx]

	viper.Set(config.KeyBrokerPullList, pullList)
	viper.Set(config.KeyBrokerPullDefault, pullIdx)
	viper.Set(config.KeyBrokerPushList, pushList)
	viper.Set(config.KeyBrokerPushDefault, pushIdx)
	viper.Set(config.KeyBrokerFallbackList, defaults.BrokerFallbackList)
	viper.Set(config.KeyBrokerFallbackDefault, defaults.BrokerFallbackDefault)

	type bid struct {
		BID string `json:"broker_id"`
	}

	c, _ := statsd.New()
	s := &Server{
		modepullrx: regexp.MustCompile(defaults.AgentPullModeRx),
		modepushrx: regexp.MustCompile(defaults.AgentPushModeRx),
		stats:      c,
	}
	handler := s.broker()

	tt := []struct {
		method string
		path   string
		status int
		msg    string
	}{
		{"GET", "/broker", http.StatusNotFound, "Not Found"},
		{"POST", "/broker/", http.StatusMethodNotAllowed, "Method Not Allowed"},
		{"GET", "/broker/", http.StatusBadRequest, "agent_mode required"},
		{"GET", "/broker/?agent_mode=invalid", http.StatusBadRequest, "invalid agent_mode"},
		{"GET", "/broker/?agent_mode=reverse", http.StatusOK, "broker_id"},
		{"GET", "/broker/?agent=reverse", http.StatusOK, "broker_id"},
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

	tt2 := []struct {
		method string
		path   string
		status int
		val    string
	}{
		{"GET", "/broker/?agent_mode=reverse", http.StatusOK, pullID},
		{"GET", "/broker/?agent=reverse", http.StatusOK, pullID},
		{"GET", "/broker/?agent_mode=revonly", http.StatusOK, pullID},
		{"GET", "/broker/?agent=revonly", http.StatusOK, pullID},
		{"GET", "/broker/?agent_mode=pull", http.StatusOK, pullID},
		{"GET", "/broker/?agent=pull", http.StatusOK, pullID},
		{"GET", "/broker/?agent_mode=json", http.StatusOK, pullID},
		{"GET", "/broker/?agent=json", http.StatusOK, pullID},
		{"GET", "/broker/?agent_mode=push", http.StatusOK, pushID},
		{"GET", "/broker/?agent=push", http.StatusOK, pushID},
		{"GET", "/broker/?agent_mode=trap", http.StatusOK, pushID},
		{"GET", "/broker/?agent=trap", http.StatusOK, pushID},
		{"GET", "/broker/?agent_mode=httptrap", http.StatusOK, pushID},
		{"GET", "/broker/?agent=httptrap", http.StatusOK, pushID},
	}

	for _, tst := range tt2 {
		t.Logf("\t%s %s", tst.method, tst.path)

		req := httptest.NewRequest(tst.method, "http://cosi"+tst.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != tst.status {
			t.Fatalf("expected %d, got %d %s", tst.status, resp.StatusCode, http.StatusText(resp.StatusCode))
		}

		var v bid
		if err := json.Unmarshal(body, &v); err != nil {
			t.Fatalf("expected NO error, got %v", err)
		}

		if v.BID != tst.val {
			t.Fatalf("expected %s got %s", tst.val, v.BID)
		}
	}

}
