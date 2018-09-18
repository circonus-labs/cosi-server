// Copyright © 2018 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package cmd

import (
	"fmt"
	stdlog "log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/circonus-labs/cosi-server/internal/config/defaults"
	"github.com/circonus-labs/cosi-server/internal/release"
	"github.com/circonus-labs/cosi-server/internal/server"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cosi-server",
	Short: "Circonus One Step Install Server",
	Long: `The COSI server exposes the endpoint the cosi install
script communicates with for OS support checks, agent installation
information, and visualization templates.`,
	PersistentPreRunE: initLogging,
	Run: func(cmd *cobra.Command, args []string) {
		//
		// show version and exit
		//
		if viper.GetBool(config.KeyShowVersion) {
			fmt.Printf("%s v%s - commit: %s, date: %s, tag: %s\n", release.NAME, release.VERSION, release.COMMIT, release.DATE, release.TAG)
			return
		}

		//
		// show configuration and exit
		//
		if viper.GetString(config.KeyShowConfig) != "" {
			if err := config.ShowConfig(os.Stdout); err != nil {
				log.Fatal().Err(err).Msg("show-config")
			}
			return
		}

		log.Info().
			Int("pid", os.Getpid()).
			Str("name", release.NAME).
			Str("ver", release.VERSION).Msg("Starting")

		s, err := server.New()
		if err != nil {
			log.Fatal().Err(err).Msg("initializing")
		}

		// check for rpm file
		if viper.GetString(config.KeyRPMFile) != "" {
			rpmFile := viper.GetString(config.KeyRPMFile)
			if _, err := os.Stat(rpmFile); os.IsNotExist(err) {
				log.Fatal().Err(err).Str("rpm_file", rpmFile).Msg("RPM file not found")
			}
		} else {
			// check default if rpm file not specified in configuration
			rpmFile := filepath.Join(viper.GetString(config.KeyContentPath), "files", "cosi-"+release.VERSION+"-1.noarch.rpm")
			if _, err := os.Stat(rpmFile); err == nil {
				viper.Set(config.KeyRPMFile, rpmFile)
			}
		}

		// add config to expvar stats and start server

		config.StatConfig()

		if err := s.Start(); err != nil {
			log.Fatal().Err(err).Msg("starting server")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zlog := zerolog.New(zerolog.SyncWriter(os.Stderr)).With().Timestamp().Logger()
	log.Logger = zlog

	stdlog.SetFlags(0)
	stdlog.SetOutput(zlog)

	cobra.OnInitialize(initConfig)

	desc := func(desc, env string) string {
		return fmt.Sprintf("[ENV: %s] %s", env, desc)
	}

	//
	// Basic
	//
	{
		var (
			longOpt     = "config"
			shortOpt    = "c"
			description = "config file (default is " + defaults.EtcPath + "/" + release.NAME + ".(json|toml|yaml)"
		)
		RootCmd.PersistentFlags().StringVarP(&cfgFile, longOpt, shortOpt, "", description)
	}

	{
		var (
			key         = config.KeyListen
			longOpt     = "listen"
			shortOpt    = "l"
			envVar      = release.ENVPREFIX + "_LISTEN"
			description = "Listen spec e.g. :80, [::1], [::1]:80, 127.0.0.1, 127.0.0.1:80, foo.bar.baz, foo.bar.baz:80 " + `(default "` + defaults.Listen + `")`
		)

		RootCmd.Flags().StringSliceP(longOpt, shortOpt, []string{}, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
	}

	{
		const (
			key         = config.KeyContentPath
			longOpt     = "content-dir"
			envVar      = release.ENVPREFIX + "_CONTENT_DIR"
			description = "Content directory"
		)

		RootCmd.Flags().String(longOpt, defaults.ContentPath, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.ContentPath)
	}

	{
		const (
			key         = config.KeyPackageConfigFile
			longOpt     = "package-conf"
			envVar      = release.ENVPREFIX + "_PACKAGE_CONF"
			description = "Package configuration file"
		)

		RootCmd.Flags().String(longOpt, defaults.PackageConfigFile, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.PackageConfigFile)
	}

	{
		const (
			key         = config.KeyPackageBaseURL
			longOpt     = "package-url"
			envVar      = release.ENVPREFIX + "_PACKAGE_URL"
			description = "Package base URL"
		)

		RootCmd.Flags().String(longOpt, defaults.BasePackageURL, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.BasePackageURL)
	}

	//
	// SSL
	//
	{
		const (
			key          = config.KeySSLListen
			longOpt      = "ssl-listen"
			defaultValue = ""
			envVar       = release.ENVPREFIX + "_SSL_LISTEN"
			description  = "SSL listen address and port [IP]:[PORT] - setting enables SSL"
		)

		RootCmd.Flags().String(longOpt, defaultValue, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
	}

	{
		const (
			key         = config.KeySSLCertFile
			longOpt     = "ssl-cert-file"
			envVar      = release.ENVPREFIX + "_SSL_CERT_FILE"
			description = "SSL Certificate file (PEM cert and CAs concatenated together)"
		)

		RootCmd.Flags().String(longOpt, defaults.SSLCertFile, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.SSLCertFile)
	}

	{
		const (
			key         = config.KeySSLKeyFile
			longOpt     = "ssl-key-file"
			envVar      = release.ENVPREFIX + "_SSL_KEY_FILE"
			description = "SSL Key file"
		)

		RootCmd.Flags().String(longOpt, defaults.SSLKeyFile, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.SSLKeyFile)
	}

	{
		const (
			key         = config.KeySSLVerify
			longOpt     = "ssl-verify"
			envVar      = release.ENVPREFIX + "_SSL_VERIFY"
			description = "Enable SSL verification"
		)

		RootCmd.Flags().Bool(longOpt, defaults.SSLVerify, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.SSLVerify)
	}

	//
	// Template caching
	//
	{
		const (
			key         = config.KeyEnableTemplateCache
			longOpt     = "cache-templates"
			envVar      = release.ENVPREFIX + "_CACHE_TEMPLATES"
			description = "Enable template caching"
		)

		RootCmd.Flags().Bool(longOpt, defaults.EnableTemplateCache, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.EnableTemplateCache)
	}

	//
	// Validation regular expression defaults (config file only)
	//
	// OS Type
	viper.SetDefault(config.KeyParamTypeRx, defaults.ParamTypeRx)
	// OS Distro
	viper.SetDefault(config.KeyParamDistroRx, defaults.ParamDistroRx)
	viper.SetDefault(config.KeyIsRHELDistroRx, defaults.IsRHELDistroRx)
	viper.SetDefault(config.KeyIsSolarisDistroRx, defaults.IsSolarisDistroRx)
	// OS Version
	viper.SetDefault(config.KeyParamVersionRx, defaults.ParamVersionRx)
	viper.SetDefault(config.KeyParamVersionCleanerRx, defaults.ParamVersionCleanerRx)
	// System architecture
	viper.SetDefault(config.KeyParamArchRx, defaults.ParamArchRx)
	// Agent and broker modes
	viper.SetDefault(config.KeyParamAgentModeRx, defaults.ParamAgentModeRx)
	viper.SetDefault(config.KeyAgentPullModeRx, defaults.AgentPullModeRx)
	viper.SetDefault(config.KeyAgentPushModeRx, defaults.AgentPushModeRx)
	// Templates
	viper.SetDefault(config.KeyTemplateTypeRx, defaults.TemplateTypeRx)
	viper.SetDefault(config.KeyTemplateNameRx, defaults.TemplateNameRx)

	//
	// Brokers - default brokers to use for the different agent modes
	//
	viper.SetDefault(config.KeyBrokerFallbackList, defaults.BrokerFallbackList)
	viper.SetDefault(config.KeyBrokerFallbackDefault, defaults.BrokerFallbackDefault)
	viper.SetDefault(config.KeyBrokerPushList, defaults.BrokerPushList)
	viper.SetDefault(config.KeyBrokerPushDefault, defaults.BrokerPushDefault)
	viper.SetDefault(config.KeyBrokerPullList, defaults.BrokerPullList)
	viper.SetDefault(config.KeyBrokerPullDefault, defaults.BrokerPullDefault)

	// RPM installer file name
	viper.SetDefault(config.KeyRPMFile, defaults.RPMFile)

	// Local packages
	viper.SetDefault(config.KeyLocalPackages, defaults.LocalPackages)
	viper.SetDefault(config.KeyLocalPackagePath, defaults.LocalPackagePath)

	//
	// StatsD client
	//
	{
		const (
			key         = config.KeyStatsdAddress
			longOpt     = "statsd-address"
			envVar      = release.ENVPREFIX + "_STATSD_ADDRESS"
			description = "StatsD server address"
		)

		RootCmd.Flags().String(longOpt, defaults.StatsdAddress, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.StatsdAddress)
	}
	{
		const (
			key         = config.KeyStatsdInterval
			longOpt     = "statsd-interval"
			envVar      = release.ENVPREFIX + "_STATSD_INTERVAL"
			description = "StatsD meric flushing interval"
		)

		RootCmd.Flags().Duration(longOpt, defaults.StatsdInterval, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.StatsdInterval)
	}
	{
		const (
			key         = config.KeyStatsdPrefix
			longOpt     = "statsd-prefix"
			envVar      = release.ENVPREFIX + "_STATSD_PREFIX"
			description = "StatsD metric name prefix"
		)

		RootCmd.Flags().String(longOpt, defaults.StatsdPrefix, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.StatsdPrefix)
	}

	//
	// Miscellenous
	//
	{
		const (
			key         = config.KeyDebug
			longOpt     = "debug"
			shortOpt    = "d"
			envVar      = release.ENVPREFIX + "_DEBUG"
			description = "Enable debug messages"
		)

		RootCmd.Flags().BoolP(longOpt, shortOpt, defaults.Debug, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.Debug)
	}

	{
		const (
			key         = config.KeyLogLevel
			longOpt     = "log-level"
			envVar      = release.ENVPREFIX + "_LOG_LEVEL"
			description = "Log level [(panic|fatal|error|warn|info|debug|disabled)]"
		)

		RootCmd.Flags().String(longOpt, defaults.LogLevel, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.LogLevel)
	}

	{
		const (
			key         = config.KeyLogPretty
			longOpt     = "log-pretty"
			envVar      = release.ENVPREFIX + "_LOG_PRETTY"
			description = "Output formatted/colored log lines [ignored on windows]"
		)

		RootCmd.Flags().Bool(longOpt, defaults.LogPretty, desc(description, envVar))
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
		viper.BindEnv(key, envVar)
		viper.SetDefault(key, defaults.LogPretty)
	}

	{
		const (
			key          = config.KeyShowVersion
			longOpt      = "version"
			shortOpt     = "V"
			defaultValue = false
			description  = "Show version and exit"
		)
		RootCmd.Flags().BoolP(longOpt, shortOpt, defaultValue, description)
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
	}

	{
		const (
			key         = config.KeyShowConfig
			longOpt     = "show-config"
			description = "Show config (json|toml|yaml) and exit"
		)

		RootCmd.Flags().String(longOpt, "", description)
		viper.BindPFlag(key, RootCmd.Flags().Lookup(longOpt))
	}
}

// initLogging initializes zerolog
func initLogging(cmd *cobra.Command, args []string) error {
	//
	// Enable formatted output
	//
	if viper.GetBool(config.KeyLogPretty) {
		if runtime.GOOS != "windows" {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
		} else {
			log.Warn().Msg("log-pretty not applicable on this platform")
		}
	}

	//
	// Enable debug logging, if requested
	// otherwise, default to info level and set custom level, if specified
	//
	if viper.GetBool(config.KeyDebug) {
		viper.Set(config.KeyLogLevel, "debug")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("--debug flag, forcing debug log level")
	} else {
		if viper.IsSet(config.KeyLogLevel) {
			level := viper.GetString(config.KeyLogLevel)

			switch level {
			case "panic":
				zerolog.SetGlobalLevel(zerolog.PanicLevel)
			case "fatal":
				zerolog.SetGlobalLevel(zerolog.FatalLevel)
			case "error":
				zerolog.SetGlobalLevel(zerolog.ErrorLevel)
			case "warn":
				zerolog.SetGlobalLevel(zerolog.WarnLevel)
			case "info":
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			case "debug":
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			case "disabled":
				zerolog.SetGlobalLevel(zerolog.Disabled)
			default:
				return errors.Errorf("Unknown log level (%s)", level)
			}

			log.Debug().Str("log-level", level).Msg("Logging level")
		}
	}

	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(defaults.EtcPath)
		viper.AddConfigPath(".")
		viper.SetConfigName(release.NAME)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		f := viper.ConfigFileUsed()
		if f != "" {
			log.Fatal().Err(err).Str("config_file", f).Msg("Unable to load config file")
		}
	}
}
