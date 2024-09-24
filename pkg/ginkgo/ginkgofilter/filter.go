package ginkgofilter

import (
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
)

// FilterTestCasesBySuite filters test cases based on the suite.
func FilterTestCasesBySuite(testCases []*ginkgo.TestCase, suite string) []*ginkgo.TestCase {
	var filtered []*ginkgo.TestCase

	for _, testCase := range testCases {
		metadata := extractAnnotations(testCase.Name)
		metadata = append(metadata, testCase.Labels...)

		for _, a := range metadata {
			lca := strings.ToLower(a)

			if strings.HasPrefix(lca, "suite:") && strings.EqualFold(strings.TrimPrefix(lca, "suite:"), suite) {
				filtered = append(filtered, testCase)
			}
		}
	}

	return filtered
}

// FilterTestCasesByEnvironment filters test cases based on the environment flags. Filters apply to either
// annotations in the test name, or test labels.
func FilterTestCasesByEnvironment(testCases []*ginkgo.TestCase, envFlags sets.Set[string]) []*ginkgo.TestCase {
	var filtered []*ginkgo.TestCase

testCaseLoop:
	for _, testCase := range testCases {
		testHasInclude := false
		metadata := extractAnnotations(testCase.Name)
		metadata = append(metadata, testCase.Labels...)

		for _, a := range metadata {
			lca := strings.ToLower(a)

			if strings.HasPrefix(lca, "skipped:") {
				condition := strings.TrimPrefix(lca, "skipped:")
				if envFlags.Has(condition) {
					continue testCaseLoop // skip this test
				}
			}

			if strings.HasPrefix(strings.ToLower(a), "include:") {
				testHasInclude = true
				condition := strings.TrimPrefix(lca, "include:")
				if envFlags.Has(condition) {
					filtered = append(filtered, testCase) // include this test for sure
					break
				}
			}
		}
		if !testHasInclude {
			filtered = append(filtered, testCase)
		}
	}

	return filtered
}

func extractAnnotations(testName string) []string {
	// Define a regex to match all text within square brackets but exclude the brackets themselves
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindAllStringSubmatch(testName, -1)

	var annotations []string
	for _, match := range matches {
		if len(match) > 1 {
			annotations = append(annotations, match[1])
		}
	}
	return annotations
}
