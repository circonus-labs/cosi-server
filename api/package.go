// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// FetchPackage retrieves information about what agent package to use for a
// specific operating system from the cosi-server API
func (c *Client) FetchPackage(format string) (*Package, error) {
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "text" {
		return nil, errors.Errorf("invalid format (%s)", format)
	}
	accept := "application/json"
	if format == "text" {
		accept = "text/plain"
	}

	u, err := c.cosiURL.Parse("/package/")
	if err != nil {
		return nil, errors.Wrap(err, "setting URL path")
	}
	u.RawQuery = c.genQueryString(nil, true)

	headers := map[string]string{"Accept": accept}
	data, err := c.get(u, &headers)
	if err != nil {
		return nil, errors.Wrap(err, "fetching package")
	}

	var p Package
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, errors.Wrap(err, "parsing package")
	}

	return &p, nil
}
