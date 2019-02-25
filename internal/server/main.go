// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"context"
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alexcesaro/statsd"
	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/circonus-labs/cosi-server/internal/config/defaults"
	"github.com/circonus-labs/cosi-server/internal/packages"
	"github.com/circonus-labs/cosi-server/internal/release"
	"github.com/circonus-labs/cosi-server/internal/templates"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

func init() {
	// for random broker selection
	rand.Seed(time.Now().UnixNano())
}

// New creates a new instance of the listening server(s)
func New() (*Server, error) {
	s := Server{
		logger:              log.With().Str("pkg", "server").Logger(),
		templateContentType: "application/toml",
	}

	c, err := statsd.New(
		statsd.Address(viper.GetString(config.KeyStatsdAddress)),
		statsd.FlushPeriod(viper.GetDuration(config.KeyStatsdInterval)),
		statsd.Prefix(viper.GetString(config.KeyStatsdPrefix)))
	if err != nil {
		s.logger.Warn().Err(err).Msg("statsd metrics disabled")
	}
	s.stats = c

	if err := s.compileValidators(); err != nil {
		s.logger.Fatal().Err(err).Msg("initializing server")
		return nil, err
	}

	// load package definitions
	{
		p, err := packages.New("")
		if err != nil {
			return nil, errors.Wrap(err, "initializing package list")
		}

		s.packageList = p

		// server information is static, build it once for serving '/' (index) requests
		i := serverInfo{
			Description: "Circonus One Step Install Server",
			Supported:   p.ListSupported(),
			Version:     release.VERSION,
		}

		d, err := json.Marshal(i)
		if err != nil {
			return nil, errors.Wrap(err, "encoding server info")
		}

		s.info = string(d)

		if viper.GetBool(config.KeyLocalPackages) {
			if err := updateLocalPackageIndex(viper.GetString(config.KeyLocalPackagePath)); err != nil {
				return nil, errors.Wrap(err, "updating local package index")
			}
		}
	}

	// load templates
	{
		t, err := templates.New()
		if err != nil {
			return nil, errors.Wrap(err, "initializing templates")
		}
		s.templates = t
	}

	chain := alice.New()
	chain = chain.Append(hlog.NewHandler(s.logger))
	chain = chain.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		dur := uint64(duration / time.Microsecond)
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", status).
			Int("size", size).
			Uint64("duration_us", dur).
			Msg("request")
		s.stats.Histogram("req_duration_us", dur)
	}))
	chain = chain.Append(hlog.RemoteAddrHandler("ip"))
	chain = chain.Append(hlog.RequestIDHandler("req_id", "Request-Id"))

	router := http.NewServeMux()
	router.Handle(`/`, chain.Then(s.index()))
	router.Handle(`/robots.txt`, chain.Then(s.robots()))
	router.Handle(`/package/`, chain.Then(s.agentPackage()))
	if viper.GetBool(config.KeyLocalPackages) {
		router.Handle(`/packages/`, chain.Then(http.StripPrefix(`/packages/`, http.FileServer(http.Dir(viper.GetString(config.KeyLocalPackagePath))))))
	}
	router.Handle(`/template/`, chain.Then(s.template()))
	router.Handle(`/broker/`, chain.Then(s.broker()))
	router.Handle(`/install/conf/`, chain.Then(s.config())) // TODO: deprecate, in favor of /install/config/
	router.Handle(`/install/config/`, chain.Then(s.config()))
	if viper.GetString(config.KeyRPMFile) != "" {
		router.Handle(`/install/rpm/`, chain.Then(s.rpm()))
	}
	router.Handle(`/install/`, chain.Then(s.install()))
	router.Handle(`/utils/`, chain.Then(s.tool())) // TODO: deprecate, in favor of /tool/
	router.Handle(`/tool/`, chain.Then(s.tool()))

	// HTTP listener (1-n)
	{
		serverList := viper.GetStringSlice(config.KeyListen)
		if len(serverList) == 0 {
			serverList = []string{defaults.Listen}
		}
		for idx, addr := range serverList {
			ta, err := parseListen(addr)
			if err != nil {
				s.logger.Error().Err(err).Int("id", idx).Str("addr", addr).Msg("resolving address")
				return nil, errors.Wrap(err, "HTTP Server")
			}

			svr := httpServer{
				address: ta,
				server: &http.Server{
					Addr:    ta.String(),
					Handler: router,
				},
			}
			svr.server.SetKeepAlivesEnabled(false)

			s.svrHTTP = append(s.svrHTTP, &svr)
		}
	}

	// HTTPS listener (singular)
	if addr := viper.GetString(config.KeySSLListen); addr != "" {
		ta, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			s.logger.Error().Err(err).Str("addr", addr).Msg("resolving address")
			return nil, errors.Wrap(err, "SSL Server")
		}

		certFile := viper.GetString(config.KeySSLCertFile)
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			s.logger.Error().Err(err).Str("cert_file", certFile).Msg("SSL server")
			return nil, errors.Wrapf(err, "SSL server cert file")
		}

		keyFile := viper.GetString(config.KeySSLKeyFile)
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			s.logger.Error().Err(err).Str("key_file", keyFile).Msg("SSL server")
			return nil, errors.Wrapf(err, "SSL server key file")
		}

		svr := sslServer{
			address:  ta,
			certFile: certFile,
			keyFile:  keyFile,
			server: &http.Server{
				Addr:    ta.String(),
				Handler: router,
			},
		}

		svr.server.SetKeepAlivesEnabled(false)
		s.svrHTTPS = &svr
	}

	return &s, nil
}

// Start main listening server(s)
func (s *Server) Start() error {
	if len(s.svrHTTP) == 0 && s.svrHTTPS == nil {
		return errors.New("No servers defined")
	}

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go s.startHTTPS(ctx, &wg)

	for _, svrHTTP := range s.svrHTTP {
		wg.Add(1)
		go s.startHTTP(ctx, svrHTTP, &wg)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	s.logger.Info().Msg("interrupt, shutting down")

	s.logger.Info().Msg("telling children to stop")
	cancel()

	wg.Wait()

	s.stats.Close()

	return nil
}

func (s *Server) startHTTP(ctx context.Context, svr *httpServer, wg *sync.WaitGroup) error {
	defer wg.Done()

	if svr == nil {
		s.logger.Debug().Msg("No listen configured, skipping server")
		return nil
	}
	if svr.address == nil || svr.server == nil {
		s.logger.Debug().Msg("listen not configured, skipping server")
		return nil
	}

	go func() {
		s.logger.Info().Str("listen", svr.address.String()).Msg("Starting")
		if err := svr.server.ListenAndServe(); err != nil {
			s.logger.Fatal().Err(err).Msg("HTTP Server")
		}
	}()

	<-ctx.Done()
	s.logger.Info().Msg("caller requested cancellation")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svr.server.Shutdown(shutdownCtx)

	s.logger.Info().Msg("http server gracefully stopped")

	return nil
}

func (s *Server) startHTTPS(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	if s.svrHTTPS == nil {
		s.logger.Debug().Msg("No SSL listen configured, skipping server")
		return nil
	}

	go func() {
		s.logger.Info().Str("listen", s.svrHTTPS.server.Addr).Msg("SSL starting")
		if err := s.svrHTTPS.server.ListenAndServeTLS(s.svrHTTPS.certFile, s.svrHTTPS.keyFile); err != nil {
			s.logger.Fatal().Err(err).Msg("SSL Server")
		}
	}()

	<-ctx.Done()
	s.logger.Info().Msg("caller requested cancellation")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.svrHTTPS.server.Shutdown(shutdownCtx)

	s.logger.Info().Msg("https server gracefully stopped")

	return nil
}

// parseListen parses and fixes listen spec
func parseListen(spec string) (*net.TCPAddr, error) {
	// empty, default
	if spec == "" {
		spec = defaults.Listen
	}
	// only a port, prefix with colon
	if ok, _ := regexp.MatchString(`^[0-9]+$`, spec); ok {
		spec = ":" + spec
	}
	// ipv4 w/o port, add default
	if strings.Contains(spec, ".") && !strings.Contains(spec, ":") {
		spec += defaults.Listen
	}
	// ipv6 w/o port, add default
	if ok, _ := regexp.MatchString(`^\[[a-f0-9:]+\]$`, spec); ok {
		spec += defaults.Listen
	}

	host, port, err := net.SplitHostPort(spec)
	if err != nil {
		return nil, errors.Wrap(err, "parsing listen")
	}

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, errors.Wrap(err, "resolving listen")
	}

	return addr, nil
}
