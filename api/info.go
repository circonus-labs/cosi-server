// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
)

// FetchInfo retrieves information about the cosi-server. description, version, and
// list of supported operating systems
func (c *Client) FetchInfo() (*ServerInfo, error) {
	u, err := c.cosiURL.Parse("/")
	if err != nil {
		return nil, errors.Wrap(err, "setting URL path")
	}

	data, err := c.get(u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fetching server info")
	}

	var si ServerInfo
	if err := json.Unmarshal(data, &si); err != nil {
		return nil, errors.Wrap(err, "parsing server info")
	}

	// pre sort the supported operating systems for consistency
	sort.Strings(si.Supported)

	return &si, nil
}
