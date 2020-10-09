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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
	Use:   "package [LOCAL_CHART] [DESTDIR]",
	Short: "package a runnable chart",
	Long: `This commands package a chart from a local directory to a .tar.gz compressed archive, which is stored in the destination directory given as argument.
The package archive is named after the chart metadata ("name" and "version") present in the "metadata.yaml" file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Error("Need 2 arguments, chartpath source and a destination dir")
			os.Exit(1)
		}
		testchart := &test.TestChart{}
		defer testchart.Cleanup()
		err := testchart.Package(args[0], args[1])
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		log.WithFields(log.Fields{
			"name":    testchart.Name(),
			"version": testchart.Version(),
		}).Info("Chart packaged")
	},
}

func init() {

	RootCmd.AddCommand(packageCmd)
}
