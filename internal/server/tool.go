// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/viper"
	"github.com/xi2/httpgzip"
)

func (s *Server) tool() http.Handler {
	return httpgzip.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/tool/" && r.URL.Path != "/utils/" {
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

				args, err := s.validateRequiredParams(r)
				if err != nil {
					hlog.FromRequest(r).Error().Err(err).Msg("invalid parameter")
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusBadRequest))
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				//
				// tracking metrics
				//
				// full
				s.stats.Increment(fmt.Sprintf("%s`%s`%s`%s`%s", r.URL.Path, args.osType, args.osDistro, args.osVers, args.sysArch))
				// os type
				s.stats.Increment(fmt.Sprintf("%s`%s", r.URL.Path, args.osType))
				// system arch
				s.stats.Increment(fmt.Sprintf("%s`%s", r.URL.Path, args.sysArch))
				// os dist
				s.stats.Increment(fmt.Sprintf("%s`%s", r.URL.Path, args.osDistro))
				// os dist arch
				s.stats.Increment(fmt.Sprintf("%s`%s`%s", r.URL.Path, args.osDistro, args.sysArch))
				// os dist ver
				s.stats.Increment(fmt.Sprintf("%s`%s`%s", r.URL.Path, args.osDistro, args.osVers))
				// os dist ver arch
				s.stats.Increment(fmt.Sprintf("%s`%s`%s`%s", r.URL.Path, args.osDistro, args.osVers, args.sysArch))

				// example of expected redirect url:
				// https://github.com/circonus-labs/cosi-tool/releases/download/v0.2.0/cosi-tool_0.2.0_linux_x86_64.tar.gz
				//
				cosiVer := viper.GetString(config.KeyCosiToolVersion)
				redirURL := fmt.Sprintf("%s/%s/cosi-tool_%s_%s_x86_64.tar.gz",
					viper.GetString(config.KeyCosiToolBaseURL),
					cosiVer,
					strings.Replace(cosiVer, "v", "", 1),
					args.osType)

				http.Redirect(w, r, redirURL, http.StatusTemporaryRedirect)
			}),
		nil)
}
