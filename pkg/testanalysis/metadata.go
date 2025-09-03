package testanalysis

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// NewMetadataLoader creates a new metadata loader for the given file
func NewMetadataLoader(metadataFile string) *MetadataLoader {
	return &MetadataLoader{
		metadataFile: metadataFile,
	}
}

// Load reads and parses the metadata file
func (ml *MetadataLoader) Load() error {
	data, err := os.ReadFile(ml.metadataFile)
	if err != nil {
		return fmt.Errorf("failed to read metadata file %s: %w", ml.metadataFile, err)
	}

	if err := json.Unmarshal(data, &ml.tests); err != nil {
		return fmt.Errorf("failed to parse metadata file %s: %w", ml.metadataFile, err)
	}

	return nil
}

// FindTest finds a test by exact name match
func (ml *MetadataLoader) FindTest(testName string) (*TestSpec, error) {
	for _, test := range ml.tests {
		if test.Name == testName {
			return &test, nil
		}
	}
	return nil, fmt.Errorf("test not found: %s", testName)
}

// FindTestByPartialName finds a test by partial name match (case insensitive)
func (ml *MetadataLoader) FindTestByPartialName(partialName string) ([]*TestSpec, error) {
	var matches []*TestSpec
	partialLower := strings.ToLower(partialName)

	for i, test := range ml.tests {
		if strings.Contains(strings.ToLower(test.Name), partialLower) {
			matches = append(matches, &ml.tests[i])
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no tests found matching: %s", partialName)
	}

	return matches, nil
}

// ListTests returns all available test names
func (ml *MetadataLoader) ListTests() []string {
	names := make([]string, len(ml.tests))
	for i, test := range ml.tests {
		names[i] = test.Name
	}
	return names
}

// GetTestCount returns the number of loaded tests
func (ml *MetadataLoader) GetTestCount() int {
	return len(ml.tests)
}
