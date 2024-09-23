package ginkgofilter

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
)

func TestFilterTestCases(t *testing.T) {
	tests := []struct {
		name      string
		testCases []*ginkgo.TestCase
		envFlags  sets.Set[string]
		want      []*ginkgo.TestCase
	}{
		{
			name: "Skip AWS platform test",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:Platform:AWS]"},
				{Name: "Test2 [Include:Platform:AWS]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want: []*ginkgo.TestCase{
				{Name: "Test2 [Include:Platform:AWS]"},
				{Name: "Test3"},
			},
		},
		{
			name: "No match for skipped conditions",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:Platform:GCP]"},
				{Name: "Test2"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:Platform:GCP]"},
				{Name: "Test2"},
			},
		},
		{
			name: "Run only specific environment",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Include:platform:gcp]"},
				{Name: "Test2 [Include:platform:aws]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want: []*ginkgo.TestCase{
				{Name: "Test2 [Include:platform:aws]"},
				{Name: "Test3"},
			},
		},
		{
			name: "Skip test with no matching environment",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:topology:ha]"},
				{Name: "Test2 [platform:aws]"},
			},
			envFlags: sets.New[string]("platform:aws", "network:v6"),
			want: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:topology:ha]"},
				{Name: "Test2 [platform:aws]"},
			},
		},
		{
			name:      "Empty test cases list",
			testCases: []*ginkgo.TestCase{},
			envFlags:  sets.New[string]("platform:aws"),
			want:      []*ginkgo.TestCase{},
		},
		{
			name: "No env flags provided",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [platform:aws]"},
				{Name: "Test2 [Skipped:platform:gcp]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string](),
			want: []*ginkgo.TestCase{
				{Name: "Test1 [platform:aws]"},
				{Name: "Test2 [Skipped:platform:gcp]"},
				{Name: "Test3"},
			},
		},
		{
			name: "Skip condition in middle of test name",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:platform:aws] Test with AWS"},
				{Name: "Test2 [Skipped:platform:gcp]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want: []*ginkgo.TestCase{
				{Name: "Test2 [Skipped:platform:gcp]"},
				{Name: "Test3"},
			},
		},
		{
			name: "Case insensitive skip",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:Platform:AWS]"},
				{Name: "Test2 [platform:aws]"},
				{Name: "Test3"},
			},
			envFlags: sets.New[string]("platform:aws"),
			want: []*ginkgo.TestCase{
				{Name: "Test2 [platform:aws]"},
				{Name: "Test3"},
			},
		},
		{
			name: "Multiple environment conditions",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Skipped:Platform:AWS]"},
				{Name: "Test2 [Include:NetworkStack:v6]"},
				{Name: "Test3 [Skipped:topology:ha]"},
			},
			envFlags: sets.New("platform:aws", "networkstack:v6"),
			want: []*ginkgo.TestCase{
				{Name: "Test2 [Include:NetworkStack:v6]"},
				{Name: "Test3 [Skipped:topology:ha]"},
			},
		},
		{
			name: "Multiple run-only conditions, no matches",
			testCases: []*ginkgo.TestCase{
				{Name: "Test1 [Include:platform:gcp]"},
				{Name: "Test2 [Include:networkstack:v4]"},
			},
			envFlags: sets.New[string]("platform:aws", "networkstack:v6"),
			want:     []*ginkgo.TestCase{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterTestCases(tt.testCases, tt.envFlags); !reflect.DeepEqual(got, tt.want) {
				if len(tt.want) == 0 && len(got) == 0 {
					return
				}

				var gotSlice, wantSlice []string
				for _, g := range got {
					gotSlice = append(gotSlice, g.Name)
				}
				for _, w := range tt.want {
					wantSlice = append(wantSlice, w.Name)
				}

				t.Errorf("FilterTestCases() = %v, wanted %v", strings.Join(gotSlice, ", "), strings.Join(wantSlice, ", "))
			}
		})
	}
}
