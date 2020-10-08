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
	"fmt"
	"os"

	"github.com/ghodss/yaml"

	"github.com/mudler/charty/pkg/runner"
	test "github.com/mudler/charty/pkg/testchart"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	helmoptions "helm.sh/helm/v3/pkg/cli/values"
	getter "helm.sh/helm/v3/pkg/getter"
)

var startCmd = &cobra.Command{
	Use:   "start <package name> <package name> <package name> ...",
	Short: "start a package or a tree",
	Long:  `start packages or trees from luet tree definitions. Packages are in [category]/[name]-[version] form`,
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

		provider := getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		}

		opts := helmoptions.Options{ValueFiles: valuesFiles, Values: set}

		res, err := opts.MergeValues(getter.Providers{provider})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		testchart := &test.TestChart{Values: res}
		defer testchart.Cleanup()

		runnerOpts := helmoptions.Options{ValueFiles: runFiles, Values: run}

		res, err = runnerOpts.MergeValues(getter.Providers{provider})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		startOptions := runner.Options{}
		out, err := yaml.Marshal(res)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err = yaml.Unmarshal(out, &startOptions); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		testrunner := &runner.TestRunner{}
		for _, a := range args {
			err := testchart.Load(a)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			out, err := testrunner.Run(testchart, startOptions)
			for _, r := range out {
				fmt.Println("===========")
				fmt.Println(fmt.Sprintf("[name: %s] [command: %s]", r.Command.Name, r.Command.Run))
				if len(r.PreOutput) > 0 {
					fmt.Println(fmt.Sprintf("  [pre-script]: %s", r.PreOutput))
				}
				fmt.Println(fmt.Sprintf("  output: %s", r.Output))
				if len(r.PostOutput) > 0 {
					fmt.Println(fmt.Sprintf("  [post-script]: %s", r.PostOutput))
				}
				if r.Error != nil {
					fmt.Println(fmt.Sprintf("  error: %s", r.Error))
				} else {
					fmt.Println("  -> OK")
				}
			}
			if err != nil {
				fmt.Println("===========")
				fmt.Println("Execution failed, bailing out.")
				fmt.Println(err)
				os.Exit(1)
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
