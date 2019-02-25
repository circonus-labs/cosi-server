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

	releases := []Release{}

	packageFiles := map[string][]string{}
	for _, fi := range fl {
		if fi.IsDir() {
			continue
		}
		if !strings.HasPrefix(fi.Name(), "circonus-agent") {
			continue
		}

		parts := strings.Split(fi.Name(), "-")
		if len(parts) < 3 {
			continue
		}
		ver := parts[2]
		if _, ok := packageFiles[ver]; !ok {
			packageFiles[ver] = []string{}
		}
		packageFiles[ver] = append(packageFiles[ver], fi.Name())
	}

	verList := []string{}
	for ver := range packageFiles {
		verList = append(verList, ver)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(verList)))

	for _, ver := range verList {
		releases = append(releases, Release{
			Version:  ver,
			Packages: packageFiles[ver],
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
