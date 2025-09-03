package testanalysis

// TestSpec represents a test specification loaded from metadata
type TestSpec struct {
	Name                string         `json:"name"`
	Labels              map[string]any `json:"labels"`
	Resources           map[string]any `json:"resources"`
	Source              string         `json:"source"`
	CodeLocations       []string       `json:"codeLocations"`
	Lifecycle           string         `json:"lifecycle"`
	EnvironmentSelector map[string]any `json:"environmentSelector"`
}

// CodeBlock represents an extracted block of source code
type CodeBlock struct {
	Type        string // "describe", "it", "beforeEach", etc.
	Source      string // The actual source code
	File        string // Source file path
	StartLine   int    // Starting line number
	EndLine     int    // Ending line number
	Description string // Human readable description
	IsFiltered  bool   // True if this is a synthetic/filtered version of original code
}

// TestSourceAnalysis contains all the extracted source code for a test
type TestSourceAnalysis struct {
	TestName        string      `json:"testName"`
	GinkgoBlocks    []CodeBlock `json:"ginkgoBlocks"`
	HelperFunctions []CodeBlock `json:"helperFunctions"`
}

// MetadataLoader handles loading and searching test metadata
type MetadataLoader struct {
	metadataFile string
	tests        []TestSpec
}

// SourceExtractor handles extracting source code from Go files
type SourceExtractor struct {
	workingDir string
}
