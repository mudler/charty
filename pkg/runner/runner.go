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
	multierror "github.com/hashicorp/go-multierror"
	"github.com/ionrock/procs"
	"github.com/pkg/errors"
)

type Chart interface {
	RunnerDirectory() string
}

type Command struct {
	Pre  string `yaml:"pre"`
	Post string `yaml:"post"`
	Run  string `yaml:"run"`
	Name string `yaml:"name"`
}
type Options struct {
	Commands Commands `yaml:"commands"`
	Pre      []string `yaml:"pre"`
	Post     []string `yaml:"post"`
}

type TestRunner struct{}
type Commands []Command

type CommandOutput struct {
	PreOutput, PostOutput, Output string
	Error                         error
	Command                       Command
}

func (c Command) Start(dir string) CommandOutput {
	var err error
	var preoutput, postoutput string
	var res error
	if len(c.Pre) > 0 {
		preoutput, res = runProc(c.Pre, dir)
		if res != nil {
			err = multierror.Append(err, res)
		}
	}

	run, res := runProc(c.Run, dir)
	if res != nil {
		err = multierror.Append(err, res)
	}

	if len(c.Post) > 0 {
		postoutput, res = runProc(c.Post, dir)
		if res != nil {
			err = multierror.Append(err, res)
		}
	}

	return CommandOutput{
		PreOutput:  preoutput,
		PostOutput: postoutput,
		Output:     run,
		Error:      err,
		Command:    c,
	}
}

func (l Commands) Start(dir string) []CommandOutput {
	var res []CommandOutput

	for _, t := range l {
		res = append(res, t.Start(dir))
	}
	return res
}

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

func (t *TestRunner) Run(c Chart, o Options) ([]CommandOutput, error) {
	res := []CommandOutput{}
	var ret error

	if out, err := t.runAndFail(o.Pre, c.RunnerDirectory()); err != nil {
		res = append(res, CommandOutput{Command: Command{Name: "global-pre-run"}, Error: err, Output: out})
		ret = multierror.Append(ret, err)
		return res, ret
	}

	results := o.Commands.Start(c.RunnerDirectory())
	for _, r := range results {
		if r.Error != nil {
			ret = multierror.Append(ret, r.Error)
		}
	}
	res = append(res, results...)

	if out, err := t.runAndFail(o.Post, c.RunnerDirectory()); err != nil {
		res = append(res, CommandOutput{Command: Command{Name: "global-post-run"}, Error: err, Output: out})
		ret = multierror.Append(ret, err)
		return res, ret
	}

	return res, ret
}

func runProc(cmd, dir string) (string, error) {
	p := procs.NewProcess(cmd)
	p.Dir = dir
	err := p.Run()
	if err != nil {
		return "", err
	}
	out, err := p.Output()
	return string(out), err
}
