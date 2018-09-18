// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package config

import "time"

// Statsd defines the statsd client options
type Statsd struct {
	Address  string        `json:"address" yaml:"address" toml:"address"`
	Interval time.Duration `json:"interval" yaml:"interval" toml:"interval"`
	Prefix   string        `json:"prefix" yaml:"prefix" toml:"prefix"`
}

// Brokers defines the default broker to use for the different agent modes
type Brokers struct {
	Fallback        []string `json:"fallback" yaml:"fallback" toml:"fallback"`                                                         // fallback is *required*
	FallbackDefault int      `mapstructure:"fallback_default" json:"fallback_default" yaml:"fallback_default" toml:"fallback_default"` // offset into Fallback array or -1 for random
	Push            []string `json:"push" yaml:"push" toml:"push"`                                                                     // e.g. httptrap
	PushDefault     int      `mapstructure:"push_default" json:"push_default" yaml:"push_default" toml:"push_default"`                 // offset into Push array or -1 for random
	Pull            []string `json:"pull" yaml:"pull" toml:"pull"`                                                                     // e.g. reverse
	PullDefault     int      `mapstructure:"pull_default" json:"pull_default" yaml:"pull_default" toml:"pull_default"`                 // offset into Pull array or -1 for random
}

// Log defines the running config.log structure
type Log struct {
	Level  string `json:"level" yaml:"level" toml:"level"`
	Pretty bool   `json:"pretty" yaml:"pretty" toml:"pretty"`
}

// SSL defines the running config.ssl structure
type SSL struct {
	Listen   string `json:"listen" yaml:"listen" toml:"listen"`
	CertFile string `mapstructure:"cert_file" json:"cert_file" yaml:"cert_file" toml:"cert_file"`
	KeyFile  string `mapstructure:"key_file" json:"key_file" yaml:"key_file" toml:"key_file"`
	Verify   bool   `json:"verify" yaml:"verify" toml:"verify"`
}

// Validators defines various validation regular expressions
type Validators struct {
	ParamTypeRegex           string `mapstructure:"param_type_regex" json:"param_type_regex" yaml:"param_type_regex" toml:"param_type_regex"`
	ParamDistroRegex         string `mapstructure:"param_distro_regex" json:"param_distro_regex" yaml:"param_distro_regex" toml:"param_distro_regex"`
	IsRHELDistroRegex        string `mapstructure:"is_rhel_distro_regex" json:"is_rhel_distro_regex" yaml:"is_rhel_distro_regex" toml:"is_rhel_distro_regex"`
	IsSolarisDistroRegex     string `mapstructure:"is_solaris_distro_regex" json:"is_solaris_distro_regex" yaml:"is_solaris_distro_regex" toml:"is_solaris_distro_regex"`
	ParamVersionRegex        string `mapstructure:"param_version_regex" json:"param_version_regex" yaml:"param_version_regex" toml:"param_version_regex"`
	ParamVersionCleanerRegex string `mapstructure:"param_version_cleaner_regex" json:"param_version_cleaner_regex" yaml:"param_version_cleaner_regex" toml:"param_version_cleaner_regex"`
	ParamArchRegex           string `mapstructure:"param_arch_regex" json:"param_arch_regex" yaml:"param_arch_regex" toml:"param_arch_regex"`
	ParamAgentModeRegex      string `mapstructure:"param_agent_mode_regex" json:"param_agent_mode_regex" yaml:"param_agent_mode_regex" toml:"param_agent_mode_regex"`
	TemplateTypeRegex        string `mapstructure:"template_type_regex" json:"template_type_regex" yaml:"template_type_regex" toml:"template_type_regex"`
	TemplateNameRegex        string `mapstructure:"template_name_regex" json:"template_name_regex" yaml:"template_name_regex" toml:"template_name_regex"`
}

// Config defines the running config structure
type Config struct {
	Listen            []string   `json:"listen" yaml:"listen" toml:"listen"`
	ContentPath       string     `mapstructure:"content_path" json:"content_path" yaml:"content_path" toml:"content_path"`
	PackageConfigFile string     `mapstructure:"package_config_file" json:"package_config_file" yaml:"package_config_file" toml:"package_config_file"`
	PackageBaseURL    string     `mapstructure:"package_base_url" json:"package_base_url" yaml:"package_base_url" toml:"package_base_url"`
	SSL               SSL        `json:"ssl" yaml:"ssl" toml:"ssl"`
	CacheTemplates    bool       `mapstructure:"enable_template_cache" json:"enable_template_cache" yaml:"enable_template_cache" toml:"enable_template_cache"`
	Validators        Validators `json:"validators" yaml:"validators" toml:"validators"`
	Brokers           Brokers    `json:"brokers" yaml:"brokers" toml:"brokers"`
	RPMFile           string     `mapstructure:"rpm_file" json:"rpm_file" yaml:"rpm_file" toml:"rpm_file"`
	Statsd            Statsd     `json:"statsd" yaml:"statsd" toml:"statsd"`
	Debug             bool       `json:"debug" yaml:"debug" toml:"debug"`
	Log               Log        `json:"log" yaml:"log" toml:"log"`
	LocalPackages     bool       `mapstructure:"local_packages"`
	LocalPackagePath  string     `mapstructure:"local_package_path"`
	CosiToolVersion   string     `mapstructure:"cosi_tool_version"`
	CosiToolBaseURL   string     `mapstructure:"cosi_tool_base_url"`
}

//
// NOTE: adding a Key* MUST be reflected in the Config structures above
//
const (
	// KeyListen primary address and port to listen on
	KeyListen = "listen"

	// KeyContentPath content directory
	KeyContentPath = "content_path"

	// KeyPackageConfigFile defines the package configuration file
	KeyPackageConfigFile = "package_config_file"

	// KeyPackageBaseURL defines the base url for packages
	KeyPackageBaseURL = "package_base_url"

	// KeyLocalPackages toggles serving agent packages from local directory
	KeyLocalPackages = "local_packages"
	// KeyPackagePath defines directory from which to serve local packages
	KeyLocalPackagePath = "local_package_path"

	// KeyCosiToolVersion defines the version of the cosi tool to install
	KeyCosiToolVersion = "cosi_tool_version"
	// KeyCosiToolBaseURL defines the base url from which to retrieve the cosi tool file
	KeyCosiToolBaseURL = "cosi_tool_base_url"

	//
	// SSL
	//

	// KeySSLListen ssl address and prot to listen on
	KeySSLListen = "ssl.listen"

	// KeySSLCertFile pem certificate file for SSL
	KeySSLCertFile = "ssl.cert_file"

	// KeySSLKeyFile key for ssl.cert_file
	KeySSLKeyFile = "ssl.key_file"

	// KeySSLVerify controls verification for ssl connections
	KeySSLVerify = "ssl.verify"

	// KeyEnableTemplateCache controls template caching
	KeyEnableTemplateCache = "enable_template_cache"

	// KeyParamTypeRx defines the parameter 'type' (os type) validation regular expression
	KeyParamTypeRx = "validators.param_type_regex"

	// KeyParamDistroRx defines the parameter 'dist' (os distro) validation regular expression
	KeyParamDistroRx = "validators.param_distro_regex"
	// KeyIsRHELDistroRx defines the is rhel distro regular expression
	KeyIsRHELDistroRx = "validators.is_rhel_distro_regex"
	// KeyIsSolarisDistroRx defines the is solaris distro regular expression
	KeyIsSolarisDistroRx = "validators.is_solaris_distro_regex"

	// KeyParamVersionRx defines the parameter 'vers' (os version) validation regular expression
	KeyParamVersionRx = "validators.param_version_regex"
	// KeyParamVersionCleanerRx defines the parameter 'vers' (os version) cleaner regular expression
	KeyParamVersionCleanerRx = "validators.param_version_cleaner_regex"

	// KeyParamArchRx defines the parameter 'arch' (system architecture) validation regular expression
	KeyParamArchRx = "validators.param_arch_regex"

	// KeyParamAgentModeRx defines the parameter 'agent' (agent mode) validation regular expression
	KeyParamAgentModeRx = "validators.param_agent_mode_regex"
	// KeyAgentPushModeRx defines the default regular expression used to determine if the agent/broker is PUSH mode
	KeyAgentPushModeRx = "validators.agent_push_mode_regex"
	// KeyAgentPullModeRx defines the default regular expression used to determine if the agent/broker is PULL mode
	KeyAgentPullModeRx = "validators.agent_pull_mode_regex"

	// KeyTemplateTypeRx defines the template type validation regular expression
	KeyTemplateTypeRx = "validators.template_type_regex"

	// KeyTemplateNameRx defines the template name validation regular expression
	KeyTemplateNameRx = "validators.template_name_regex"

	// KeyBrokerFallbackList list of fallback broker IDs
	KeyBrokerFallbackList = "brokers.fallback"
	// KeyBrokerFallbackDefault index into fallback broker list to use as default
	KeyBrokerFallbackDefault = "brokers.fallback_default"
	// KeyBrokerPushList list of push broker IDs
	KeyBrokerPushList = "brokers.push"
	// KeyBrokerPushDefault index into push broker list to use as default
	KeyBrokerPushDefault = "brokers.push_default"
	// KeyBrokerPullList list of pull broker IDs
	KeyBrokerPullList = "brokers.pull"
	// KeyBrokerPullDefault index into pull broker list to use as default
	KeyBrokerPullDefault = "brokers.pull_default"

	// KeyRPMFile is the name of the RPM to server for /install/rpm/
	KeyRPMFile = "rpm_installer_file"

	//
	// statsd client
	//

	// KeyStatsdAddress defines the network address of a StatsD server
	KeyStatsdAddress = "statsd.address"
	// KeyStatsdInterval defines the submission interval
	KeyStatsdInterval = "statsd.interval"
	// KeyStatsdPrefix defines the prefix for all metric names
	KeyStatsdPrefix = "statsd.prefix"

	//
	// generic flags
	//

	// KeyDebug enables debug messages
	KeyDebug = "debug"

	// KeyLogLevel logging level (panic, fatal, error, warn, info, debug, disabled)
	KeyLogLevel = "log.level"

	// KeyLogPretty output formatted log lines (for running in foreground)
	KeyLogPretty = "log.pretty"

	//
	// intentionally *NOT* reflected in config, they are command line flags ONLY
	//

	// KeyShowConfig - show configuration and exit
	KeyShowConfig = "show_config"

	// KeyShowVersion - show version information and exit
	KeyShowVersion = "version"
)
