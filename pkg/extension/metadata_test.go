package extension

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"

	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

func TestExtensionMetadata_FindRemovedTestsWithoutRename(t *testing.T) {
	tests := []struct {
		name          string
		old           et.ExtensionTestSpecs
		new           et.ExtensionTestSpecs
		obsoleteTests []string
		want          []string
		wantErr       bool
	}{
		{
			name: "allows a test to be renamed",
			old: et.ExtensionTestSpecs{
				{
					Name: "this test has a tpyo",
				},
			},
			new: et.ExtensionTestSpecs{
				{
					Name:       "this test doesn't have a typo",
					OtherNames: sets.New[string]("this test has a tpyo"),
				},
			},
			wantErr: false,
		},
		{
			name: "fails when a test is removed",
			old: et.ExtensionTestSpecs{
				{
					Name: "this test was deleted",
				},
			},
			new:     et.ExtensionTestSpecs{},
			want:    []string{"this test was deleted"},
			wantErr: true,
		},
		{
			name: "succeeds when a test is removed and it's marked obsolete",
			old: et.ExtensionTestSpecs{
				{
					Name: "this test was deleted",
				},
			},
			new:           et.ExtensionTestSpecs{},
			obsoleteTests: []string{"this test was deleted"},
			wantErr:       false,
		},
		{
			name: "fails when a test is renamed without other names",
			old: et.ExtensionTestSpecs{
				{
					Name: "this test has a tpyo",
				},
			},
			new: et.ExtensionTestSpecs{
				{
					Name: "this test doesn't have a typo",
				},
			},
			want:    []string{"this test has a tpyo"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := NewExtension("openshift", "testing", "dummy")
			ext.AddSpecs(tt.new)
			ext.IgnoreObsoleteTests(tt.obsoleteTests...)

			m := ExtensionMetadata{
				extension: ext,
				specs:     tt.old,
			}

			got, err := m.FindRemovedTestsWithoutRename()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindRemovedTestsWithoutRename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindRemovedTestsWithoutRename() got = %v, want %v", got, tt.want)
			}
		})
	}
}
