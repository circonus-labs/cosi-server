// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func (c *Client) get(requrl *url.URL, hdrs *map[string]string) ([]byte, error) {
	if requrl == nil {
		return nil, errors.New("invalid request url (nil)")
	}
	if requrl.String() == "" {
		return nil, errors.New("invalid request url (empty)")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", requrl.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "cosi-server preparing request")
	}

	if hdrs != nil {
		for k, v := range *hdrs {
			req.Header.Set(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "cosi-server request")
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading cosi-server response")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%s - %s - %s", resp.Status, requrl.String(), strings.TrimSpace(string(data)))
	}

	return data, nil
}
