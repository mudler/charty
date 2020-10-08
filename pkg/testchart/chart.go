/*
Copyright Ettore Di Giacinto <mudler@gentoo.org>.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package chart

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/karrick/godirwalk"
	"github.com/mholt/archiver/v3"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

type TestChart struct {
	name     string
	version  string
	defaults map[string]interface{}

	tmpExecutionDir string

	Values map[string]interface{}
}

type chartMeta struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}
type values map[string]interface{}

func (t *TestChart) RunnerDirectory() string {
	return t.tmpExecutionDir
}

func (t *TestChart) Name() string {
	return t.name
}

func (t *TestChart) Version() string {
	return t.version
}

func (t *TestChart) Defaults() values {
	return t.defaults
}

func (t *TestChart) Cleanup() error {
	return os.RemoveAll(t.tmpExecutionDir)
}

func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func (t *TestChart) loadDefaults(chartpath string) error {

	// load default values
	var defaults values

	dat, err := ioutil.ReadFile(filepath.Join(chartpath, "values.yaml"))
	if err != nil {
		return errors.Wrap(err, "while reading values file from test chart")
	}

	if err := yaml.Unmarshal(dat, &defaults); err != nil {
		return errors.Wrap(err, "while unmarshalling values file from test chart")
	}

	t.defaults = defaults
	return nil
}

func (t *TestChart) loadMeta(chartpath string) error {

	var meta chartMeta
	dat, err := ioutil.ReadFile(filepath.Join(chartpath, "metadata.yaml"))
	if err != nil {
		return errors.Wrap(err, "while reading metadata file from test chart")
	}

	if err := yaml.Unmarshal(dat, &meta); err != nil {
		return errors.Wrap(err, "while unmarshalling metadata file from test chart")
	}

	t.name = meta.Name
	t.version = meta.Version
	return nil
}

func (t *TestChart) Load(chartpath string) error {

	_, err := os.Stat(chartpath)
	// Get chart if it's not a folder
	if os.IsNotExist(err) {
		// not a dir

		if isValidUrl(chartpath) {
			tmpfile, err := ioutil.TempFile(os.TempDir(), "example")
			if err != nil {
				return errors.Wrap(err, "while creating tempfile")
			}

			defer os.Remove(tmpfile.Name()) // clean up
			//download and extract
			err = downloadFile(tmpfile.Name(), chartpath)
			if err != nil {
				return errors.Wrap(err, "while downloading chart")
			}

			chartpath = tmpfile.Name()
		}

		// Extract archives
		dir, err := ioutil.TempDir(os.TempDir(), "prefix")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)

		err = archiver.Unarchive(chartpath, dir)
		chartpath = dir
	}

	// prepare dir for the runner
	dir, err := ioutil.TempDir(os.TempDir(), "prefix")
	if err != nil {
		return errors.Wrap(err, "while creating tempdir")
	}
	t.tmpExecutionDir = dir

	if err := t.loadDefaults(chartpath); err != nil {
		return errors.Wrap(err, "while reading test chart defaults")
	}
	if err := t.loadMeta(chartpath); err != nil {
		return errors.Wrap(err, "while reading test metadata")
	}

	// render templates
	templates := filepath.Join(chartpath, "templates")
	err = godirwalk.Walk(templates, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			relativepath := strings.ReplaceAll(osPathname, strings.TrimSuffix(chartpath, "/"), "")
			relativepath = strings.ReplaceAll(relativepath, "/templates", "")

			if de.IsDir() {
				os.MkdirAll(filepath.Join(t.tmpExecutionDir, relativepath), os.ModePerm)
				return nil //godirwalk.SkipThis
			}

			dat, err := ioutil.ReadFile(osPathname)
			if err != nil {
				return errors.Wrap(err, "while reading source data")
			}
			rendered, err := t.render(string(dat))
			if err != nil {
				return errors.Wrap(err, "while rendering template")
			}

			if err := ioutil.WriteFile(filepath.Join(t.tmpExecutionDir, relativepath), []byte(rendered), os.ModePerm); err != nil {
				return errors.Wrap(err, "while writing `"+relativepath+"` from template")
			}

			return nil
		},
		Unsorted: true,
	})

	if err != nil {
		return err
	}
	return nil
}

func (t *TestChart) render(template string) (string, error) {
	c := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    t.name,
			Version: t.version,
		},
		Templates: []*chart.File{
			{Name: "templates", Data: []byte(template)},
		},
		Values: map[string]interface{}{"Values": t.defaults},
	}

	v, err := chartutil.CoalesceValues(c, map[string]interface{}{"Values": t.Values})
	if err != nil {
		return "", errors.Wrap(err, "while interpolating template with default variables")
	}
	out, err := engine.Render(c, v)
	if err != nil {
		return "", errors.Wrap(err, "while rendering template")
	}

	return out["templates"], nil
}
