package flags

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// OutputFlags contains information for specifying multiple test names.
type OutputFlags struct {
	Output string
}

func NewOutputFlags() *OutputFlags {
	return &OutputFlags{
		Output: "json",
	}
}

func (f *OutputFlags) BindFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&f.Output,
		"output",
		"o",
		f.Output,
		"output mode")
}

func (o *OutputFlags) Marshal(v interface{}) ([]byte, error) {
	switch o.Output {
	case "", "json":
		j, err := json.MarshalIndent(&v, "", "  ")
		if err != nil {
			return nil, err
		}
		return j, nil
	case "jsonl":
		// Check if v is a slice or array
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			var result []byte
			for i := 0; i < val.Len(); i++ {
				item := val.Index(i).Interface()
				j, err := json.Marshal(item)
				if err != nil {
					return nil, err
				}
				result = append(result, j...)
				result = append(result, '\n') // Append newline after each item
			}
			return result, nil
		}
		return nil, errors.New("jsonl format requires a slice or array")
	default:
		return nil, errors.Errorf("invalid output format: %s", o.Output)
	}
}
