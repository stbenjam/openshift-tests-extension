package ginkgo

import (
	"bytes"
	"testing"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "clean JSON object",
			input: `{"name": "test", "result": "passed"}`,
			want:  `{"name": "test", "result": "passed"}`,
		},
		{
			name:  "clean JSON array",
			input: `[{"name": "test", "result": "passed"}]`,
			want:  `[{"name": "test", "result": "passed"}]`,
		},
		{
			name:  "klog INFO lines before JSON object",
			input: "I0625 19:25:00.000000   12345 client.go:123] Connecting to server\nI0625 19:25:01.000000   12345 client.go:456] Connected\n{\"name\": \"test\", \"result\": \"passed\"}\n",
			want:  `{"name": "test", "result": "passed"}`,
		},
		{
			name:  "klog WARNING lines before JSON array",
			input: "W0625 19:25:00.000000   12345 controller.go:47] some warning\n[{\"name\": \"test\", \"result\": \"passed\"}]\n",
			want:  `[{"name": "test", "result": "passed"}]`,
		},
		{
			name:  "Ginkgo reporter output before JSON",
			input: "FAIL!\nRan 1 of 1 Specs in 5.123 seconds\n1 Passed | 0 Failed | 0 Pending | 0 Skipped\n{\"name\": \"test\", \"result\": \"failed\"}\n",
			want:  `{"name": "test", "result": "failed"}`,
		},
		{
			name:  "trailing non-JSON after JSON payload",
			input: "{\"name\": \"test\", \"result\": \"passed\"}\nFAIL!\nRan 1 of 1 Specs\n",
			want:  `{"name": "test", "result": "passed"}`,
		},
		{
			name:  "klog lines before AND Ginkgo summary after JSON",
			input: "I0625 19:25:00.000000   12345 client.go:123] Starting\n[{\"name\": \"test\", \"result\": \"passed\"}]\nRan 1 of 1 Specs in 2.000 seconds\nSUCCESS!\n",
			want:  `[{"name": "test", "result": "passed"}]`,
		},
		{
			name:    "no JSON in output - only log lines",
			input:   "I0625 19:25:00.000000   12345 client.go:123] Connecting\nW0625 some warning\n",
			wantErr: true,
		},
		{
			name:    "empty output",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only whitespace and newlines",
			input:   "  \n\n  \n",
			wantErr: true,
		},
		{
			name:  "JSON with leading whitespace on the line",
			input: "I0625 log line\n  {\"name\": \"test\", \"result\": \"passed\"}\n",
			want:  `{"name": "test", "result": "passed"}`,
		},
		{
			name:  "multiline pretty-printed JSON object",
			input: "I0625 log line\n{\n  \"name\": \"test\",\n  \"result\": \"passed\"\n}\n",
			want:  "{\n  \"name\": \"test\",\n  \"result\": \"passed\"\n}",
		},
		{
			name:  "JSON with trailing whitespace and CRLF",
			input: "{\"name\": \"test\", \"result\": \"passed\"}\r\n",
			want:  `{"name": "test", "result": "passed"}`,
		},
		{
			name:  "non-JSON bracket lines before valid JSON array",
			input: "[FAILED]\n[BeforeEach]\n[{\"name\": \"test\", \"result\": \"passed\"}]\n",
			want:  `[{"name": "test", "result": "passed"}]`,
		},
		{
			name:  "non-JSON brace lines before valid JSON object",
			input: "{not-json}\n{also not json\n{\"name\": \"test\", \"result\": \"passed\"}\n",
			want:  `{"name": "test", "result": "passed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("extractJSON() expected error, got nil with result: %s", string(got))
				}
				return
			}
			if err != nil {
				t.Fatalf("extractJSON() unexpected error: %v", err)
			}

			// Compare as raw JSON strings (byte-level, since we're testing extraction boundaries)
			if string(got) != tt.want {
				t.Errorf("extractJSON() =\n  %s\nwant:\n  %s", string(got), tt.want)
			}
		})
	}
}

func TestNewTestResultFromOutput(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantName   string
		wantResult string
		wantErr    bool
	}{
		{
			name:       "clean JSON object",
			input:      `{"name": "my-test", "result": "passed", "duration": 1000}`,
			wantName:   "my-test",
			wantResult: "passed",
		},
		{
			name:       "clean JSON array with one element",
			input:      `[{"name": "my-test", "result": "failed", "duration": 2000}]`,
			wantName:   "my-test",
			wantResult: "failed",
		},
		{
			name:       "klog lines before JSON object",
			input:      "I0625 19:25:00.000000   12345 client.go:123] Connecting\n{\"name\": \"my-test\", \"result\": \"passed\", \"duration\": 500}\n",
			wantName:   "my-test",
			wantResult: "passed",
		},
		{
			name:       "klog lines before JSON array",
			input:      "I0625 19:25:00.000000   12345 setup.go:99] Initializing\nW0625 19:25:01.000000   12345 warn.go:10] Deprecation\n[{\"name\": \"my-test\", \"result\": \"passed\", \"duration\": 100}]\n",
			wantName:   "my-test",
			wantResult: "passed",
		},
		{
			name:       "Ginkgo reporter output surrounding JSON",
			input:      "FAIL!\nRan 1 of 1 Specs in 5.123 seconds\n{\"name\": \"my-test\", \"result\": \"failed\", \"duration\": 5123}\nSUCCESS! -- 1 Passed | 0 Failed\n",
			wantName:   "my-test",
			wantResult: "failed",
		},
		{
			name:    "empty output",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no JSON at all",
			input:   "I0625 just log lines\nno json here\n",
			wantErr: true,
		},
		{
			name:    "JSON array with zero elements",
			input:   `[]`,
			wantErr: true,
		},
		{
			name:    "JSON array with two elements",
			input:   `[{"name": "a", "result": "passed"},{"name": "b", "result": "failed"}]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBufferString(tt.input)
			result, err := newTestResultFromOutput(buf)
			if tt.wantErr {
				if err == nil {
					t.Errorf("newTestResultFromOutput() expected error, got result: %+v", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("newTestResultFromOutput() unexpected error: %v", err)
			}
			if result.Name != tt.wantName {
				t.Errorf("result.Name = %q, want %q", result.Name, tt.wantName)
			}
			if string(result.Result) != tt.wantResult {
				t.Errorf("result.Result = %q, want %q", result.Result, tt.wantResult)
			}
		})
	}
}
