// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package packages

import (
	"testing"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func TestNew(t *testing.T) {
	t.Log("Testing New")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	// no settings or parameter
	{
		viper.Reset()
		_, err := New("")
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// missing file, passed
	{
		viper.Reset()
		_, err := New("testdata/missing.file")
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// missing file, set
	{
		viper.Set(config.KeyPackageConfigFile, "testdata/missing.file")
		_, err := New("")
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// invalid syntax
	{
		viper.Set(config.KeyPackageConfigFile, "testdata/invalid_syntax.yaml")
		_, err := New("")
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// invalid entry
	{
		viper.Set(config.KeyPackageConfigFile, "testdata/invalid_entry.yaml")
		_, err := New("")
		if err == nil {
			t.Fatal("expected error")
		}
	}

	// Valid
	{
		viper.Set(config.KeyPackageConfigFile, "testdata/valid.yaml")
		_, err := New("")
		if err != nil {
			t.Fatalf("expected NO error, got %v", err)
		}
	}
}

func TestListSupported(t *testing.T) {
	t.Log("Testing ListSupported")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	{
		viper.Set(config.KeyPackageConfigFile, "testdata/valid.yaml")
		p, err := New("")
		if err != nil {
			t.Fatalf("expected NO error, got %v", err)
		}
		s := p.ListSupported()
		if len(s) == 0 {
			t.Fatal("no packages supported")
		}
	}
}

func TestGetPackageInfo(t *testing.T) {
	t.Log("Testing GetPackageInfo")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	{
		viper.Set(config.KeyPackageConfigFile, "testdata/valid.yaml")
		p, err := New("")
		if err != nil {
			t.Fatalf("expected NO error, got %v", err)
		}
		// missing parameters
		if _, err := p.GetPackageInfo("", "", "", ""); err == nil {
			t.Fatalf("expected error")
		}
		if _, err := p.GetPackageInfo("Linux", "", "", ""); err == nil {
			t.Fatalf("expected error")
		}
		if _, err := p.GetPackageInfo("Linux", "Ubuntu", "", ""); err == nil {
			t.Fatalf("expected error")
		}
		if _, err := p.GetPackageInfo("Linux", "Ubuntu", "16.04", ""); err == nil {
			t.Fatalf("expected error")
		}

		// invalid parameters
		if _, err := p.GetPackageInfo("foo", "Ubuntu", "16.04", "x86_64"); err == nil {
			t.Fatalf("expected error")
		}
		if _, err := p.GetPackageInfo("Linux", "foo", "16.04", "x86_64"); err == nil {
			t.Fatalf("expected error")
		}
		if _, err := p.GetPackageInfo("Linux", "Ubuntu", "foo", "x86_64"); err == nil {
			t.Fatalf("expected error")
		}
		if _, err := p.GetPackageInfo("Linux", "Ubuntu", "16.04", "foo"); err == nil {
			t.Fatalf("expected error")
		}

		// valid
		if _, err := p.GetPackageInfo("Linux", "Ubuntu", "16.04", "x86_64"); err != nil {
			t.Fatalf("expected NO error, got %v", err)
		}
	}
}
