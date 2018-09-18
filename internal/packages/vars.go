// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package packages

// Packages defines the list of supported distro, version, architecture combinations
// and the agent package details
type Packages struct {
	supported   []string
	packageList packageList
}

// PackageInfo defines the package to use for the specific distro, version, architecture combination
type PackageInfo struct {
	URL     string `json:"package_url,omitempty" yaml:"package_url" toml:"package_url"`
	File    string `json:"package_file,omitempty" yaml:"package_file" toml:"package_file"`
	PubURL  string `json:"publisher_url,omitempty" yaml:"publisher_url" toml:"publisher_url"`
	PubName string `json:"publisher_name,omitempty" yaml:"publisher_name" toml:"publisher_name"`
	Name    string `json:"package_name,omitempty" yaml:"package_name" toml:"package_name"`
}

type osDetail struct {
	Distro      string      `json:"dist" yaml:"dist" toml:"dist"`
	Version     string      `json:"vers" yaml:"vers" toml:"vers"`
	OSType      string      `json:"type" yaml:"type" toml:"type"`
	Arch        string      `json:"arch" yaml:"arch" toml:"arch"`
	PackageInfo PackageInfo `json:"package_info" yaml:"package_info" toml:"package_info"`
}

type packageConfig []osDetail
type packageList map[string]map[string]map[string]map[string]PackageInfo
