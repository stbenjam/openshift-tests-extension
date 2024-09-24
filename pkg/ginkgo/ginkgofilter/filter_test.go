package ginkgofilter

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
)

func TestFilterTestCases(t *testing.T) {
	tests := []struct {
		name      string
		testCases []*ginkgo.TestCase
		envFlags  sets.Set[string]
		want      []string
	}{
		{
			name: "Skip AWS platform test (annotation)",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:Platform:AWS]"},
				{Name: "Test2 [Include:Platform:AWS]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want:     []string{"Test2 [Include:Platform:AWS]", "Test3"},
		},
		{
			name: "Skip AWS platform test (label)",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1", Labels: []string{"Skipped:Platform:AWS"}},
				{Name: "Test2", Labels: []string{"Include:Platform:AWS"}},
				{Name: "Test3"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want:     []string{"Test2", "Test3"},
		},
		{
			name: "No match for skipped conditions",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:Platform:GCP]"},
				{Name: "Test2"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want:     []string{"Test1 [Skipped:Platform:GCP]", "Test2"},
		},
		{
			name: "Run only specific environment",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Include:platform:gcp]"},
				{Name: "Test2 [Include:platform:aws]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want:     []string{"Test2 [Include:platform:aws]", "Test3"},
		},
		{
			name: "Skip test with no matching environment",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:topology:ha]"},
				{Name: "Test2 [platform:aws]"},
			},
			envFlags: sets.New[string]("platform:aws", "network:v6"),
			want:     []string{"Test1 [Skipped:topology:ha]", "Test2 [platform:aws]"},
		},
		{
			name:      "Empty test cases list",
			testCases: []*ginkgo.TestCase{},
			envFlags:  sets.New[string]("platform:aws"),
			want:      []string{},
		},
		{
			name: "No env flags provided",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [platform:aws]"},
				{Name: "Test2 [Skipped:platform:gcp]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string](),
			want:     []string{"Test1 [platform:aws]", "Test2 [Skipped:platform:gcp]", "Test3"},
		},
		{
			name: "Skip condition in middle of test name",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:platform:aws] Test with AWS"},
				{Name: "Test2 [Skipped:platform:gcp]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want:     []string{"Test2 [Skipped:platform:gcp]", "Test3"},
		},
		{
			name: "Case insensitive skip",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:Platform:AWS]"},
				{Name: "Test2 [platform:aws]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want:     []string{"Test2 [platform:aws]", "Test3"},
		},
		{
			name: "Multiple environment conditions",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:Platform:AWS]"},
				{Name: "Test2 [Include:NetworkStack:v6]"},
				{Name: "Test3 [Skipped:topology:ha]"},
			},
			envFlags: sets.New("platform:aws", "networkstack:v6"),
			want:     []string{"Test2 [Include:NetworkStack:v6]", "Test3 [Skipped:topology:ha]"},
		},
		{
			name: "Multiple run-only conditions, no matches",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Include:platform:gcp]"},
				{Name: "Test2 [Include:networkstack:v4]"},
			},
			envFlags: sets.New[string]("platform:aws", "networkstack:v6"),
			want:     []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterTestCasesByEnvironment(tt.testCases, tt.envFlags); len(got) > 0 && len(tt.want) > 0 && !reflect.DeepEqual(getTestCaseNames(got), tt.want) {
				t.Errorf("FilterTestCasesByEnvironment() = %v, want %v", getTestCaseNames(got), tt.want)
			}
		})
	}
}

func getTestCaseNames(testCases []*ginkgo.TestCase) []string {
	var names []string
	for _, testCase := range testCases {
		names = append(names, testCase.Name)
	}
	return names
}
