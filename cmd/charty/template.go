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

	test "github.com/mudler/charty/pkg/testchart"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template [LOCAL_CHART] [DESTDIR]",
	Short: "generate templated version of a runnable chart",
	Long:  `This command generates a templated version of the chart given in argument. The interpolation values are the default one of the chart.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Error("Need 2 arguments, chartpath source and a destination dir")
			os.Exit(1)
		}
		testchart := &test.TestChart{}
		defer testchart.Cleanup()
		err := testchart.Load(args[0])
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		if err := copy.Copy(testchart.RunnerDirectory(), args[1]); err != nil {
			log.Error(err)
			os.Exit(1)
		}

		log.WithFields(log.Fields{
			"name":    testchart.Name(),
			"version": testchart.Version(),
		}).Info("Chart generated")
	},
}

func init() {

	RootCmd.AddCommand(templateCmd)
}
