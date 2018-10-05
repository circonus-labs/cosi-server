# PRE-RELEASE Preview

A _pre-release preview_ of cosi w/[circonus-agent](https://github.com/circonus-labs/circonus-agent) is now available.

Based on the normal cosi command available in the Circonus UI (Integrations>Hosts>[New+]), make the following modifications:

```
curl -sSL https://setup.circonus.com/install | bash \
    -s -- \
    --cosiurl https://setup.circonus.com/ \
    --key <insert api key from original cosi command here> \
    --app <insert api app from original cosi command here>
```

After modifying the command it will use the new cosi version to install the circonus-agent.

The circonus-agent is comprised of:

  * replacement for NAD, written in go, with builtin plugins for the common metrics needed for cosi visuals (graphs, worksheets, & dashboards)
  * includes (if OS supports) [protocol_observer](https://github.com/circonus-labs/wirelatency), no longer needs to be built/installed manually
  * includes (if OS supports) [circonus-logwatch](https://github.com/circonus-labs/circonus-logwatch), no longer needs to be installed manually
  * includes OS/version/architecture-specific NAD plugins (non-javascript only) -- **Note:** the circonus-agent is **not** capable of using NAD _native plugins_ since they require NodeJS

The cosi-tool does **not** currently include a functional `cosi plugin` command. This capability will be included in a future release, as the individual `cosi plugin ...` sub-commands (postgres and cassandra) are completed.

Supported Operating Systems (x86_64 and/or amd64):

  * RHEL7 (CentOS, RedHat, Oracle)
  * RHEL6 (CentOS, RedHat, amzn)
  * Ubuntu18
  * Ubuntu16
  * Ubuntu14
  * Debian9
  * Debian8
  * FreeBSD 12
  * FreeBSD 11

Please continue to use the original cosi(w/NAD) for OmniOS and Raspian - cosi v2 support for these is TBD. Note: after installing NAD a binary circonus-agent can be used as a drop-in replacement (configure circonus-agent _plugins directory_ to be NAD plugins directory -- javascript plugins will not function). Binaries for OmniOS (`solaris_x86_64`) and Raspian (`linux_arm`) are available in the [circonus-agent repository](https://github.com/circonus-labs/circonus-agent/releases/latest).

---

# Circonus One Step Install

Circonus One Step Install (cosi) is comprised of two separate pieces.

1. [`cosi-tool`](https://github.com/circonus-labs/cosi-tool) contains the command line tool used to register a system with Circonus and manage the local registration.
1. [`cosi-server`](https://github.com/circonus-labs/cosi-server), this repository, contains the server used during the installation and registration process. It serves the installation script, whether a specific OS is supported, what [`circonus-agent`](https://github.com/circonus-labs/circonus-agent) package to use, and templates for creating assets in the Circonus system (checks, dashboards, graphs, rulesets, and worksheets).

---

# Circonus One Step Install Server

> NOTE: this repository is in **active** development! As such, it may not be entirely feature complete, may contain bugs, and should **not** be used in production at this time.

## Installation

1. Download from [latest release](https://github.com/circonus-labs/cosi-server/releases/latest)
1. Create an installation directory (e.g. `mkdir -p /opt/circonus/cosi-server`)
1. Unpack release archive into installation directory
1. See `sbin/cosi-serverd --help`
    1. Configure `etc/example-cosi-server.yaml` (edit, rename `cosi-server.yaml`)
    1. Configure `etc/example-circonus-packages.yaml` (edit, rename `circonus-packages.yaml`)

Unless otherwise noted, the source files are distributed under the BSD-style license found in the [LICENSE](LICENSE) file.
