// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package templates

import (
	"testing"

	"github.com/circonus-labs/cosi-server/api"
	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/circonus-labs/cosi-server/internal/config/defaults"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func TestNew(t *testing.T) {
	t.Log("Testing New")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	// no settings or parameter
	{
		viper.Reset()
		_, err := New()
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// missing
	{
		viper.Set(config.KeyContentPath, "testdata/missing")
		_, err := New()
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// not dir, passed
	{
		viper.Set(config.KeyContentPath, "testdata/not_dir")
		_, err := New()
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// valid
	{
		viper.Set(config.KeyContentPath, "testdata/")
		_, err := New()
		if err != nil {
			t.Fatalf("expected NO error, got %v", err)
		}
	}
}

func TestGet(t *testing.T) {
	t.Log("Testing Get")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	viper.Set(config.KeyTemplateTypeRx, defaults.TemplateTypeRx)
	viper.Set(config.KeyTemplateNameRx, defaults.TemplateNameRx)
	viper.Set(config.KeyEnableTemplateCache, defaults.EnableTemplateCache)

	viper.Set(config.KeyContentPath, "testdata/")
	tmpl, err := New()
	if err != nil {
		t.Fatalf("expected NO error, got %v", err)
	}

	// missing category and name
	{
		_, err := tmpl.Get("", "", "", "", "", "")
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// invalid category
	{
		_, err := tmpl.Get("", "", "", "", "#graph", "")
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// missing name
	{
		_, err := tmpl.Get("", "", "", "", "graph", "")
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// invalid name
	{
		_, err := tmpl.Get("", "", "", "", "graph", "#test")
		if err == nil {
			t.Fatal("expected error")
		}
	}

	templateType := "graph"
	tt := []struct {
		key         string
		expectError bool
	}{
		{"default", false},
		{"ostype", false},
		{"osdistro", false},
		{"osvers", false},
		{"sysarch", false},
		{"missing", true},
		{"empty", true},
		{"cached", false}, // get it into cache
		{"cached", false}, // retrieve from cache
	}

	for _, tst := range tt {
		t.Logf("\t%s", tst.key)
		td, err := tmpl.Get("linux", "ubuntu", "16.04", "x86_64", templateType, tst.key)
		if tst.expectError {
			if err == nil {
				t.Fatal("expected error")
			}
			continue
		}
		if err != nil {
			t.Fatalf("expected NO error, got (%v)", err)
		}

		v := api.Template{}
		if err := toml.Unmarshal(*td, &v); err != nil {
			t.Fatalf("error parsing template (%s)", err)
			if v.Type != templateType || v.Name != tst.key {
				t.Fatalf("expected template ID (%s-%s) got (%s-%s)", templateType, v.Name, v.Type, tst.key)
			}
		}
	}
}
