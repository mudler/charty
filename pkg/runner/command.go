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
	log "github.com/sirupsen/logrus"
)

type Command struct {
	Pre  string `yaml:"pre"`
	Post string `yaml:"post"`
	Run  string `yaml:"run"`
	Name string `yaml:"name"`
}
type Commands []Command

type CommandOutput struct {
	PreOutput, PostOutput, Output string
	Error                         error
	Command                       Command
	Testrun                       bool
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
		Testrun:    true,
	}
}

func (r CommandOutput) Log() {
	if len(r.PreOutput) > 0 {
		log.WithFields(log.Fields{
			"name":    r.Command.Name,
			"command": r.Command.Run,
			"success": r.Error == nil,
		}).Info(r.PreOutput)
	}

	if len(r.PostOutput) > 0 {
		log.WithFields(log.Fields{
			"name":    r.Command.Name,
			"command": r.Command.Run,
			"success": r.Error == nil,
		}).Info(r.PostOutput)
	}

	if r.Error != nil {
		log.WithFields(log.Fields{
			"name":    r.Command.Name,
			"command": r.Command.Run,
			"success": r.Error == nil,
		}).Error(r.Output + "\n error: \n" + r.Error.Error())
	} else {
		log.WithFields(log.Fields{
			"name":    r.Command.Name,
			"command": r.Command.Run,
			"success": r.Error == nil,
		}).Info(r.Output)
	}
}

func (l Commands) Start(dir string) []CommandOutput {
	var res []CommandOutput

	for _, t := range l {
		run := t.Start(dir)
		res = append(res, run)
		run.Log()
	}
	return res
}
