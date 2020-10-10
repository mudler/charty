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
	"github.com/davecgh/go-spew/spew"
	"github.com/ghodss/yaml"
	"github.com/mudler/charty/pkg/runner"
	test "github.com/mudler/charty/pkg/testchart"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	helmoptions "helm.sh/helm/v3/pkg/cli/values"
	getter "helm.sh/helm/v3/pkg/getter"
	"os"
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
	Use:     "start [CHART1] [CHART2] [flags]",
	Short:   "start a runnable helm-templated chart!",
	Aliases: []string{"run"},
	Long: `This command starts a chart.                                                                                                                                                                                                        
                                                                                                                                                                                                                                              
The start argument must be a path to a packaged chart,                                                                                                                                                                   
a path to an unpacked chart directory or a URL.                                                                                                                                                                                               
                                                                                                                                                                                                                                              
To override values in a chart, use either the '--values' flag and pass in a file                                                                                                                                                              
or use the '--set' flag and pass configuration from the command line.                                                                                                                                                                                                             
                                                                                                                                                                                                                                              
    $ charty start -f myvalues.yaml ./tests                                                                                                                                                                                           
                                                                                                                                                                                                                                              
or                                                                                                                                                                                                                                            
                                                                                                                                                                                                                                              
    $ charty start --set name=prod ./tests                                                                                                                                                                                            
                                                                                                                                                                         
                                                                                                                                                                                                                                              
You can specify the '--values'/'-f' flag multiple times. The priority will be given to the                                                                                                                                                    
last (right-most) file specified. For example, if both myvalues.yaml and override.yaml                                                                                                                                                        
contained a key called 'Test', the value set in override.yaml would take precedence:                                                                                                                                                          
                                                                                                                                                                                                                                              
    $ charty start -f myvalues.yaml -f override.yaml ./tests                                                                                                                                                                         
                                                                                                                       
You can specify the '--set' flag multiple times. The priority will be given to the                                     
last (right-most) set specified. For example, if both 'bar' and 'newbar' values are                                                                                                                                                           
set for a key called 'foo', the 'newbar' value would take precedence:                                                  
                                                                                                                       
    $ charty start --set foo=bar --set foo=newbar ./tests                                                                                                                                                                            
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("set", cmd.Flags().Lookup("set"))
		viper.BindPFlag("values", cmd.Flags().Lookup("values"))
		viper.BindPFlag("run", cmd.Flags().Lookup("run"))
		viper.BindPFlag("run-files", cmd.Flags().Lookup("run-files"))

	},
	Run: func(cmd *cobra.Command, args []string) {
		set := viper.GetStringSlice("set")
		run := viper.GetStringSlice("run")
		runFiles := viper.GetStringSlice("run-files")
		valuesFiles := viper.GetStringSlice("values")

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

			log.WithFields(log.Fields{
				"name":    testchart.Name(),
				"version": testchart.Version(),
				"chart":   a,
			}).Info(spew.Sprintf("Chart values: %v ", testchart.Values))

			log.WithFields(log.Fields{
				"name":    testchart.Name(),
				"version": testchart.Version(),
				"chart":   a,
			}).Info(spew.Sprintf("Chart runtime options: %v ", testchart.RuntimeDefaults()))

			log.Info("===========")

			out, err := testrunner.Run(testchart, startOptions)
			scripts := len(out)
			var totalTime float64

			for _, r := range out {
				if r.Testrun {
					tests++
				}

				if r.Error != nil {
					errors++
				}
				totalTime += r.Elapsed
			}

			log.Info("===========")

			if err != nil {
				log.WithFields(log.Fields{
					"errors":        errors,
					"scripts":       scripts,
					"tests":         tests,
					"total_time(s)": totalTime,
				}).Error("Error summary\n" + err.Error())
				os.Exit(1)
			} else {
				log.WithFields(log.Fields{
					"errors":        errors,
					"scripts":       scripts,
					"tests":         tests,
					"total_time(s)": totalTime,
				}).Info("Success!")
			}
		}
	},
}

func init() {
	startCmd.Flags().StringSliceP("set", "s", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	startCmd.Flags().StringSlice("run", []string{}, "set runtime values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	startCmd.Flags().StringSlice("run-files", []string{}, "specify runtimes values in a YAML file or a URL (can specify multiple)")
	startCmd.Flags().StringSliceP("values", "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")

	RootCmd.AddCommand(startCmd)
}
