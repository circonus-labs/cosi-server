// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package cmd

import (
	"testing"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	t.Log("Testing initConfig")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	initConfig()
}

func TestInitLogging(t *testing.T) {
	t.Log("Testing initLogging")
	zerolog.SetGlobalLevel(zerolog.Disabled)

	logLevels := []string{
		"panic",
		"fatal",
		"error",
		"warn",
		"info",
		"debug",
		"disabled",
	}

	for _, level := range logLevels {
		t.Logf("level %s", level)
		viper.Set(config.KeyLogLevel, level)
		err := initLogging(nil, []string{})
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}
		viper.Reset()
	}

	t.Log("level invalid")
	{
		viper.Set(config.KeyLogLevel, "invalid")
		expect := "Unknown log level (invalid)"
		err := initLogging(nil, []string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != expect {
			t.Fatalf("expected (%s) got (%s)", expect, err)
		}
		viper.Reset()
	}

	t.Log("debug flag")
	{
		viper.Set(config.KeyDebug, true)
		err := initLogging(nil, []string{})
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}
		viper.Reset()
	}
}
