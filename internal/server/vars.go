// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"context"
	"net"
	"net/http"
	"regexp"

	"github.com/alexcesaro/statsd"
	"github.com/circonus-labs/cosi-server/internal/packages"
	"github.com/circonus-labs/cosi-server/internal/templates"
	"github.com/rs/zerolog"
)

// Server defines the listening servers
type Server struct {
	ctx                 context.Context
	logger              zerolog.Logger
	svrHTTP             []*httpServer
	svrHTTPS            *sslServer
	packageList         *packages.Packages
	templates           *templates.Templates
	info                string
	typerx              *regexp.Regexp
	distrx              *regexp.Regexp
	versrx              *regexp.Regexp
	vercleanrx          *regexp.Regexp
	archrx              *regexp.Regexp
	rhelrx              *regexp.Regexp
	solarisrx           *regexp.Regexp
	modepushrx          *regexp.Regexp
	modepullrx          *regexp.Regexp
	stats               *statsd.Client
	templateContentType string
}

type httpServer struct {
	address *net.TCPAddr
	server  *http.Server
}

type sslServer struct {
	address  *net.TCPAddr
	certFile string
	keyFile  string
	server   *http.Server
}

// serverInfo is returned for a / request
type serverInfo struct {
	Description string   `json:"description"`
	Supported   []string `json:"supported"`
	Version     string   `json:"version"`
}

// params holds validated query parameters
type params struct {
	osType   string
	osDistro string
	osVers   string
	sysArch  string
}

// templateSpec holds validated template information from url path
type templateSpec struct {
	Type string
	Name string
}
