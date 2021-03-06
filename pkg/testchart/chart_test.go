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
package chart_test

import (
	"io/ioutil"
	"path/filepath"

	test "github.com/mudler/charty/pkg/testchart"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chart", func() {
	Context("Loading", func() {

		var testchart *test.TestChart

		BeforeEach(func() {
			testchart = &test.TestChart{Values: map[string]interface{}{"bar": "test"}}
		})

		AfterEach(func() {
			testchart.Cleanup()
		})

		It("renders templates", func() {
			err := testchart.Load("../../test/fixture")
			Expect(err).ToNot(HaveOccurred())

			dat, err := ioutil.ReadFile(filepath.Join(testchart.RunnerDirectory(), "test.sh"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(dat)).To(Equal(`echo "Foo testreal"`))
		})

		It("overrides defaults", func() {
			testchart.Values = map[string]interface{}{"foo": "foo"}
			err := testchart.Load("../../test/fixture")
			Expect(err).ToNot(HaveOccurred())

			dat, err := ioutil.ReadFile(filepath.Join(testchart.RunnerDirectory(), "test.sh"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(dat)).To(Equal(`echo "Foo testfoo"`))
		})
	})
})
