# PRE-RELEASE

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
