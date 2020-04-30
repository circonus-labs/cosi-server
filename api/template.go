// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

// FetchTemplate retrieves the template for the specified ID from the cosi-server API.
// ID type-name -- e.g. graph-vm, dashboard-system, check-system, etc.
func (c *Client) FetchTemplate(id string) (*Template, error) {
	data, err := c.FetchRawTemplate(id)
	if err != nil {
		return nil, err
	}

	// escape any percent signs so that unmarshalling the
	// json will not result in fmt interpolation error messages
	// embedded in the config attribute
	// data = bytes.Replace(data, []byte("%"), []byte("%%"), -1)

	t := Template{}
	if err := toml.Unmarshal(data, &t); err != nil {
		return nil, errors.Wrapf(checkJSONError(data, err), "parsing %s template", id)
	}

	return &t, nil
}

// FetchRawTemplate retrieves a template from the cosi-server API and
// returns the raw data (does not parse the JSON) or an error. This
// call is used by cosi-tool when it intends to store the template
// on disk.
func (c *Client) FetchRawTemplate(id string) ([]byte, error) {
	if id == "" {
		return nil, errors.New("invalid id (empty)")
	}

	tType, tName, err := parseTemplateID(id)
	if err != nil {
		return nil, errors.Wrap(err, "parsing id")
	}

	u, err := c.cosiURL.Parse(fmt.Sprintf("/template/%s/%s/", tType, tName))
	if err != nil {
		return nil, errors.Wrap(err, "setting URL path")
	}
	u.RawQuery = c.genQueryString(nil, true)

	data, err := c.get(u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fetching template")
	}

	return data, nil
}

func parseTemplateID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "-", 2)
	if len(idParts) != 2 {
		return "", "", errors.Errorf("invalid id format (%s)", id)
	}
	return idParts[0], idParts[1], nil
}

func checkJSONError(data []byte, err error) error {
	if jsonError, ok := err.(*json.SyntaxError); ok {
		line, character, lcErr := lineAndCharacter(string(data), int(jsonError.Offset))
		ne := errors.Wrapf(err, "cannot parse JSON schema due to a syntax error at line %d, character %d: %v", line, character, jsonError.Error())
		if lcErr != nil {
			ne = errors.Wrapf(ne, "Couldn't find the line and character position of the error due to error %v\n", lcErr)
		}
		return ne
	}
	if jsonError, ok := err.(*json.UnmarshalTypeError); ok {
		line, character, lcErr := lineAndCharacter(string(data), int(jsonError.Offset))
		ne := errors.Wrapf(err, "JSON type '%v' cannot be converted into the Go '%v' type on struct '%s', field '%v'. See input file line %d, character %d\n", jsonError.Value, jsonError.Type.Name(), jsonError.Struct, jsonError.Field, line, character)
		if lcErr != nil {
			ne = errors.Wrapf(ne, "Couldn't find the line and character position of the error due to error %v\n", lcErr)
		}
		return ne
	}

	return err
}

func lineAndCharacter(input string, offset int) (line int, character int, err error) {
	lf := rune(0x0A)
	if offset > len(input) || offset < 0 {
		return 0, 0, errors.Errorf("couldn't find offset %d within the input", offset)
	}
	// Humans tend to count from 1.
	line = 1
	for i, b := range input {
		if b == lf {
			line++
			character = 0
		}
		character++
		if i == offset {
			break
		}
	}

	return line, character, nil
}
