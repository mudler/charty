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

package runner_test

import (
	runner "github.com/mudler/charty/pkg/runner"
	test "github.com/mudler/charty/pkg/testchart"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func globstring(out []runner.CommandOutput) string {
	var str string
	for _, o := range out {
		str = str + o.Output
	}
	return str
}

var _ = Describe("Testrunner", func() {
	Context("local charts", func() {

		var testchart *test.TestChart
		var testrunner *runner.TestRunner

		BeforeEach(func() {
			testchart = &test.TestChart{Values: map[string]interface{}{"bar": "test"}}
			testrunner = &runner.TestRunner{}
		})

		AfterEach(func() {
			testchart.Cleanup()
		})

		It("executes test correctly", func() {
			err := testchart.Load("../../test/fixture")
			Expect(err).ToNot(HaveOccurred())
			out, err := testrunner.Run(testchart, runner.Options{})

			Expect(globstring(out)).To(Equal("Foo testreal\n"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("interpolates", func() {
			testchart.Values = map[string]interface{}{"foo": "foo"}
			err := testchart.Load("../../test/fixture")
			Expect(err).ToNot(HaveOccurred())
			out, err := testrunner.Run(testchart, runner.Options{})

			Expect(globstring(out)).To(Equal("Foo testfoo\n"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("catches failures and overrides chart settings", func() {
			testchart.Values = map[string]interface{}{"foo": "foo"}
			err := testchart.Load("../../test/fixture")
			Expect(err).ToNot(HaveOccurred())
			_, err = testrunner.Run(testchart, runner.Options{
				Commands: []runner.Command{{
					Name: "test",
					Run:  "bash fail.sh",
				}},
			})

			Expect(err).To(HaveOccurred())
		})
	})
})
