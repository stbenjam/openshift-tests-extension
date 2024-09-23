package ginkgofilter

import (
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
)

// FilterTestCases filters test cases based on the environment flags
func FilterTestCases(testCases []*ginkgo.TestCase, envFlags sets.Set[string]) []*ginkgo.TestCase {
	var filtered []*ginkgo.TestCase

testCaseLoop:
	for _, testCase := range testCases {
		testHasInclude := false
		annotations := extractAnnotations(testCase.Name)

		for _, a := range annotations {
			la := strings.ToLower(a)

			if strings.HasPrefix(strings.ToLower(a), "skipped:") {
				condition := strings.TrimPrefix(la, "skipped:")
				if envFlags.Has(condition) {
					continue testCaseLoop // skip this test
				}
			}

			if strings.HasPrefix(strings.ToLower(a), "include:") {
				testHasInclude = true
				condition := strings.TrimPrefix(la, "include:")
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
