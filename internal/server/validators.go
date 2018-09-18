// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/viper"
)

func (s *Server) compileValidators() error {
	{
		rx, err := regexp.Compile(viper.GetString(config.KeyParamTypeRx))
		if err != nil {
			return errors.Wrap(err, "os type regex")
		}
		s.typerx = rx
	}
	// os distro
	{
		rx, err := regexp.Compile(viper.GetString(config.KeyParamDistroRx))
		if err != nil {
			return errors.Wrap(err, "os distro regex")
		}
		s.distrx = rx
	}
	// os version
	{
		// match
		{
			rx, err := regexp.Compile(viper.GetString(config.KeyParamVersionRx))
			if err != nil {
				return errors.Wrap(err, "os version regex")
			}
			s.versrx = rx
		}
		// clean
		{
			rx, err := regexp.Compile(viper.GetString(config.KeyParamVersionCleanerRx))
			if err != nil {
				return errors.Wrap(err, "os version cleaner regex")
			}
			s.vercleanrx = rx
		}
	}
	// os arch
	{
		rx, err := regexp.Compile(viper.GetString(config.KeyParamArchRx))
		if err != nil {
			return errors.Wrap(err, "system architecture regex")
		}
		s.archrx = rx
	}
	// RHEL distro (if matched, only use version major as version)
	{
		rx, err := regexp.Compile(viper.GetString(config.KeyIsRHELDistroRx))
		if err != nil {
			return errors.Wrap(err, "rhel regex")
		}
		s.rhelrx = rx
	}
	// Solaris type distro (which use 'pkg' not rpm or deb)
	{
		rx, err := regexp.Compile(viper.GetString(config.KeyIsSolarisDistroRx))
		if err != nil {
			return errors.Wrap(err, "solaris regex")
		}
		s.solarisrx = rx
	}

	// Agent Modes
	{
		rx, err := regexp.Compile(viper.GetString(config.KeyAgentPullModeRx))
		if err != nil {
			return errors.Wrap(err, "agent pull mode regex")
		}
		s.modepullrx = rx
	}
	{
		rx, err := regexp.Compile(viper.GetString(config.KeyAgentPushModeRx))
		if err != nil {
			return errors.Wrap(err, "agent push mode regex")
		}
		s.modepushrx = rx
	}

	return nil
}

func (s *Server) validateRequiredParams(r *http.Request) (*params, error) {
	p := r.URL.Query()
	pinfo := params{}

	// OS type
	{
		paramErr := errors.New("invalid system 'type' specified")
		osType := strings.ToLower(p.Get("type"))
		if osType == "" {
			return nil, paramErr
		}
		if !s.typerx.MatchString(osType) {
			hlog.FromRequest(r).Error().Str("type_param", osType).Str("type_regex", s.typerx.String()).Msg("OS Type not matched")
			return nil, paramErr
		}
		pinfo.osType = osType
	}

	// OS Distribution
	{
		paramErr := errors.New("invalid system 'dist' specified")
		osDistro := strings.ToLower(p.Get("dist"))
		if osDistro == "" {
			return nil, paramErr
		}
		if !s.distrx.MatchString(osDistro) {
			hlog.FromRequest(r).Error().Str("dist_param", osDistro).Str("dist_regex", s.distrx.String()).Msg("OS Distro not matched")
			return nil, paramErr
		}
		pinfo.osDistro = osDistro
	}

	// OS Version
	{
		paramErr := errors.New("invalid system 'vers' specified")
		osVers := p.Get("vers")
		if osVers == "" {
			return nil, paramErr
		}
		if !s.versrx.MatchString(osVers) {
			hlog.FromRequest(r).Error().Str("vers_param", osVers).Str("vers_regex", s.versrx.String()).Msg("OS Version not matched")
			return nil, paramErr
		}
		clean := s.vercleanrx.ReplaceAllString(osVers, "")
		if clean == "" {
			return nil, paramErr
		}
		pinfo.osVers = clean
		if s.rhelrx.MatchString(pinfo.osDistro) {
			v := strings.Split(clean, ".")
			if len(v) == 0 {
				return nil, paramErr
			}
			// only use the major version for RHEL distros
			pinfo.osVers = v[0]
		}
	}

	// System architecture
	{
		paramErr := errors.New("invalid system 'arch' specified")
		sysArch := p.Get("arch")
		if sysArch == "" {
			return nil, paramErr
		}
		if !s.archrx.MatchString(sysArch) {
			hlog.FromRequest(r).Error().Str("arch_param", sysArch).Str("arch_regex", s.archrx.String()).Msg("System Architecture not matched")
			return nil, paramErr
		}
		pinfo.sysArch = sysArch
	}

	return &pinfo, nil
}

func (s *Server) validateTemplateSpec(r *http.Request) (*templateSpec, error) {
	spec := r.URL.Path
	tinfo := templateSpec{}

	// expecting a string such as "/template/type/name/"
	specItems := strings.Split(spec, "/")
	// should result in: specItems []string{"", "template", "type", "name", ""}
	if len(specItems) != 5 {
		hlog.FromRequest(r).Error().Str("spec", spec).Msg("invalid template spec")
		return nil, errors.New("invalid template specification")
	}

	tinfo.Type = specItems[2]
	tinfo.Name = specItems[3]

	if !s.templates.Typerx.MatchString(tinfo.Type) {
		hlog.FromRequest(r).Error().Str("type_param", tinfo.Type).Str("type_regex", s.templates.Typerx.String()).Msg("Template type not matched")
		return nil, errors.New("invalid template type")
	}

	if !s.templates.Namerx.MatchString(tinfo.Name) {
		hlog.FromRequest(r).Error().Str("name_param", tinfo.Name).Str("name_regex", s.templates.Namerx.String()).Msg("Template name not matched")
		return nil, errors.New("invalid template name")
	}

	return &tinfo, nil
}
