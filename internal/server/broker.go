// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/viper"
	"github.com/xi2/httpgzip"
)

const noValidBrokerFound = -2

func (s *Server) broker() http.Handler {
	return httpgzip.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/broker/" {
					hlog.FromRequest(r).Error().Msg("not found")
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusNotFound))
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				if r.Method != http.MethodGet {
					hlog.FromRequest(r).Error().Str("method", r.Method).Msg("invalid method")
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusMethodNotAllowed))
					http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
					return
				}

				// validate the agent mode requested
				args := r.URL.Query()
				mode := ""
				if m, ok := args["agent_mode"]; ok {
					mode = m[0]
				}
				if mode == "" {
					// use old 'agent' argument
					// TODO: deprecate 'agent' argument after initial release
					if m, ok := args["agent"]; ok {
						mode = m[0]
					}
				}
				if mode == "" {
					hlog.FromRequest(r).Error().Str("path", r.URL.Path).Str("query", r.URL.RawQuery).Msg("invalid parameters")
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusBadRequest))
					http.Error(w, "agent_mode required", http.StatusBadRequest)
					return
				}
				if !s.modepullrx.MatchString(mode) && !s.modepushrx.MatchString(mode) {
					hlog.FromRequest(r).Error().Str("path", r.URL.Path).Str("query", r.URL.RawQuery).Msg("invalid parameters")
					s.stats.Increment(fmt.Sprintf("%s`%d`invalid_mode", r.URL.Path, http.StatusBadRequest))
					http.Error(w, "invalid agent_mode", http.StatusBadRequest)
					return
				}

				brokerID, err := s.selectBroker(mode)
				if err != nil {
					hlog.FromRequest(r).Error().Err(err).Str("path", r.URL.Path).Str("mode", mode).Msg("broker selection error")
					s.stats.Increment(fmt.Sprintf("%s`%d`select_err", r.URL.Path, http.StatusInternalServerError))
					s.stats.Increment(fmt.Sprintf("%s`%s`%d", r.URL.Path, err.Error(), http.StatusInternalServerError))
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				if brokerID == noValidBrokerFound { // give up...
					hlog.FromRequest(r).Error().Err(err).Str("path", r.URL.Path).Str("mode", mode).Msg("no broker found")
					s.stats.Increment(fmt.Sprintf("%s`%d`no_broker_found", r.URL.Path, http.StatusNotFound))
					s.stats.Increment(fmt.Sprintf("%s`%s`%d", r.URL.Path, mode, http.StatusNotFound))
					http.Error(w, "unable to identify valid broker", http.StatusNotFound)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, fmt.Sprintf(`{"broker_id": "%d"}`, brokerID))
				s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusOK))
				s.stats.Increment(fmt.Sprintf("%s`%s`%d", r.URL.Path, mode, http.StatusOK))
			}),
		nil)
}

func (s *Server) selectBroker(mode string) (int64, error) {
	if s.modepullrx.MatchString(mode) {
		id, err := getBrokerID(
			viper.GetStringSlice(config.KeyBrokerPullList),
			viper.GetInt(config.KeyBrokerPullDefault))
		if err != nil {
			return noValidBrokerFound, errors.Wrap(err, "pull mode")
		}
		return id, nil
	}

	if s.modepushrx.MatchString(mode) {
		id, err := getBrokerID(
			viper.GetStringSlice(config.KeyBrokerPushList),
			viper.GetInt(config.KeyBrokerPushDefault))
		if err != nil {
			return noValidBrokerFound, errors.Wrap(err, "push mode")
		}
		return id, nil
	}

	// try the fallback broker config
	id, err := getBrokerID(
		viper.GetStringSlice(config.KeyBrokerFallbackList),
		viper.GetInt(config.KeyBrokerFallbackDefault))
	if err != nil {
		return noValidBrokerFound, errors.Wrap(err, "fallback mode")
	}

	return id, nil
}

func getBrokerID(list []string, defaultIdx int) (int64, error) {
	switch len(list) {
	case 0:
		return noValidBrokerFound, nil
	case 1:
		return strconv.ParseInt(list[0], 10, 32)
	}

	idx := defaultIdx
	if idx == -1 { // select random broker from list
		idx = rand.Intn(len(list))
	}
	if idx >= 0 && len(list) > idx {
		return strconv.ParseInt(list[idx], 10, 32)
	}
	return noValidBrokerFound, errors.Errorf("invalid index %d for list len %d", idx, len(list))
}
