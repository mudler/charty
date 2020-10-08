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

package cmd

import (
	"os"

	"github.com/ghodss/yaml"
	"github.com/mudler/charty/pkg/runner"
	test "github.com/mudler/charty/pkg/testchart"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	helmoptions "helm.sh/helm/v3/pkg/cli/values"
	getter "helm.sh/helm/v3/pkg/getter"
)

func mergeOptions(valuesFiles, set []string) map[string]interface{} {
	provider := getter.Provider{
		Schemes: []string{"http", "https"},
		New:     getter.NewHTTPGetter,
	}
	opts := helmoptions.Options{ValueFiles: valuesFiles, Values: set}

	res, err := opts.MergeValues(getter.Providers{provider})
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	return res
}

func runtimeOptions(merged map[string]interface{}) runner.Options {
	startOptions := runner.Options{}
	out, err := yaml.Marshal(merged)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if err = yaml.Unmarshal(out, &startOptions); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	return startOptions
}

var startCmd = &cobra.Command{
	Use:   "start <testchart> <testchart_foo>",
	Short: "start a test run",
	Long:  `start testing charts`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("set", cmd.Flags().Lookup("set"))
		viper.BindPFlag("values-files", cmd.Flags().Lookup("values-files"))
		viper.BindPFlag("run", cmd.Flags().Lookup("run"))
		viper.BindPFlag("run-files", cmd.Flags().Lookup("run-files"))

	},
	Run: func(cmd *cobra.Command, args []string) {
		set := viper.GetStringSlice("set")
		run := viper.GetStringSlice("run")
		runFiles := viper.GetStringSlice("run-files")
		valuesFiles := viper.GetStringSlice("values-files")

		testchart := &test.TestChart{Values: mergeOptions(valuesFiles, set)}
		defer testchart.Cleanup()

		startOptions := runtimeOptions(mergeOptions(runFiles, run))

		testrunner := &runner.TestRunner{}
		for _, a := range args {
			errors := 0
			tests := 0

			err := testchart.Load(a)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}

			log.WithFields(log.Fields{
				"name":    testchart.Name(),
				"version": testchart.Version(),
				"chart":   a,
			}).Info("Starting chart")

			out, err := testrunner.Run(testchart, startOptions)
			scripts := len(out)

			for _, r := range out {
				if r.Testrun {
					tests++
				}

				if r.Error != nil {
					errors++
				}
			}

			if err != nil {
				log.WithFields(log.Fields{
					"errors":  errors,
					"scripts": scripts,
					"tests":   tests,
				}).Error("Error summary\n" + err.Error())
				os.Exit(1)
			} else {
				log.WithFields(log.Fields{
					"errors":  errors,
					"scripts": scripts,
					"tests":   tests,
				}).Info("Success!")
			}
		}
	},
}

func init() {

	startCmd.Flags().Bool("clean", true, "Build all packages without considering the packages present in the build directory")
	startCmd.Flags().StringSlice("set", []string{}, "cli settings to override values")
	startCmd.Flags().StringSlice("run", []string{}, "cli settings to override values")
	startCmd.Flags().StringSlice("run-files", []string{}, "cli settings to override values")
	startCmd.Flags().StringSlice("values-files", []string{}, "cli settings to override values")

	RootCmd.AddCommand(startCmd)
}
