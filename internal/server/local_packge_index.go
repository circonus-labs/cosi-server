// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package server

import (
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type Release struct {
	Version  string
	Packages []string
}

// updateLocalPackageIndex manages the index.html served on the
// packages/ endpoint when serving packages locally
func updateLocalPackageIndex(pkgDir string) error {
	releases, err := scanPackages(pkgDir)
	if err != nil {
		return err
	}

	if err := writeIndex(pkgDir, releases); err != nil {
		return err
	}
	return nil
}

func scanPackages(pkgDir string) ([]Release, error) {
	if pkgDir == "" {
		return nil, errors.New("invalid package path (empty)")
	}

	fl, err := ioutil.ReadDir(pkgDir)
	if err != nil {
		return nil, err
	}

	pf := []string{}
	for _, fi := range fl {
		if fi.IsDir() {
			continue
		}
		if !strings.HasPrefix(fi.Name(), "circonus-agent") {
			continue
		}

		pf = append(pf, fi.Name())
	}

	return releases(pf)
}

func releases(packageFiles []string) ([]Release, error) {
	if len(packageFiles) == 0 {
		return nil, errors.New("invalid package file list (empty)")
	}

	pkgsByVer := map[string][]string{}
	for _, pkgFileName := range packageFiles {
		// compensate for semver pre-release versions (e.g. alpha/beta/rc/etc.)
		// RHEL based packages will not build without replacing the '-' in semver
		// format 1.0.0-alpha.1
		pkgName := strings.Replace(pkgFileName, `~`, `-`, -1)

		parts := strings.Split(pkgName, "-")
		if len(parts) < 3 {
			continue
		}
		ver := parts[2]
		if len(parts) > 3 {
			if !strings.HasPrefix(parts[3], `1.`) {
				ver += "-" + parts[3]
			}
		}
		if _, ok := pkgsByVer[ver]; !ok {
			pkgsByVer[ver] = []string{}
		}
		pkgsByVer[ver] = append(pkgsByVer[ver], pkgFileName)
	}

	verList := []string{}
	for ver := range pkgsByVer {
		verList = append(verList, ver)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(verList)))

	releases := []Release{}
	for _, ver := range verList {
		releases = append(releases, Release{
			Version:  ver,
			Packages: pkgsByVer[ver],
		})
	}

	return releases, nil
}

func writeIndex(pkgDir string, releases []Release) error {
	tmplDoc := `
    <!DOCTYPE html>
    <html>
    <head><title>Circonus Agent Packages</title><meta charset="UTF-8"></head>
    <body>
	<h1>Circonus Agent Packages</h1>
	<p>
	Note: These are packages used by cosi only. If target operating system or architecture is 
	not listed here (e.g. Windows, Illumos, Arm), please check the circonus-agent
	<a href="https://github.com/circonus-labs/circonus-agent/releases">releases</a> 
	page as there may be an agent available.
	</p>
    {{range .}}
    <h4>v{{ .Version}}</h4>
    <ul>
    <li> <a href="https://github.com/circonus-labs/circonus-agent/releases/tag/v{{ .Version }}">release information</a></li>
    {{range .Packages}}
    <li><a href="{{ . }}">{{ . }}</a></li>
    {{end}}
    </ul>
    {{end}}
    </body></html>
    `

	tmpl, err := template.New("index").Parse(tmplDoc)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(pkgDir, "index.html"))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(f, releases); err != nil {
		f.Close()
		return err
	}

	return f.Close()
}
