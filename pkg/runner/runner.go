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

package runner

import (
	"bytes"
	"io"
	"os"

	"github.com/codeskyblue/kexec"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Chart interface {
	RunnerDirectory() string
	RuntimeDefaults() map[string]interface{}
}

type Options struct {
	Commands Commands `yaml:"commands"`
	Pre      []string `yaml:"pre"`
	Post     []string `yaml:"post"`
}

type TestRunner struct{}

func (t *TestRunner) runAndFail(c []string, path string) (string, error) {
	var o string
	for _, p := range c {
		out, err := runProc(p, path)
		if err != nil {
			return o, errors.Wrap(err, "failed running "+p)
		}
		o = o + out
	}
	return o, nil
}

func interfaceToOptions(m map[string]interface{}) (Options, error) {
	dat, err := yaml.Marshal(m)
	if err != nil {
		return Options{}, err
	}
	var opts Options
	err = yaml.Unmarshal(dat, &opts)
	return opts, err
}

func (t *TestRunner) Run(c Chart, o Options) ([]CommandOutput, error) {
	res := []CommandOutput{}
	var ret error

	// Merge runtime options with what provided from the chart
	opts, err := interfaceToOptions(c.RuntimeDefaults())
	if err != nil {
		return res, err
	}

	if err := mergo.Merge(&opts, o, mergo.WithOverride); err != nil {
		return res, err
	}

	if out, err := t.runAndFail(opts.Pre, c.RunnerDirectory()); err != nil {
		res = append(res, CommandOutput{Command: Command{Name: "global-pre-run"}, Error: err, Output: out})
		ret = multierror.Append(ret, err)
		return res, ret
	}

	results := opts.Commands.Start(c.RunnerDirectory())
	for _, r := range results {
		if r.Error != nil {
			ret = multierror.Append(ret, r.Error)
		}
	}
	res = append(res, results...)

	if out, err := t.runAndFail(opts.Post, c.RunnerDirectory()); err != nil {
		res = append(res, CommandOutput{Command: Command{Name: "global-post-run"}, Error: err, Output: out})
		ret = multierror.Append(ret, err)
		return res, ret
	}

	return res, ret
}

func runProc(cmd, dir string) (string, error) {

	p := kexec.CommandString(cmd)

	var b bytes.Buffer
	p.Stdout = io.MultiWriter(os.Stdout, &b)
	p.Stderr = io.MultiWriter(os.Stderr, &b)
	p.Dir = dir
	if err := p.Run(); err != nil {
		return b.String(), err
	}

	p.Wait()

	return b.String(), nil
}
