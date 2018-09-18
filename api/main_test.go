// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	t.Log("Testing New")

	tests := []struct {
		name          string
		ostype        string
		osdist        string
		osvers        string
		sysarch       string
		cosiurl       string
		shouldError   bool
		expectedError error
	}{
		{"valid", "test", "test", "test", "test", "", false, nil},
		{"invalid os type", "", "test", "test", "test", "", true, errors.New("invalid OSType (empty)")},
		{"invalid os distro", "test", "", "test", "test", "", true, errors.New("invalid OSDistro (empty)")},
		{"invalid os version", "test", "test", "", "test", "", true, errors.New("invalid OSVersion (empty)")},
		{"invalid sys arch", "test", "test", "test", "", "", true, errors.New("invalid SysArch (empty)")},
		{"invalid cosi url", "test", "test", "test", "test", "://foo/bar", true, errors.New("invalid CosiURL: parse ://foo/bar: missing protocol scheme")},
	}

	for _, test := range tests {
		t.Logf("\t%s", test.name)
		_, err := New(&Config{
			CosiURL:   test.cosiurl,
			OSType:    test.ostype,
			OSDistro:  test.osdist,
			OSVersion: test.osvers,
			SysArch:   test.sysarch,
		})

		if test.shouldError {
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != test.expectedError.Error() {
				t.Fatalf("unexpected error (%s)", err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("expected no error, got (%s)", err)
		}
	}

	t.Log("invalid config (nil)")
	{
		_, err := New(nil)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "invalid config (nil)" {
			t.Fatalf("unexpected error (%s)", err)
		}
	}

}
