package cmddiff

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
)

// DiffFlags contains all flags for the diff command
type DiffFlags struct {
	File            string
	Base            string
	RenameThreshold float64
	Format          string
	Only            []string
	Strict          bool
	Verbose         bool
}

func NewDiffFlags() *DiffFlags {
	return &DiffFlags{
		RenameThreshold: 0.9,
		Format:          "json",
	}
}

func (f *DiffFlags) BindFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&f.File, "file", "f", f.File, "path to the metadata JSON file in the repo (required)")
	fs.StringVarP(&f.Base, "base", "b", f.Base, "SHA or branch/tag used to compute the merge base (required)")
	fs.Float64Var(&f.RenameThreshold, "rename-threshold", f.RenameThreshold, "Jaccard threshold on codeLocations for rename detection")
	fs.StringVar(&f.Format, "format", f.Format, "output format: json|table|markdown")
	fs.StringSliceVar(&f.Only, "only", f.Only, "filter output to only show: added|removed|modified|renamed (repeatable)")
	fs.BoolVar(&f.Strict, "strict", f.Strict, "exit non-zero if differences found")
	fs.BoolVar(&f.Verbose, "verbose", f.Verbose, "verbose logging")
}

// TestMetadata represents a single test entry in the JSON file
type TestMetadata struct {
	Name                string                 `json:"name"`
	Labels              map[string]interface{} `json:"labels"`
	Resources           map[string]interface{} `json:"resources"`
	Source              string                 `json:"source"`
	CodeLocations       []string               `json:"codeLocations"`
	Lifecycle           string                 `json:"lifecycle"`
	EnvironmentSelector map[string]interface{} `json:"environmentSelector"`
}

// DiffResult represents the result of comparing two test metadata sets
type DiffResult struct {
	New                 []TestMetadata `json:"new,omitempty"`
	Removed             []TestMetadata `json:"removed,omitempty"`
	PotentiallyModified []ModifiedTest `json:"potentiallyModified,omitempty"`
	Modified            []ModifiedTest `json:"modified,omitempty"` // AST-confirmed modifications
	Renamed             []RenamedTest  `json:"renamed,omitempty"`
}

// ModifiedTest represents a test that was potentially changed
type ModifiedTest struct {
	Name     string       `json:"name"`
	BaseTest TestMetadata `json:"base"`
	HeadTest TestMetadata `json:"head"`
}

// RenamedTest represents a test that was renamed
type RenamedTest struct {
	OldName    string       `json:"oldName"`
	NewName    string       `json:"newName"`
	BaseTest   TestMetadata `json:"base"`
	HeadTest   TestMetadata `json:"head"`
	Similarity float64      `json:"similarity"`
}

// NewDiffCommand creates the diff command
func NewDiffCommand(registry *extension.Registry) *cobra.Command {
	flags := NewDiffFlags()

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Detect added, removed, modified, and renamed tests by comparing metadata files",
		Long: `Detect added, removed, modified, and optional renamed tests by comparing a JSON metadata 
file in the current worktree against the version at the merge base with a given ref.

Uses AST-based analysis to precisely identify actual source code changes, ignoring line number
shifts and focusing only on real content modifications.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(flags)
		},
	}

	flags.BindFlags(cmd.Flags())
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("base")

	return cmd
}

func runDiff(flags *DiffFlags) error {
	// Validate flags
	if err := validateFlags(flags); err != nil {
		return err
	}

	// Step 1: Resolve merge base
	mergeBase, err := getMergeBase(flags.Base)
	if err != nil {
		return fmt.Errorf("failed to get merge base: %w", err)
	}

	if flags.Verbose {
		fmt.Fprintf(os.Stderr, "Merge base: %s\n", mergeBase)
	}

	// Step 2: Read files
	baseTests, err := readTestsFromGit(mergeBase, flags.File, flags.Verbose)
	if err != nil {
		return fmt.Errorf("failed to read base file: %w", err)
	}

	headTests, err := readTestsFromWorkTree(flags.File, flags.Verbose)
	if err != nil {
		return fmt.Errorf("failed to read head file: %w", err)
	}

	if flags.Verbose {
		fmt.Fprintf(os.Stderr, "Base tests: %d, Head tests: %d\n", len(baseTests), len(headTests))
	}

	// Step 3: Perform diff
	result, err := performDiff(baseTests, headTests, mergeBase, flags)
	if err != nil {
		return fmt.Errorf("failed to perform diff: %w", err)
	}

	// Step 3a: Enhanced AST analysis for precise change detection
	if len(result.PotentiallyModified) > 0 {
		if flags.Verbose {
			fmt.Fprintf(os.Stderr, "Performing AST analysis on %d potentially modified tests...\n", len(result.PotentiallyModified))
		}

		enhancedResult, err := performASTAnalysis(result, mergeBase, flags)
		if err != nil {
			if flags.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: AST analysis failed: %v\n", err)
			}
		} else {
			result = enhancedResult
		}
	}

	// Step 4: Filter results
	result = filterResults(result, flags.Only)

	// Step 5: Output results
	if err := outputResults(result, flags.Format); err != nil {
		return fmt.Errorf("failed to output results: %w", err)
	}

	// Step 6: Handle exit codes
	if flags.Strict && hasDifferences(result) {
		os.Exit(2)
	}

	return nil
}

func validateFlags(flags *DiffFlags) error {
	if flags.File == "" {
		return fmt.Errorf("--file flag is required")
	}
	if flags.Base == "" {
		return fmt.Errorf("--base flag is required")
	}
	if flags.RenameThreshold < 0 || flags.RenameThreshold > 1 {
		return fmt.Errorf("--rename-threshold must be between 0 and 1")
	}
	validFormats := map[string]bool{"json": true, "table": true, "markdown": true}
	if !validFormats[flags.Format] {
		return fmt.Errorf("--format must be one of: json, table, markdown")
	}
	validOnlyTypes := map[string]bool{"added": true, "new": true, "removed": true, "modified": true, "potentiallyModified": true, "renamed": true}
	for _, only := range flags.Only {
		if !validOnlyTypes[only] {
			return fmt.Errorf("--only must be one of: added, removed, modified, renamed")
		}
	}
	return nil
}

func getMergeBase(base string) (string, error) {
	cmd := exec.Command("git", "merge-base", "HEAD", base)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func readTestsFromGit(ref, filePath string, verbose bool) ([]TestMetadata, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", ref, filePath))
	output, err := cmd.Output()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "File %s not found at ref %s, treating as empty\n", filePath, ref)
		}
		return []TestMetadata{}, nil
	}

	return parseTests(output)
}

func readTestsFromWorkTree(filePath string, verbose bool) ([]TestMetadata, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if verbose {
			fmt.Fprintf(os.Stderr, "File %s not found in worktree, treating as empty\n", filePath)
		}
		return []TestMetadata{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return parseTests(data)
}

func parseTests(data []byte) ([]TestMetadata, error) {
	var tests []TestMetadata
	if len(data) == 0 {
		return tests, nil
	}

	if err := json.Unmarshal(data, &tests); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return tests, nil
}

func performDiff(baseTests, headTests []TestMetadata, mergeBase string, flags *DiffFlags) (*DiffResult, error) {
	baseMap := make(map[string]TestMetadata)
	headMap := make(map[string]TestMetadata)

	// Create maps for quick lookup
	for _, test := range baseTests {
		baseMap[test.Name] = test
	}
	for _, test := range headTests {
		headMap[test.Name] = test
	}

	result := &DiffResult{}

	// Use the provided merge base for content comparison

	// Get list of modified files to determine potentially modified tests
	modifiedFiles, err := getModifiedFiles(mergeBase)
	if err != nil && flags.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: Could not get modified files: %v\n", err)
	}

	// Find added, removed, and potentially modified tests
	for name, headTest := range headMap {
		if baseTest, exists := baseMap[name]; exists {
			// Test exists in both - check if potentially modified
			if isPotentiallyModified(headTest, modifiedFiles) {
				modified := ModifiedTest{
					Name:     name,
					BaseTest: baseTest,
					HeadTest: headTest,
				}
				result.PotentiallyModified = append(result.PotentiallyModified, modified)
			}
		} else {
			// Test only in head - potentially added or renamed
			result.New = append(result.New, headTest)
		}
	}

	for name, baseTest := range baseMap {
		if _, exists := headMap[name]; !exists {
			// Test only in base - potentially removed or renamed
			result.Removed = append(result.Removed, baseTest)
		}
	}

	// Detect renames using exact code location matching
	renames := detectExactRenames(result.Removed, result.New)
	result.Renamed = renames

	// Remove renamed tests from new/removed
	result.New = filterOutRenamed(result.New, renames, false)
	result.Removed = filterOutRenamed(result.Removed, renames, true)

	return result, nil
}

func filterOutRenamed(tests []TestMetadata, renames []RenamedTest, isRemoved bool) []TestMetadata {
	renamedNames := make(map[string]bool)

	for _, rename := range renames {
		if isRemoved {
			renamedNames[rename.OldName] = true
		} else {
			renamedNames[rename.NewName] = true
		}
	}

	var filtered []TestMetadata
	for _, test := range tests {
		if !renamedNames[test.Name] {
			filtered = append(filtered, test)
		}
	}

	return filtered
}

func filterResults(result *DiffResult, only []string) *DiffResult {
	if len(only) == 0 {
		return result
	}

	onlyMap := make(map[string]bool)
	for _, o := range only {
		onlyMap[o] = true
	}

	filtered := &DiffResult{}

	if onlyMap["added"] || onlyMap["new"] {
		filtered.New = result.New
	}
	if onlyMap["removed"] {
		filtered.Removed = result.Removed
	}
	if onlyMap["modified"] || onlyMap["potentiallyModified"] {
		filtered.PotentiallyModified = result.PotentiallyModified
		filtered.Modified = result.Modified
	}
	if onlyMap["renamed"] {
		filtered.Renamed = result.Renamed
	}

	return filtered
}

func outputResults(result *DiffResult, format string) error {
	switch format {
	case "json":
		return outputJSON(result)
	case "table":
		return outputTable(result)
	case "markdown":
		return outputMarkdown(result)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func outputJSON(result *DiffResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputTable(result *DiffResult) error {
	fmt.Println("DIFF RESULTS")
	fmt.Println("============")

	if len(result.New) > 0 {
		fmt.Printf("\nNEW (%d):\n", len(result.New))
		for _, test := range result.New {
			fmt.Printf("  + %s\n", test.Name)
		}
	}

	if len(result.Removed) > 0 {
		fmt.Printf("\nREMOVED (%d):\n", len(result.Removed))
		for _, test := range result.Removed {
			fmt.Printf("  - %s\n", test.Name)
		}
	}

	if len(result.PotentiallyModified) > 0 {
		fmt.Printf("\nPOTENTIALLY MODIFIED (%d):\n", len(result.PotentiallyModified))
		for _, test := range result.PotentiallyModified {
			fmt.Printf("  ~ %s (file containing test was modified)\n", test.Name)
		}
	}

	if len(result.Modified) > 0 {
		fmt.Printf("\nMODIFIED (%d):\n", len(result.Modified))
		for _, test := range result.Modified {
			fmt.Printf("  * %s (AST analysis confirmed changes)\n", test.Name)
		}
	}

	if len(result.Renamed) > 0 {
		fmt.Printf("\nRENAMED (%d):\n", len(result.Renamed))
		for _, test := range result.Renamed {
			fmt.Printf("  > %s -> %s\n", test.OldName, test.NewName)
		}
	}

	return nil
}

func outputMarkdown(result *DiffResult) error {
	fmt.Println("# Diff Results")

	if len(result.New) > 0 {
		fmt.Printf("\n## New (%d)\n\n", len(result.New))
		for _, test := range result.New {
			fmt.Printf("- ➕ `%s`\n", test.Name)
		}
	}

	if len(result.Removed) > 0 {
		fmt.Printf("\n## Removed (%d)\n\n", len(result.Removed))
		for _, test := range result.Removed {
			fmt.Printf("- ❌ `%s`\n", test.Name)
		}
	}

	if len(result.PotentiallyModified) > 0 {
		fmt.Printf("\n## Potentially Modified (%d)\n\n", len(result.PotentiallyModified))
		for _, test := range result.PotentiallyModified {
			fmt.Printf("- ⚠️ `%s`\n", test.Name)
			fmt.Printf("  - Reason: File containing test was modified\n")
		}
	}

	if len(result.Modified) > 0 {
		fmt.Printf("\n## Modified (%d)\n\n", len(result.Modified))
		for _, test := range result.Modified {
			fmt.Printf("- ✏️ `%s`\n", test.Name)
			fmt.Printf("  - Reason: AST analysis confirmed source code changes\n")
		}
	}

	if len(result.Renamed) > 0 {
		fmt.Printf("\n## Renamed (%d)\n\n", len(result.Renamed))
		for _, test := range result.Renamed {
			fmt.Printf("- 🔄 `%s` → `%s`\n", test.OldName, test.NewName)
		}
	}

	return nil
}

func hasDifferences(result *DiffResult) bool {
	return len(result.New) > 0 || len(result.Removed) > 0 || len(result.PotentiallyModified) > 0 || len(result.Modified) > 0 || len(result.Renamed) > 0
}

// getModifiedFiles returns list of files modified between merge base and working tree
func getModifiedFiles(mergeBase string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", mergeBase)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get modified files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// isPotentiallyModified checks if any file referenced by the test was modified
func isPotentiallyModified(test TestMetadata, modifiedFiles []string) bool {
	// Extract unique file paths from test's code locations
	testFiles := make(map[string]bool)
	for _, location := range test.CodeLocations {
		filePath, _, err := parseCodeLocation(location)
		if err != nil {
			continue
		}
		testFiles[filePath] = true
	}

	// Check if any of the test's files were modified
	for _, modifiedFile := range modifiedFiles {
		if testFiles[modifiedFile] {
			return true
		}
	}

	return false
}

// detectExactRenames finds renames by matching identical code locations
func detectExactRenames(removed, added []TestMetadata) []RenamedTest {
	var renames []RenamedTest

	for _, removedTest := range removed {
		for _, addedTest := range added {
			// Check if code locations are identical
			if codeLocationsEqual(removedTest.CodeLocations, addedTest.CodeLocations) {
				renames = append(renames, RenamedTest{
					OldName:    removedTest.Name,
					NewName:    addedTest.Name,
					BaseTest:   removedTest,
					HeadTest:   addedTest,
					Similarity: 1.0, // Exact match
				})
				break // Each removed test can only match one added test
			}
		}
	}

	return renames
}

// codeLocationsEqual checks if two code location slices are identical
func codeLocationsEqual(loc1, loc2 []string) bool {
	if len(loc1) != len(loc2) {
		return false
	}

	// Create maps for comparison (order shouldn't matter)
	map1 := make(map[string]bool)
	map2 := make(map[string]bool)

	for _, loc := range loc1 {
		map1[loc] = true
	}
	for _, loc := range loc2 {
		map2[loc] = true
	}

	// Check if maps are identical
	for loc := range map1 {
		if !map2[loc] {
			return false
		}
	}

	return true
}

// parseCodeLocation extracts file path and line number from a code location string
func parseCodeLocation(codeLocation string) (string, int, error) {
	parts := strings.Split(codeLocation, ":")
	if len(parts) < 2 {
		return "", 0, fmt.Errorf("invalid code location format: %s", codeLocation)
	}

	filePath := parts[0]
	lineNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid line number in code location %s: %w", codeLocation, err)
	}

	return filePath, lineNum, nil
}
