// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package api

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
)

var (
	templateDir = path.Join("..", "content", "templates")
)

func TestParseTemplateID(t *testing.T) {
	t.Log("Testing parseTemplateID")

	tests := []struct {
		name        string
		id          string
		tTypeExpect string
		tNameExpect string
		shouldError bool
		errorExpect error
	}{
		{"valid", "check-system", "check", "system", false, nil},
		{"valid (parts>2)", "foo-bar-baz", "foo", "bar-baz", false, nil}, // errors.New("invalid id format (foo-bar-baz)")},
		{"invalid (parts<2)", "foo", "", "", true, errors.New("invalid id format (foo)")},
	}

	for _, test := range tests {
		t.Logf("\t%s", test.name)

		tType, tName, err := parseTemplateID(test.id)
		if test.shouldError {
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != test.errorExpect.Error() {
				t.Fatalf("unexpected error (%s)", err)
			}
			continue
		}

		if err != nil {
			t.Fatalf("unexpected error (%s)", err)
		}

		if tType != test.tTypeExpect {
			t.Fatalf("unexpected type (%s)", tType)
		}

		if tName != test.tNameExpect {
			t.Fatalf("unexpected name (%s)", tName)
		}
	}
}

func genTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/template/check/system/":
			data, err := ioutil.ReadFile(path.Join(templateDir, "check-system"+TemplateFileExtension))
			if err != nil {
				t.Fatalf("Fetching (%s) (%s)", r.URL.Path, err)
			}
			_, _ = w.Write(data)
		case "/template/dashboard/system/":
			data, err := ioutil.ReadFile(path.Join(templateDir, "linux", "dashboard-system"+TemplateFileExtension))
			if err != nil {
				t.Fatalf("Fetching (%s) (%s)", r.URL.Path, err)
			}
			_, _ = w.Write(data)
		case "/template/graph/cpu/":
			data, err := ioutil.ReadFile(path.Join(templateDir, "graph-cpu"+TemplateFileExtension))
			if err != nil {
				t.Fatalf("Fetching (%s) (%s)", r.URL.Path, err)
			}
			_, _ = w.Write(data)
		case "/template/worksheet/system/":
			data, err := ioutil.ReadFile(path.Join(templateDir, "worksheet-system"+TemplateFileExtension))
			if err != nil {
				t.Fatalf("Fetching (%s) (%s)", r.URL.Path, err)
			}
			_, _ = w.Write(data)
		case "/template/json/syntax1/":
			_, _ = w.Write([]byte(`{"type":"foo", "id":"bar}`))
		default:
			_, _ = w.Write([]byte("invalid"))
		}
	}))
}

func TestTemplate(t *testing.T) {
	t.Log("Testing Template")

	ts := genTestServer(t)

	cfg := &Config{
		OSType:    "Linux",
		OSDistro:  "CentOS",
		OSVersion: "7.1.1408",
		SysArch:   "x86_64",
		CosiURL:   ts.URL,
	}

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error (%s)", err)
	}

	tests := []struct {
		name        string
		id          string
		shouldError bool
		errorExpect error
	}{
		{"invalid (empty id)", "", true, errors.New("invalid id (empty)")},
		{"invalid (id format)", "foo", true, errors.New("parsing id: invalid id format (foo)")},
		{"valid (check)", "check-system", false, nil},
		{"valid (dashboard)", "dashboard-system", false, nil},
		{"valid (graph)", "graph-cpu", false, nil},
		{"valid (worksheet)", "worksheet-system", false, nil},
	}

	for _, test := range tests {
		t.Logf("\t%s", test.name)

		// NOTE: tests both FetchTempalte and FetchRawTemplate
		_, err := c.FetchTemplate(test.id)
		if test.shouldError {
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != test.errorExpect.Error() {
				t.Fatalf("unexpected error (%s)", err)
			}
			continue
		}

		if err != nil {
			t.Fatalf("unexpected error (%s)", err)
		}
	}

	ts.Close()
}
