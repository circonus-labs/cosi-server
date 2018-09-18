// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package templates

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/circonus-labs/cosi-server/api"
	"github.com/circonus-labs/cosi-server/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// New creates new instance of Templates
func New() (*Templates, error) {
	t := Templates{
		logger:   log.With().Str("pkg", "templates").Logger(),
		useCache: viper.GetBool(config.KeyEnableTemplateCache),
		fileExt:  api.TemplateFileExtension,
		cache:    map[string][]byte{},
		index:    map[string]string{},
	}

	trx, err := regexp.Compile(viper.GetString(config.KeyTemplateTypeRx))
	if err != nil {
		return nil, errors.Wrap(err, "typerx compile")
	}
	t.Typerx = trx

	nrx, err := regexp.Compile(viper.GetString(config.KeyTemplateNameRx))
	if err != nil {
		return nil, errors.Wrap(err, "namerx compile")
	}
	t.Namerx = nrx

	t.templateDir = viper.GetString(config.KeyContentPath)
	if t.templateDir == "" {
		return nil, errors.New("content path not set")
	}

	t.templateDir = path.Join(t.templateDir, "templates")
	stat, err := os.Stat(t.templateDir)
	if err != nil {
		return nil, errors.Wrap(err, "invalid template path (access)")
	}
	if !stat.IsDir() {
		return nil, errors.New("invalid template path (not a directory)")
	}

	return &t, nil
}

// Get a specific template
func (t *Templates) Get(osType, osDist, osVers, osArch, tType, tName string) (*[]byte, error) {
	// Note: os* vars would already been validated in the request handler
	//       passing blanks will simply result in the default being returned.
	if tType == "" || !t.Typerx.MatchString(tType) {
		return nil, errors.New("invalid template type")
	}
	if tName == "" || !t.Namerx.MatchString(tName) {
		return nil, errors.New("invalid template name")
	}

	spec := &tspec{
		ttype:   strings.ToLower(tType),
		tname:   strings.ToLower(tName),
		ostype:  strings.ToLower(osType),
		osdist:  strings.ToLower(osDist),
		osvers:  strings.ToLower(osVers),
		sysarch: strings.ToLower(osArch),
	}

	tlist := t.makeTemplateList(spec)

	template, err := t.getTemplate(&tlist)
	if err != nil {
		return nil, errors.Wrap(err, "get template")
	}

	return template, nil
}

func (t *Templates) getTemplate(tlist *[]tinfo) (*[]byte, error) {
	foundIdx := 0
	foundKey := ""
	template := []byte{}

	for idx, ti := range *tlist {
		if t.useCache {
			if tkey, inIndex := t.index[ti.key]; inIndex {
				if data, cached := t.cache[tkey]; cached {
					return &data, nil
				}
			}
		}

		data, err := ioutil.ReadFile(ti.filename)
		if err != nil {
			if os.IsNotExist(err) {
				continue // ignore, try next template spec
			}
			return nil, err
		}
		if len(data) == 0 {
			return nil, errors.New("invalid template found (empty)")
		}

		foundIdx = idx
		foundKey = ti.key
		if t.useCache {
			t.cache[foundKey] = data
		}
		template = data
		break
	}

	if foundKey == "" {
		key := (*tlist)[0].key // will be the *most* specific spec
		t.logger.Warn().Str("spec", key).Msg("no template found for spec")
		return nil, errors.New("no template found")
	}

	if t.useCache {
		// add found key to index for all of the more specific keys preceding it
		// to short-circuit hitting the filesystem constantly checking for files
		// that will not be found going forward
		for i := foundIdx; i >= 0; i-- {
			key := (*tlist)[i].key
			t.index[key] = foundKey
		}
	}

	return &template, nil
}

func (t *Templates) makeTemplateList(s *tspec) []tinfo {
	templateFileName := fmt.Sprintf("%s-%s%s", s.ttype, s.tname, t.fileExt)
	sep := "-"
	tlist := []tinfo{}

	if s.ostype != "" {
		if s.osdist != "" {
			if s.osvers != "" {
				if s.sysarch != "" {
					tlist = append(tlist, tinfo{
						key:      strings.Join([]string{s.ostype, s.osdist, s.osvers, s.sysarch, s.ttype, s.tname}, sep),
						filename: path.Join(t.templateDir, s.ostype, s.osdist, s.osvers, s.sysarch, templateFileName),
					})
				}
				tlist = append(tlist, tinfo{
					key:      strings.Join([]string{s.ostype, s.osdist, s.osvers, s.ttype, s.tname}, sep),
					filename: path.Join(t.templateDir, s.ostype, s.osdist, s.osvers, templateFileName),
				})
			}
			tlist = append(tlist, tinfo{
				key:      strings.Join([]string{s.ostype, s.osdist, s.ttype, s.tname}, sep),
				filename: path.Join(t.templateDir, s.ostype, s.osdist, templateFileName),
			})
		}
		tlist = append(tlist, tinfo{
			key:      strings.Join([]string{s.ostype, s.ttype, s.tname}, sep),
			filename: path.Join(t.templateDir, s.ostype, templateFileName),
		})
	}

	tlist = append(tlist, tinfo{
		key:      strings.Join([]string{s.ttype, s.tname}, sep),
		filename: path.Join(t.templateDir, templateFileName),
	})

	return tlist
}
