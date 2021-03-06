# v0.5.8

* upd: dependencies
* upd: semver sort packages

# v0.5.7

* upd: lint fix
* fix: templates for plugins with dashes in name

# v0.5.6

* fix: disk template to use {{.CheckID}} for check_id attribute

# v0.5.5 [CIRC-4272]

* fix: correct rc.conf agent name for freebsd
* fix: installer script, validate pkg extension against rpm|deb|tgz rather than tar.gz

# v0.5.4

* upd: handle pre-release packages in index builder
* upd: dependencies
* add: amzn 2 linux config
* upd: strict linting
* upd: build vars moved to release pkg
* add: go tidy and lint before hooks
* upd: config syntax for archives
* add: illumos target
* upd: rename/reorg release sub-pkg

# v0.5.3

* add: `host` as a default tag
* upd: dependencies
* fix: local package index template, close anchor tag in package links

# v0.5.2

* add: local package index.html management
* upd: installer to exit when NAD install detected, with pointer to agent repo manual instructions
* fix: ensure streamtags are enabled in agent before running registration

# v0.5.1

* fix: option names in config file
* upd: installer error message when process not found after start

# v0.5.0

* upd: add a `--timeout` argument (for curl command) to installer, default: `--timeout 120`
* upd: enable stream tags by default (circonus-agent v0.19.0+ required)
* upd: dependencies

# v0.4.3

* upd: default cosiurl to setup.circonus.com

# v0.4.2

* upd: dependencies (yaml)
* upd: api example

# v0.4.1

* upd: example circonus-packages to use circonus-agent instead of nad
* upd: dependencies

# v0.4.0

* upd: release file names use x86_64, facilitate automated builds and testing

# v0.3.0

* doc: pre-release comment
* doc: repo link
* upd: release, switch owner
* upd: release, tighten etc file glob
* upd: dependencies
* upd: refactor/condense api, prep for pre-alpha release

# v0.2.2

* fix: diskstats datapoint formulas
* fix: lower case dir names for template tests
* fix: toml key `filters` not `filter` for graph templates
* fix: typo, restart not start for init.d after agent config update
* fix: remove ruleset create, already in registration

# v0.2.1

* upd: add request query parameters to `tool/` endpoint tests
* upd: switch nad to circonus-agent in installer
* fix: switch broker IDs to strings so a slice can be retrieved from viper
* upd: turn off draft for releases

# v0.2.0

* add: tool config (version/base url) to facilitate testing tool packages in dev
* fix: package request url
* fix: remove line ending from text/plain package response
* upd: change `circonus-packages.yaml` to `example-circonus-packages.yaml` installed during provisioning
* upd: go1.11 build constraint

# v0.1.2

* add: local package serving for testing agent dev packages (e.g. cosi-examples/server)
* upd: goreleaser glob for `content/`
* fix: template subdirectories must be lowercase (e.g. linux not Linux, omnios not OmniOS)
* fix: graph config name in Systemd
* fix: graph config name in FS
* fix: typo and missing comma in IF graphs
* fix: misplaced comma after sorting keys in CPU graph
* fix: add `legend_formula`s to composites in VM graph, was not in original template, field is required

# v0.1.1

* upd: return broker id as string in json
* add: systemd service configuration

# v0.1.0

* upd: upstream dependencies
* upd: templates
* add: dashboard template structs to api

# v0.0.4

* add: api.FetchBroker
* upd: switch templates to toml - simplified processing and json embedding
* upd: slight optimization to config file loading error handling

# v0.0.3

* add: api examples
* add: api package
* doc: start api package doc
* upd: handle templates with % in strings
* fix: omnios arch in package list
* doc: updates/corrections
* upd: go1.10 build constraint

# v0.0.2

* fix: 2 calls to json.Unmarshal (in test)
* fix: transposed vars in stat name
* fix: replace untrappable os.Kill with syscall.SIGTERM
* upd: remove redundant string in server
* upd: remove double return in http response text
* fix: simplify func/select remove unreachable return
* upd: simplify (gofmt -s)

# v0.0.1

* initial conversion from NodeJS
