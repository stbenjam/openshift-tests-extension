package cmdanalyze

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/testanalysis"
)

type AnalyzeOptions struct {
	TestName     string
	MetadataFile string
	OutputFormat string
	Verbose      bool
}

func NewAnalyzeCommand(registry *extension.Registry) *cobra.Command {
	o := &AnalyzeOptions{}

	cmd := &cobra.Command{
		Use:   "analyze-test <test-name>",
		Short: "Extract complete source code for a specific test",
		Long: `Analyze a test and extract all source code that makes up the test,
including Ginkgo blocks, setup/teardown hooks, and helper functions.

The test name should be the full test name as it appears in the metadata.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TestName = args[0]
			return o.Run()
		},
	}

	cmd.Flags().StringVar(&o.MetadataFile, "metadata", "", "Path to test metadata JSON file (auto-detected if not specified)")
	cmd.Flags().StringVar(&o.OutputFormat, "output", "json", "Output format: json, summary")
	cmd.Flags().BoolVarP(&o.Verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func (o *AnalyzeOptions) Run() error {
	// Auto-detect metadata file if not specified
	if o.MetadataFile == "" {
		var err error
		o.MetadataFile, err = o.findMetadataFile()
		if err != nil {
			return fmt.Errorf("failed to find metadata file: %w", err)
		}
	}

	if o.Verbose {
		fmt.Fprintf(os.Stderr, "Using metadata file: %s\n", o.MetadataFile)
		fmt.Fprintf(os.Stderr, "Analyzing test: %s\n", o.TestName)
	}

	// Load metadata
	loader := testanalysis.NewMetadataLoader(o.MetadataFile)
	if err := loader.Load(); err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	if o.Verbose {
		fmt.Fprintf(os.Stderr, "Loaded %d tests from metadata\n", loader.GetTestCount())
	}

	// Find the specific test
	testSpec, err := loader.FindTest(o.TestName)
	if err != nil {
		// Try partial name match
		matches, partialErr := loader.FindTestByPartialName(o.TestName)
		if partialErr != nil {
			return fmt.Errorf("test not found: %w", err)
		}

		if len(matches) > 1 {
			fmt.Fprintf(os.Stderr, "Multiple tests match '%s':\n", o.TestName)
			for _, match := range matches {
				fmt.Fprintf(os.Stderr, "  - %s\n", match.Name)
			}
			return fmt.Errorf("please specify a more specific test name")
		}

		testSpec = matches[0]
		if o.Verbose {
			fmt.Fprintf(os.Stderr, "Found test by partial match: %s\n", testSpec.Name)
		}
	}

	// Extract source code
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	extractor := testanalysis.NewSourceExtractor(workingDir)
	analysis, err := extractor.AnalyzeTest(testSpec)
	if err != nil {
		return fmt.Errorf("failed to analyze test: %w", err)
	}

	// Output results
	switch o.OutputFormat {
	case "json":
		return o.outputJSON(analysis)
	case "summary":
		return o.outputSummary(analysis)
	default:
		return fmt.Errorf("unknown output format: %s", o.OutputFormat)
	}
}

func (o *AnalyzeOptions) findMetadataFile() (string, error) {
	// Look for metadata files in .openshift-tests-extension/
	metadataDir := ".openshift-tests-extension"

	if _, err := os.Stat(metadataDir); os.IsNotExist(err) {
		return "", fmt.Errorf("metadata directory not found: %s", metadataDir)
	}

	// Look for JSON files in the metadata directory
	entries, err := os.ReadDir(metadataDir)
	if err != nil {
		return "", fmt.Errorf("failed to read metadata directory: %w", err)
	}

	var jsonFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			jsonFiles = append(jsonFiles, filepath.Join(metadataDir, entry.Name()))
		}
	}

	if len(jsonFiles) == 0 {
		return "", fmt.Errorf("no JSON metadata files found in %s", metadataDir)
	}

	if len(jsonFiles) == 1 {
		return jsonFiles[0], nil
	}

	// Multiple files found - prefer example-tests
	for _, file := range jsonFiles {
		if filepath.Base(file) == "openshift_payload_example-tests.json" {
			return file, nil
		}
	}

	// Return the first one as fallback
	return jsonFiles[0], nil
}

func (o *AnalyzeOptions) outputJSON(analysis *testanalysis.TestSourceAnalysis) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(analysis)
}

func (o *AnalyzeOptions) outputSummary(analysis *testanalysis.TestSourceAnalysis) error {
	fmt.Printf("Test Analysis: %s\n", analysis.TestName)
	fmt.Printf("=====================================\n\n")

	fmt.Printf("Ginkgo Blocks (%d):\n", len(analysis.GinkgoBlocks))
	for i, block := range analysis.GinkgoBlocks {
		fmt.Printf("  %d. %s (%s:%d)\n", i+1, block.Description, block.File, block.StartLine)
		if o.Verbose {
			fmt.Printf("     Type: %s\n", block.Type)
		}
	}

	if len(analysis.HelperFunctions) > 0 {
		fmt.Printf("\nHelper Functions (%d):\n", len(analysis.HelperFunctions))
		for i, helper := range analysis.HelperFunctions {
			fmt.Printf("  %d. %s (%s:%d-%d)\n", i+1, helper.Description, helper.File, helper.StartLine, helper.EndLine)
		}
	}

	return nil
}
