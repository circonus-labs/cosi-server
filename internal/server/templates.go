// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/hlog"
	"github.com/xi2/httpgzip"
)

func (s *Server) template() http.Handler {
	return httpgzip.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.URL.Path, "/template/") {
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

				tinfo, err := s.validateTemplateSpec(r)
				if err != nil {
					hlog.FromRequest(r).Error().Err(err).Msg("invalid template specification")
					s.stats.Increment(fmt.Sprintf("%s`%d`spec", r.URL.Path, http.StatusBadRequest))
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusBadRequest))
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				args, err := s.validateRequiredParams(r)
				if err != nil {
					hlog.FromRequest(r).Error().Err(err).Msg("invalid parameter")
					s.stats.Increment(fmt.Sprintf("%s`%d`params", r.URL.Path, http.StatusBadRequest))
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusBadRequest))
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				t, err := s.templates.Get(args.osType, args.osDistro, args.osVers, args.sysArch, tinfo.Type, tinfo.Name)
				if err != nil {
					if strings.Contains(err.Error(), "no template found") {
						hlog.FromRequest(r).Warn().Err(err).Msg("fetching template")
						s.stats.Increment(fmt.Sprintf("%s`%d`no_template_found", r.URL.Path, http.StatusNotFound))
						http.Error(w, err.Error(), http.StatusNotFound)
					} else {
						hlog.FromRequest(r).Error().Err(err).Msg("fetching template")
						s.stats.Increment(fmt.Sprintf("%s`%d`%s", r.URL.Path, http.StatusInternalServerError, err.Error()))
						s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusInternalServerError))
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}

				w.Header().Set("Content-Type", s.templateContentType)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(*t) // binary write, so % used in strings is not interpolated
				s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusOK))
			}),
		nil)
}
