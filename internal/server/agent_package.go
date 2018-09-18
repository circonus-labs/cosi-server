// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/hlog"
	"github.com/xi2/httpgzip"
)

func (s *Server) agentPackage() http.Handler {
	return httpgzip.NewHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/package/" {
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

				pkg, err := s.packageList.GetPackageInfo(args.osType, args.osDistro, args.osVers, args.sysArch)
				if err != nil {
					hlog.FromRequest(r).Error().Err(err).Interface("args", args).Msg("unsupported os")
					// generic unsupported metric
					s.stats.Increment(fmt.Sprintf("%s`%d`unsupported", r.URL.Path, http.StatusNotFound))
					// full unsupported metric with what was requested
					s.stats.Increment(fmt.Sprintf("%s`%s`%s`%s`%s`%d", r.URL.Path, args.osType, args.osDistro, args.osVers, args.sysArch, http.StatusNotFound))
					// unsupported with reason (error)
					s.stats.Increment(fmt.Sprintf("%s`%s`%d", r.URL.Path, err.Error(), http.StatusNotFound))
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}

				// handle redirect
				if _, ok := r.URL.Query()["redirect"]; ok {
					if pkg.URL != "" && pkg.File != "" {
						pkgURL := pkg.URL + pkg.File
						s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusTemporaryRedirect))
						s.stats.Increment(fmt.Sprintf("%s`%d`%s", r.URL.Path, http.StatusTemporaryRedirect, pkgURL))
						http.Redirect(w, r, pkgURL, http.StatusTemporaryRedirect)
						return
					}
				}

				w.Header().Set("Cache-Control", "private, max-age=300")

				if accept := r.Header.Get("Accept"); accept == "*/*" || accept == "application/json" {
					// if the client can handle json, send back package info structure
					data, err := json.Marshal(pkg)
					if err != nil {
						hlog.FromRequest(r).Error().Err(err).Interface("pkg", pkg).Msg("json encoding")
						s.stats.Increment(fmt.Sprintf("%s`%d`encode_err", r.URL.Path, http.StatusInternalServerError))
						s.stats.Increment(fmt.Sprintf("%s`%s`%d", r.URL.Path, err.Error(), http.StatusInternalServerError))
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, string(data))
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusOK))
					return
				}

				sep := "%%" // the text returned is parsed by cosi-install, a bash script
				w.Header().Set("Content-Type", "text/plain")
				// NOTE: do **not** return a line ending with result string - the script is
				//       parsing a compound result from request plus status code from curl.
				if s.solarisrx.MatchString(args.osDistro) {
					// pkg based os - package name, publishser, publisher url
					fmt.Fprintf(w, "%s%s%s%s%s", pkg.Name, sep, pkg.PubName, sep, pkg.PubURL)
					s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusOK))
					return
				}
				// rpm or deb - package url and package file name
				fmt.Fprintf(w, "%s%s%s", pkg.URL, sep, pkg.File)
				s.stats.Increment(fmt.Sprintf("%s`%d", r.URL.Path, http.StatusOK))
			}),
		nil)
}
