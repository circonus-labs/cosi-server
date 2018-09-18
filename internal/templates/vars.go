// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package templates

import (
	"regexp"

	"github.com/rs/zerolog"
)

// Templates represents the various templates available to cosi
type Templates struct {
	// there aren't thousands of templates for the default agent(s).
	// additionally, the templates themselves are not overly large.
	// the content of the templates will be cached in ready-to-serve
	// TOML format.
	useCache    bool
	cache       map[string][]byte
	index       map[string]string // mapping of keys to cache entries
	logger      zerolog.Logger
	templateDir string
	Typerx      *regexp.Regexp
	Namerx      *regexp.Regexp
	fileExt     string // template file extension
}

type tinfo struct {
	key      string
	filename string
}

type tspec struct {
	ostype  string
	osdist  string
	osvers  string
	sysarch string
	ttype   string
	tname   string
}
