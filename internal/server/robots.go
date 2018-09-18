// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/hlog"
	"github.com/xi2/httpgzip"
)

func (s *Server) robots() http.Handler {
	return httpgzip.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/robots.txt" {
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

				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "User-agent: *\nDisallow: /")
			}),
		nil)
}
