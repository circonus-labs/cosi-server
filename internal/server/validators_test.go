// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/circonus-labs/cosi-server/internal/config/defaults"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func TestValidateRequiredParams(t *testing.T) {
	t.Log("Testing validateRequiredParams")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	viper.Set(config.KeyParamTypeRx, defaults.ParamTypeRx)
	viper.Set(config.KeyParamDistroRx, defaults.ParamDistroRx)
	viper.Set(config.KeyParamVersionRx, defaults.ParamVersionRx)
	viper.Set(config.KeyParamVersionCleanerRx, defaults.ParamVersionCleanerRx)
	viper.Set(config.KeyParamArchRx, defaults.ParamArchRx)
	viper.Set(config.KeyIsRHELDistroRx, defaults.IsRHELDistroRx)
	viper.Set(config.KeyPackageConfigFile, "../packages/testdata/valid.yaml")
	viper.Set(config.KeyContentPath, "../templates/testdata")
	s, err := New()
	if err != nil {
		t.Fatalf("expected NO error, got %v", err)
	}
	tt := []struct {
		ostype    string
		dist      string
		vers      string
		arch      string
		shouldErr bool
	}{
		{"Linux", "Ubuntu", "16.04", "x86_64", false},
		{"Linux", "Ubuntu", "14.04", "x86_64", false},
		{"Linux", "Ubuntu", "14.04", "i386", false},
		{"Linux", "debian", "9", "x86_64", false},
		{"Linux", "debian", "8", "x86_64", false},
		{"Linux", "debian", "7", "x86_64", false},
		{"Linux", "CentOS", "7.4.1708", "x86_64", false},
		{"Linux", "CentOS", "6.8", "x86_64", false},
		{"Linux", "RedHat", "7.4", "x86_64", false},
		{"Linux", "RedHat", "6.9", "x86_64", false},
		{"Linux", "Oracle", "7.4", "x86_64", false},
		{"Linux", "amzn", "2017.03", "x86_64", false},
		{"Linux", "amzn", "2016.03", "x86_64", false},
		{"Solaris", "OmniOS", "r151014", "x86_64", false},
		{"BSD", "FreeBSD", "11.0", "amd64", false},
		// Add more distros as they are supported
		{"", "Ubuntu", "16.04", "x86_64", true},
		{"foo!", "Ubuntu", "16.04", "x86_64", true},
		{"Linux", "", "16.04", "x86_64", true},
		{"Linux", "foo!", "16.04", "x86_64", true},
		{"Linux", "Ubuntu", "", "x86_64", true},
		{"Linux", "Ubuntu", "foo", "x86_64", true},
		{"Linux", "Ubuntu", "16.04", "", true},
		{"Linux", "Ubuntu", "16.04", "foo", true},
		// add more bad sequences as needed
	}

	for _, tst := range tt {
		t.Logf("\t%v", tst)
		target := fmt.Sprintf("/?type=%s&dist=%s&vers=%s&arch=%s", tst.ostype, tst.dist, tst.vers, tst.arch)
		r := httptest.NewRequest("", target, nil)
		_, err := s.validateRequiredParams(r)
		if tst.shouldErr {
			if err == nil {
				t.Fatal("expected error")
			}
		} else {
			if err != nil {
				t.Fatalf("expected NO error, got %v", err)
			}
		}
	}
}

func TestValidateTemplateSpec(t *testing.T) {
	t.Log("Testing validateTemplateSpec")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	viper.Set(config.KeyPackageConfigFile, "../packages/testdata/valid.yaml")
	viper.Set(config.KeyContentPath, "../templates/testdata")
	s, err := New()
	if err != nil {
		t.Fatalf("expected NO error, got %v", err)
	}
	tt := []struct {
		spec      string
		shouldErr bool
	}{
		{"/templates/check/system/", false},
		{"/templates/graph/vm/", false},
		{"/templates/worksheet/system/", false},
		{"/templates/dashboard/system/", false},
		{"/foo", true},
		{"/foo/bar", true},
		{"/templates", true},
		{"/templates/", true},
		{"/templates/foo", true},
		{"/templates/bar", true},
		{"/templates/bar/", true},
	}

	for _, tst := range tt {
		t.Logf("\t%v", tst)

		r := httptest.NewRequest("", tst.spec, nil)
		_, err := s.validateTemplateSpec(r)
		if tst.shouldErr {
			if err == nil {
				t.Fatal("expected error")
			}
		} else {
			if err != nil {
				t.Fatalf("expected NO error, got %v", err)
			}
		}
	}
}
