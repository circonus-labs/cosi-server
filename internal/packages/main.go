// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package packages

import (
	"fmt"
	"strings"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// New creates new instance of packages
func New(file string) (*Packages, error) {
	if file == "" {
		file = viper.GetString(config.KeyPackageConfigFile)
		if file == "" {
			return nil, errors.Errorf("package configuration file not set")
		}
	}

	var c packageConfig
	if err := config.LoadConfigFile(file, &c); err != nil {
		return nil, errors.Wrap(err, "loading package configuration")
	}

	list := packageList{}
	supported := []string{}

	for _, item := range c {
		if item.PackageInfo.File == "" && item.PackageInfo.Name == "" {
			log.Warn().
				Str("pkg", "packages").
				Str("type", item.OSType).
				Str("dist", item.Distro).
				Str("vers", item.Version).
				Str("arch", item.Arch).
				Msg("invalid package information - no file/name provided, ignoring")
			continue
		}

		ostype := strings.ToLower(item.OSType)
		dist := strings.ToLower(item.Distro)

		if _, ok := list[ostype]; !ok {
			list[ostype] = map[string]map[string]map[string]PackageInfo{}
		}
		if _, ok := list[ostype][dist]; !ok {
			list[ostype][dist] = map[string]map[string]PackageInfo{}
		}
		if _, ok := list[ostype][dist][item.Version]; !ok {
			list[ostype][dist][item.Version] = map[string]PackageInfo{}
		}
		list[ostype][dist][item.Version][item.Arch] = item.PackageInfo
		supported = append(supported, fmt.Sprintf("%s %s %s", item.Distro, item.Version, item.Arch))
		log.Debug().
			Str("pkg", "packages").
			Str("type", item.OSType).
			Str("dist", item.Distro).
			Str("vers", item.Version).
			Str("arch", item.Arch).
			Msg("added")
	}

	if len(list) == 0 {
		return nil, errors.New("no valid packages found")
	}

	return &Packages{supported, list}, nil
}

// ListSupported returns list of supported distro, version, architecture combinations
func (p *Packages) ListSupported() []string {
	return p.supported
}

// GetPackageInfo returns package information for type, distro, version, architecture
// combination or error indicating what is not supported
func (p *Packages) GetPackageInfo(ostype, distro, version, arch string) (*PackageInfo, error) {
	if ostype == "" {
		return nil, errors.Errorf("invalid OS type (blank)")
	}
	if distro == "" {
		return nil, errors.Errorf("invalid OS distro (blank)")
	}
	if version == "" {
		return nil, errors.Errorf("invalid OS distro version (blank)")
	}
	if arch == "" {
		return nil, errors.Errorf("invalid system arch (blank)")
	}

	ost := strings.ToLower(ostype)
	dist := strings.ToLower(distro)

	if _, ok := p.packageList[ost]; !ok {
		return nil, errors.Errorf("unsupported OS Type (%s)", ostype)
	}
	if _, ok := p.packageList[ost][dist]; !ok {
		return nil, errors.Errorf("unsupported OS Distro (%s)", distro)
	}
	if _, ok := p.packageList[ost][dist][version]; !ok {
		return nil, errors.Errorf("unsupported %s version (v%s)", distro, version)
	}

	pi, ok := p.packageList[ost][dist][version][arch]
	if !ok {
		return nil, errors.Errorf("unsupported architecture (%s) for %s %s", arch, distro, version)
	}

	if pi.File != "" {
		if pi.URL == "" {
			pi.URL = viper.GetString(config.KeyPackageBaseURL)
		}
		pi.URL = strings.Replace(pi.URL, pi.File, "", -1)
		pathSep := "/"
		if !strings.HasSuffix(pi.URL, pathSep) {
			pi.URL += pathSep
		}
	}
	return &pi, nil
}
