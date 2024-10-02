package extensiontests

import (
	"encoding/json"
	"fmt"
	"io"
)

type ResultFormat string

var (
	JSON  ResultFormat = "json"
	JSONL ResultFormat = "jsonl"
)

type ResultWriter struct {
	out     io.Writer
	format  ResultFormat
	results ExtensionTestResults
}

func NewResultWriter(out io.Writer, format ResultFormat) (*ResultWriter, error) {
	switch format {
	case JSON, JSONL:
	// do nothing
	default:
		return nil, fmt.Errorf("unsupported result format: %s", format)
	}

	return &ResultWriter{
		out:    out,
		format: format,
	}, nil
}

func (w *ResultWriter) Write(result *ExtensionTestResult) {
	switch w.format {
	case JSONL:
		// JSONL gets written to out as we get the items
		data, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w.out, "%s\n", string(data))
	case JSON:
		w.results = append(w.results, result)
	}
}

func (w *ResultWriter) Flush() {
	switch w.format {
	case JSONL:
	// we already wrote it out
	case JSON:
		data, err := json.MarshalIndent(w.results, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w.out, "%s\n", string(data))
	}
}
