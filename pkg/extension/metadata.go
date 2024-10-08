package extension

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/openshift-eng/openshift-tests-extension/pkg/dbtime"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

type ExtensionMetadata struct {
	path      string
	extension *Extension
	specs     extensiontests.ExtensionTestSpecs
}

func (e *Extension) NewMetadataFromDisk(dir string) (*ExtensionMetadata, error) {
	metadata := &ExtensionMetadata{
		extension: e,
		path: filepath.Join(dir,
			fmt.Sprintf("%s.specs.json",
				strings.ReplaceAll(e.Component.Identifier(), ":", "_"))),
		specs: extensiontests.ExtensionTestSpecs{},
	}

	// Create the metadata directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, errors.Wrapf(err, "failed to create metadata directory %s", dir)
	}

	// Read existing specs

	source, err := os.Open(metadata.path)
	if err != nil {
		return metadata, err
	}
	if err := json.NewDecoder(source).Decode(&metadata.specs); err != nil {
		return nil, errors.Wrapf(err, "failed to decode file: %s", metadata.path)
	}

	return metadata, nil
}

func (m *ExtensionMetadata) GetCreationDate(name string) *dbtime.DBTime {
	spec, err := m.specs.FindSpecByAllNames(name)
	if err == nil && spec != nil && spec.CreatedAt != nil {
		return spec.CreatedAt
	}

	// If we don't have a historical creation date, it's now
	return dbtime.Ptr(time.Now())
}

func (m *ExtensionMetadata) FindRemovedTestsWithoutRename() ([]string,
	error) {
	currentSpecs := m.extension.specs
	currentNames := currentSpecs.Names()
	currentOtherNames := currentSpecs.OtherNames()

	var potentiallyMissing extensiontests.ExtensionTestSpecs
	for _, oldSpec := range m.specs {
		// Check if the test's name is missing from current specs
		found := false
		for _, name := range currentNames {
			if oldSpec.Name == name {
				found = true
				break
			}
		}
		if !found {
			potentiallyMissing = append(potentiallyMissing, oldSpec)
		}
	}

	var actuallyMissing extensiontests.ExtensionTestSpecs
	for _, spec := range potentiallyMissing {
		found := false
		for _, otherName := range currentOtherNames {
			if spec.Name == otherName {
				found = true
				break
			}
		}
		if !found {
			actuallyMissing = append(actuallyMissing, spec)
		}
	}

	// Filter out permitted obsolete tests
	var unpermittedMissingTests []string
	for _, spec := range actuallyMissing {
		missing := true
		for _, allowed := range m.extension.obsoleteTests {
			if spec.Name == allowed {
				missing = false
				break
			}
		}
		if missing {
			unpermittedMissingTests = append(unpermittedMissingTests, spec.Name)
		}
	}

	if len(unpermittedMissingTests) > 0 {
		return unpermittedMissingTests, fmt.Errorf("%d tests were not found", len(unpermittedMissingTests))
	}

	return nil, nil
}

func (m *ExtensionMetadata) WriteToDisk() error {
	// no missing tests, write the results
	curSpecs := m.extension.GetSpecs()

	// set creation time
	curSpecs.Walk(func(spec *extensiontests.ExtensionTestSpec) {
		spec.CreatedAt = m.GetCreationDate(spec.Name)
	})

	data, err := json.Marshal(curSpecs)
	if err != nil {
		return errors.Wrap(err, "failed to marshal specs to JSON")
	}

	// Write the JSON data to the file
	if err := os.WriteFile(m.path, data, 0644); err != nil {
		return errors.Wrapf(err, "failed to write file %s", m.path)
	}

	return nil
}
