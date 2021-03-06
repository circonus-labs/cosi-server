// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/viper"
	"github.com/xi2/httpgzip"
)

func (s *Server) rpm() http.Handler {
	return httpgzip.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/install/rpm/" {
					hlog.FromRequest(r).Error().Msg("not found")
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusNotFound))
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				if r.Method != "GET" {
					hlog.FromRequest(r).Error().Str("method", r.Method).Msg("invalid method")
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusMethodNotAllowed))
					http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
					return
				}

				rpmFile := viper.GetString(config.KeyRPMFile)

				w.Header().Set("Content-Type", "application/x-redhat-package-manager")
				w.Header().Set("Cache-Control", "no-cache, must-revalidate")
				w.Header().Set("Pragma", "no-cache")

				f, err := os.Open(path.Join(viper.GetString(config.KeyContentPath), "files", rpmFile))
				if err != nil {
					hlog.FromRequest(r).Error().Err(err).Msg(rpmFile)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				http.ServeContent(w, r, rpmFile, time.Now(), f)
			}),
		nil)
}
