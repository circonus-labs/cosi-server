// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"net/url"

	"github.com/pkg/errors"
)

// Config defines the cosi-server api client configuration
type Config struct {
	OSType    string
	OSDistro  string
	OSVersion string
	SysArch   string
	CosiURL   string
}

// Client defines a cosi-server api client
type Client struct {
	cosiURL   *url.URL
	osType    string
	osDistro  string
	osVersion string
	sysArch   string
}

// ServerInfo defines information about the cosi-server. description, version, and
// list of supported operating systems
type ServerInfo struct {
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Supported   []string `json:"supported"`
}

// GraphFilters defines the include/exclude filters for variable graphs (e.g. disk, fs, network interfaces)
type GraphFilters struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

// Package defines the agent package to use for a specific operating system
type Package struct {
	File          string `json:"package_file,omitempty"`
	URL           string `json:"package_url,omitempty"`
	Name          string `json:"package_name,omitempty"`
	PublisherURL  string `json:"publisher_url,omitempty"`
	PublisherName string `json:"publisher_name,omitempty"`
}

// TOML template structs
/* use:
   include "github.com/pelletier/go-toml"

   v := Template{}
   if err := toml.Unmarshal(data, &v); err != nil {
       log.Fatal(err)
   }
*/

// Template defines a TOML template received from the COSI API
type Template struct {
	Type        string                    `toml:"type"`        // common, required
	Name        string                    `toml:"name"`        // common, required
	Version     string                    `toml:"version"`     // common, required
	Description string                    `toml:"description"` // common
	Configs     map[string]TemplateConfig `toml:"configs"`     // common, required
	Variable    bool                      `toml:"variable"`    // graph only
	Filter      TemplateFilter            `toml:"filters"`     // graph only
}

// TemplateFilter defines the include and exclude regex lists to use
// for 'variable' graphs and datapoints
type TemplateFilter struct {
	Include []string `toml:"include"`
	Exclude []string `toml:"exclude"`
}

// TemplateConfig defines a specific configuration template instance
type TemplateConfig struct {
	Datapoints []TemplateDatapoint `toml:"datapoints"` // graph only
	Template   string              `toml:"template"`   // common, required
	Variable   bool                `toml:"variable"`   // graph only
	Widgets    []TemplateWidget    `toml:"widgets"`    // dashboard only
}

// TemplateDatapoint defines a graph datapoint template
type TemplateDatapoint struct {
	Variable bool           `toml:"variable"`
	Filter   TemplateFilter `toml:"filter"`       // variable datapoint only
	MetricRx string         `toml:"metric_regex"` // variable datapoint only
	Template string         `toml:"template"`     // required
}

// TemplateWidget defines a dashboard widget template
type TemplateWidget struct {
	GraphName string `toml:"graph_name"` // graph widget only
	Template  string `toml:"template"`   // required
}

const (
	// TemplateFileExtension defines the format and file extension for templates
	TemplateFileExtension = ".toml"
)

// New creates a new cosi-server api client
func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("invalid config (nil)")
	}

	if cfg.OSType == "" {
		return nil, errors.New("invalid OSType (empty)")
	}

	if cfg.OSDistro == "" {
		return nil, errors.New("invalid OSDistro (empty)")
	}

	if cfg.OSVersion == "" {
		return nil, errors.New("invalid OSVersion (empty)")
	}

	if cfg.SysArch == "" {
		return nil, errors.New("invalid SysArch (empty)")
	}

	if cfg.CosiURL == "" {
		cfg.CosiURL = "https://onestep.circonus.com/"
	}

	u, err := url.Parse(cfg.CosiURL)
	if err != nil {
		return nil, errors.Wrap(err, "invalid CosiURL")
	}

	c := Client{
		cosiURL:   u,
		osType:    cfg.OSType,
		osDistro:  cfg.OSDistro,
		osVersion: cfg.OSVersion,
		sysArch:   cfg.SysArch,
	}

	return &c, nil
}

func (c *Client) genQueryString(args *map[string]string, setOSArgs bool) string {
	q := url.Values{}
	if setOSArgs {
		q.Set("type", url.QueryEscape(c.osType))
		q.Set("dist", url.QueryEscape(c.osDistro))
		q.Set("vers", url.QueryEscape(c.osVersion))
		q.Set("arch", url.QueryEscape(c.sysArch))
	}
	if args != nil {
		for k, v := range *args {
			q.Set(k, url.QueryEscape(v))
		}
	}
	return q.Encode()
}
