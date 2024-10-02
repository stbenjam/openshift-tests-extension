package extensiontests

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestExtensionTestSpecs_Walk(t *testing.T) {
	specs := ExtensionTestSpecs{
		{Name: "test1"},
		{Name: "test2"},
	}

	var walkedNames []string
	specs.Walk(func(spec *ExtensionTestSpec) {
		walkedNames = append(walkedNames, spec.Name)
	})

	assert.Equal(t, []string{"test1", "test2"}, walkedNames)
}

func TestExtensionTestSpecs_MustFilter(t *testing.T) {
	specs := ExtensionTestSpecs{
		{Name: "test1"},
	}

	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, r.(string), "filter did not succeed")
		}
	}()

	// CEL expression that should fail
	specs.MustFilter([]string{"invalid_expr"})
	t.Errorf("Expected panic, but code continued")
}

func TestExtensionTestSpecs_Filter(t *testing.T) {
	tests := []struct {
		name     string
		specs    ExtensionTestSpecs
		celExprs []string
		want     ExtensionTestSpecs
		wantErr  bool
	}{
		{
			name: "simple filter on name",
			specs: ExtensionTestSpecs{
				{
					Name: "test1",
				},
				{
					Name: "test2",
				},
			},
			celExprs: []string{`name == "test1"`},
			want: ExtensionTestSpecs{
				{
					Name: "test1",
				},
			},
		},
		{
			name: "filter on tags",
			specs: ExtensionTestSpecs{
				{Name: "test1", Tags: map[string]string{"env": "prod"}},
				{Name: "test2", Tags: map[string]string{"env": "dev"}},
			},
			celExprs: []string{"tags['env'] == 'prod'"},
			want: ExtensionTestSpecs{
				{Name: "test1", Tags: map[string]string{"env": "prod"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.specs.Filter(tt.celExprs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtensionTestSpecs_AddLabel(t *testing.T) {
	specs := ExtensionTestSpecs{
		{Name: "test1", Labels: sets.New[string]()},
	}

	specs = specs.AddLabel("critical")
	assert.True(t, specs[0].Labels.Has("critical"))
}

func TestExtensionTestSpecs_RemoveLabel(t *testing.T) {
	specs := ExtensionTestSpecs{
		{Name: "test1", Labels: sets.New[string]("to_remove")},
	}
	specs = specs.RemoveLabel("to_remove")
	assert.False(t, specs[0].Labels.Has("to_remove"))
}

func TestExtensionTestSpecs_SetTag(t *testing.T) {
	specs := ExtensionTestSpecs{
		{Name: "test1", Tags: make(map[string]string)},
	}

	specs = specs.SetTag("priority", "high")
	assert.Equal(t, "high", specs[0].Tags["priority"])
}

func TestExtensionTestSpecs_UnsetTag(t *testing.T) {
	specs := ExtensionTestSpecs{
		{Name: "test1", Tags: map[string]string{"priority": "high"}},
	}

	specs = specs.UnsetTag("priority")
	_, exists := specs[0].Tags["priority"]
	assert.False(t, exists)
}
