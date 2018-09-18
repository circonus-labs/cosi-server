// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package defaults

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/circonus-labs/cosi-server/internal/release"
)

const (
	// ListenPort is the default agent tcp listening port
	ListenPort = 80

	// BasePackageURL defines the default url to retrieve agent packages from
	BasePackageURL = "http://updates.circonus.net/node-agent/packages"

	// LocalPackages toggles serving packages locally vs from public package server (for testing new agent packages)
	LocalPackages = false

	// EnableTemplateCache controls whether templates are cached
	EnableTemplateCache = true

	// ParamTypeRx defines the default 'type' (os type) parameter validation regular expression
	ParamTypeRx = `^(?i)[a-z-_]+$`

	// ParamDistroRx defines the default 'dist' (os distribution) parameter validation regular expression
	ParamDistroRx = `^(?i)[a-z]+$`
	// IsRHELDistroRx defines the default regular expression used to determine if os distro is a rhel type
	IsRHELDistroRx = `^(?i)(CentOS|Fedora|RedHat|Oracle)$`
	// IsSolarisDistroRx defines the default regular expression used to determine if os distro is a solaris type (using 'pkg' to manage packages)
	IsSolarisDistroRx = `^(?i)(OmniOS|Illumos|Solaris)$`

	// ParamVersionRx defines the default 'vers' (os version) parameter validation regular expression
	ParamVersionRx = `^[rv]?\d+(\.\d+)*$`
	// ParamVersionCleanerRx defines the default version cleaner regular expression
	ParamVersionCleanerRx = `^[rv]`

	// ParamArchRx defines the default 'arch' (system architecture) parameter validation regular expression
	ParamArchRx = `^(amd64|x86_64|i386|i686)$`

	// ParamAgentModeRx defines the default 'agent' (agent mode) parameter validation regular expression
	ParamAgentModeRx = `^(?i)(reverse|pull|push|revonly)$`
	// AgentPushModeRx defines the default regular expression used to determine if the agent/broker is PUSH mode
	AgentPushModeRx = `^(?i)(push|trap|httptrap)$`
	// AgentPullModeRx defines the default regular expression used to determine if the agent/broker is PULL mode
	AgentPullModeRx = `^(?i)(pull|reverse|revonly|json)$`

	// TemplateTypeRx defines the default template type validation regular expression
	TemplateTypeRx = `^(?i)(check|graph|worksheet|dashboard)$`

	// TemplateNameRx defines the default template name validation regular expression
	TemplateNameRx = "^(?i)[a-z0-9_-\\`]+$"

	// Debug is false by default
	Debug = false

	// LogLevel set to info by default
	LogLevel = "warn"

	// LogPretty colored/formatted output to stderr
	LogPretty = false

	// StatsdAddress defines the network address to which statsd metrics should be sent
	StatsdAddress = "127.0.0.1:8125"
	// StatsdInterval defines the submission interval for statsd metrics
	StatsdInterval = time.Duration(10 * time.Second)
	// StatsdPrefix defines a prefix to apply to all metrics
	StatsdPrefix = "cosi-server"
)

var (
	// BasePath is the "base" directory
	//
	// expected installation structure:
	// base        (e.g. /opt/circonus/cosi-server)
	//   /bin      (e.g. /opt/circonus/cosi-server/bin)
	//   /etc      (e.g. /opt/circonus/cosi-server/etc)
	//   /content  (e.g. /opt/circonus/cosi-server/content)
	//   /sbin     (e.g. /opt/circonus/cosi-server/sbin)
	BasePath = ""

	// EtcPath returns the default etc directory within base directory
	// (e.g. /opt/circonus/cosi-server/etc)
	EtcPath = ""

	// Listen defaults to all interfaces on the default ListenPort
	// valid formats:
	//      ip:port (e.g. 127.0.0.1:12345 - listen address 127.0.0.1, port 12345)
	//      ip (e.g. 127.0.0.1 - listen address 127.0.0.1, port ListenPort)
	//      port (e.g. 12345 (or :12345) - listen address all, port 12345)
	//
	Listen = fmt.Sprintf(":%d", ListenPort)

	// ContentPath returns the default content path
	// (e.g. /opt/circonus/cosi-server/content)
	ContentPath = ""

	// PackageConfigFile defines the default package configuration file
	PackageConfigFile = ""

	// SSLCertFile returns the deefault ssl cert file name
	SSLCertFile = "" // (e.g. /opt/circonus/cosi-server/etc/ccosi-server.pem)

	// SSLKeyFile returns the deefault ssl key file name
	SSLKeyFile = "" // (e.g. /opt/circonus/cosi-server/etc/ccosi-server.key)

	// SSLVerify enabled by default
	SSLVerify = true

	// RPMFile determines if the /install/rpm/ endpoint will be served
	RPMFile = ""

	httptrapBroker  = "35"
	arlingtonBroker = "1"
	sanjoseBroker   = "2"
	chicagoBroker   = "275"
	// BrokerFallbackList list of fallback brokers
	BrokerFallbackList = []string{arlingtonBroker, sanjoseBroker, chicagoBroker}
	// BrokerFallbackDefault index into fallback list
	BrokerFallbackDefault = 2
	// BrokerPushList list of push brokers
	BrokerPushList = []string{httptrapBroker}
	// BrokerPushDefault index into push list
	BrokerPushDefault = 0
	// BrokerPullList list of pull brokers
	BrokerPullList = []string{arlingtonBroker, sanjoseBroker, chicagoBroker}
	// BrokerPullDefault index into pull list
	BrokerPullDefault = 2

	// LocalPackagePath defines where to serve local agent package files from
	LocalPackagePath = ""

	// CosiToolVersion defines version tag for cosi tool to install
	CosiToolVersion = "v0.2.0"
	// CosiToolBaseURL defines the base URL
	CosiToolBaseURL = "https://github.com/circonus-labs/cosi-tool/releases/download"
)

func init() {
	var exePath string
	var resolvedExePath string
	var err error

	exePath, err = os.Executable()
	if err == nil {
		resolvedExePath, err = filepath.EvalSymlinks(exePath)
		if err == nil {
			BasePath = filepath.Clean(filepath.Join(filepath.Dir(resolvedExePath), ".."))
		}
	}

	if err != nil {
		fmt.Printf("Unable to determine path to binary %v\n", err)
		os.Exit(1)
	}

	EtcPath = filepath.Join(BasePath, "etc")
	ContentPath = filepath.Join(BasePath, "content")
	LocalPackagePath = filepath.Join(BasePath, "content", "packages")
	PackageConfigFile = filepath.Join(EtcPath, "circonus-packages.yaml")
	SSLCertFile = filepath.Join(EtcPath, release.NAME+".pem")
	SSLKeyFile = filepath.Join(EtcPath, release.NAME+".key")
}
