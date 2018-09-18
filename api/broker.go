// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// FetchBroker will use the COSI server logic to return the broker
// ID of the default SaaS broker for the given check  type.
func (c *Client) FetchBroker(checkType string) (string, error) {
	if checkType == "" {
		return "", errors.New("invalid check type (empty)")
	}

	u, err := c.cosiURL.Parse("/broker/")
	if err != nil {
		return "", errors.Wrap(err, "setting URL path")
	}
	u.RawQuery = c.genQueryString(&map[string]string{"agent_mode": checkType}, false)

	data, err := c.get(u, nil)
	if err != nil {
		return "", errors.Wrap(err, "fetching broker")
	}

	type brokerInfo struct {
		BrokerID string `json:"broker_id"`
	}
	var bi brokerInfo
	if err := json.Unmarshal(data, &bi); err != nil {
		return "", errors.Wrap(err, "parsing broker info")
	}

	return bi.BrokerID, nil
}
