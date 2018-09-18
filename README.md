# PRE-RELEASE

# Circonus One Step Install

Circonus One Step Install (cosi) is comprised of two separate pieces. One, this repository, the `cosi-server` and second, `cosi-tool`.

## cosi-server

This repository contains the server used during the installation process. The `cosi-server` provides the installation script, whether a specific OS is supported, what agent package to use, and templates for creating visualizations.

## cosi-tool

The `cosi-tool` repository contains a command line client used to perform the registration of a system with Circonus. It leverages the `cosi-server` to obtain templates for creating visuals. Additionally, the `cosi-tool` provides a set of commands for interacting with the local cosi registration.

---

> NOTE: this repository is in **active** development! As such, it may not be entirely feature complete, may contain bugs, and should **not** be used in production at this time.

## Installation

1. Download from [latest release](https://github.com/circonus-labs/cosi-server/releases/latest)
1. Create an installation directory (e.g. `mkdir -p /opt/circonus/cosi-server`)
1. Unpack release archive into installation directory
1. See `sbin/cosi-serverd --help` and `etc/example-cosi-server.yaml` for help configuring


## TODO

1. waiting on `cosi-tool`
1. deployment configuration (ansible)
