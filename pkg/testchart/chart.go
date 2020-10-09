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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/karrick/godirwalk"
	"github.com/mholt/archiver/v3"
	copy "github.com/otiai10/copy"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

type TestChart struct {
	name            string
	version         string
	defaults        map[string]interface{}
	runtimeDefaults map[string]interface{}

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

func (t *TestChart) RuntimeDefaults() map[string]interface{} {
	return t.runtimeDefaults
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

func (t *TestChart) loadRuntimeDefaults(chartpath string) error {
	var defaults values
	runtime := filepath.Join(chartpath, "runtime.yaml")
	if _, err := os.Stat(runtime); err == nil {
		dat, err := ioutil.ReadFile(runtime)
		if err != nil {
			return errors.Wrap(err, "while reading runtime file from test chart")
		}

		if err := yaml.Unmarshal(dat, &defaults); err != nil {
			return errors.Wrap(err, "while unmarshalling runtime file from test chart")
		}

		t.runtimeDefaults = defaults
	}

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

func compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// walk through every file in the folder
	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		name := strings.ReplaceAll(file, src, "")
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(name)

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}

func (t *TestChart) Package(chartpath, dest string) error {
	if err := t.loadMeta(chartpath); err != nil {
		return errors.Wrap(err, "while reading test metadata")
	}

	var buf bytes.Buffer
	if err := compress(chartpath, &buf); err != nil {
		return err
	}
	// write the .tar.gzip
	fileToWrite, err := os.OpenFile(filepath.Join(dest, fmt.Sprintf("%s-%s.tar.gz", t.name, t.version)), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fileToWrite, &buf); err != nil {
		return err
	}
	return nil
	//	return archiver.Archive([]string{chartpath}, filepath.Join(dest, fmt.Sprintf("%s-%s.tar.gz", t.name, t.version)))
}

func (t *TestChart) Load(chartpath string) error {

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
	} else if strings.Contains(chartpath, "tar.gz") {
		// Get chart if it's not a folder

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
	if err := t.loadRuntimeDefaults(chartpath); err != nil {
		return errors.Wrap(err, "while reading test runtime")
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

	// copy static
	static := filepath.Join(chartpath, "static")
	if _, err := os.Stat(static); err == nil {
		if err := copy.Copy(static, filepath.Join(t.tmpExecutionDir, "static")); err != nil {
			return err
		}
	}

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

	return out[t.name+"/templates"], nil
}
