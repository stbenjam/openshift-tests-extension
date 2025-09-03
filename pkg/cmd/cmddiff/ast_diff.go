package cmddiff

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openshift-eng/openshift-tests-extension/pkg/testanalysis"
)

// ASTModifiedTest represents a test with detailed AST-based change analysis
type ASTModifiedTest struct {
	Name               string                           `json:"name"`
	BaseTest           TestMetadata                     `json:"base"`
	HeadTest           TestMetadata                     `json:"head"`
	BaseAnalysis       *testanalysis.TestSourceAnalysis `json:"baseAnalysis,omitempty"`
	HeadAnalysis       *testanalysis.TestSourceAnalysis `json:"headAnalysis,omitempty"`
	ChangedBlocks      []BlockChange                    `json:"changedBlocks,omitempty"`
	IsActuallyModified bool                             `json:"isActuallyModified"`
}

// BlockChange represents a specific change in a Ginkgo block
type BlockChange struct {
	BlockType   string `json:"blockType"`  // "describe", "it", "beforeEach", etc.
	ChangeType  string `json:"changeType"` // "added", "removed", "modified"
	OldSource   string `json:"oldSource,omitempty"`
	NewSource   string `json:"newSource,omitempty"`
	Description string `json:"description"`
}

// EnhancedDiffResult extends DiffResult with AST-based analysis
type EnhancedDiffResult struct {
	*DiffResult                   // Embed original result
	Modified    []ASTModifiedTest `json:"modified,omitempty"`  // Real changes detected by AST
	Unchanged   []TestMetadata    `json:"unchanged,omitempty"` // Tests that appeared to change but didn't
}

// GitVersionAnalyzer handles analyzing tests at different git refs
type GitVersionAnalyzer struct {
	workingDir   string
	metadataFile string
	tempDir      string
}

// NewGitVersionAnalyzer creates a new analyzer for git-based version comparison
func NewGitVersionAnalyzer(workingDir, metadataFile string) (*GitVersionAnalyzer, error) {
	// Create temp directory for git operations
	tempDir, err := ioutil.TempDir("", "ast-diff-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	return &GitVersionAnalyzer{
		workingDir:   workingDir,
		metadataFile: metadataFile,
		tempDir:      tempDir,
	}, nil
}

// Cleanup removes temporary directories
func (g *GitVersionAnalyzer) Cleanup() error {
	return os.RemoveAll(g.tempDir)
}

// AnalyzeTestAtRef extracts the source code for a test at a specific git ref
func (g *GitVersionAnalyzer) AnalyzeTestAtRef(testSpec TestMetadata, gitRef string) (*testanalysis.TestSourceAnalysis, error) {
	if gitRef == "HEAD" || gitRef == "" {
		// Use current working tree
		return g.analyzeTestInWorkingTree(testSpec)
	}

	// Checkout files to temp directory
	tempWorkDir, err := g.checkoutFilesToTemp(testSpec, gitRef)
	if err != nil {
		return nil, fmt.Errorf("failed to checkout files for ref %s: %w", gitRef, err)
	}

	// Analyze in temp directory
	return g.analyzeTestInDirectory(testSpec, tempWorkDir)
}

// analyzeTestInWorkingTree analyzes a test in the current working directory
func (g *GitVersionAnalyzer) analyzeTestInWorkingTree(testSpec TestMetadata) (*testanalysis.TestSourceAnalysis, error) {
	return g.analyzeTestInDirectory(testSpec, g.workingDir)
}

// analyzeTestInDirectory analyzes a test in a specific directory
func (g *GitVersionAnalyzer) analyzeTestInDirectory(testSpec TestMetadata, workDir string) (*testanalysis.TestSourceAnalysis, error) {
	// Convert TestMetadata to testanalysis.TestSpec
	analysisSpec := &testanalysis.TestSpec{
		Name:          testSpec.Name,
		CodeLocations: testSpec.CodeLocations,
		// Copy other fields as needed
	}

	extractor := testanalysis.NewSourceExtractor(workDir)
	return extractor.AnalyzeTest(analysisSpec)
}

// checkoutFilesToTemp checks out all files referenced by a test to a temp directory
func (g *GitVersionAnalyzer) checkoutFilesToTemp(testSpec TestMetadata, gitRef string) (string, error) {
	// Create a subdirectory for this ref
	refTempDir := filepath.Join(g.tempDir, gitRef)
	if err := os.MkdirAll(refTempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create ref temp dir: %w", err)
	}

	// Get unique files referenced by this test
	files := g.getUniqueFilesFromTest(testSpec)

	for _, file := range files {
		if err := g.checkoutFileToTemp(file, gitRef, refTempDir); err != nil {
			// If file doesn't exist in this ref, that's ok - the test might be new
			continue
		}
	}

	return refTempDir, nil
}

// getUniqueFilesFromTest extracts unique file paths from test's CodeLocations
func (g *GitVersionAnalyzer) getUniqueFilesFromTest(testSpec TestMetadata) []string {
	fileSet := make(map[string]bool)

	for _, location := range testSpec.CodeLocations {
		// Parse "file:line" format
		parts := strings.Split(location, ":")
		if len(parts) >= 2 {
			fileSet[parts[0]] = true
		}
	}

	files := make([]string, 0, len(fileSet))
	for file := range fileSet {
		files = append(files, file)
	}

	return files
}

// checkoutFileToTemp checks out a single file from git to temp directory
func (g *GitVersionAnalyzer) checkoutFileToTemp(filePath, gitRef, tempDir string) error {
	// Create directory structure in temp
	fullTempPath := filepath.Join(tempDir, filePath)
	tempFileDir := filepath.Dir(fullTempPath)
	if err := os.MkdirAll(tempFileDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp file dir: %w", err)
	}

	// Get file content from git
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", gitRef, filePath))
	cmd.Dir = g.workingDir

	content, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get file %s from ref %s: %w", filePath, gitRef, err)
	}

	// Write to temp file
	return ioutil.WriteFile(fullTempPath, content, 0644)
}

// compareTestAnalyses compares two test source analyses to detect real changes
func compareTestAnalyses(baseAnalysis, headAnalysis *testanalysis.TestSourceAnalysis) (*ASTModifiedTest, error) {
	result := &ASTModifiedTest{
		BaseAnalysis:  baseAnalysis,
		HeadAnalysis:  headAnalysis,
		ChangedBlocks: []BlockChange{},
	}

	// Compare Ginkgo blocks
	changes := findBlockChanges(baseAnalysis.GinkgoBlocks, headAnalysis.GinkgoBlocks)
	result.ChangedBlocks = append(result.ChangedBlocks, changes...)

	// Compare helper functions (if implemented later)
	helperChanges := findBlockChanges(baseAnalysis.HelperFunctions, headAnalysis.HelperFunctions)
	result.ChangedBlocks = append(result.ChangedBlocks, helperChanges...)

	// Determine if actually modified
	result.IsActuallyModified = len(result.ChangedBlocks) > 0

	return result, nil
}

// findBlockChanges compares two sets of code blocks to find differences
func findBlockChanges(baseBlocks, headBlocks []testanalysis.CodeBlock) []BlockChange {
	var changes []BlockChange

	// Create normalized maps for content-based comparison
	baseByContent := make(map[string]testanalysis.CodeBlock)
	headByContent := make(map[string]testanalysis.CodeBlock)
	baseByType := make(map[string][]testanalysis.CodeBlock)
	headByType := make(map[string][]testanalysis.CodeBlock)

	// Index blocks by normalized content and by type
	for _, block := range baseBlocks {
		normalized := normalizeSource(block.Source)
		baseByContent[normalized] = block
		baseByType[block.Type] = append(baseByType[block.Type], block)
	}

	for _, block := range headBlocks {
		normalized := normalizeSource(block.Source)
		headByContent[normalized] = block
		headByType[block.Type] = append(headByType[block.Type], block)
	}

	// Track which blocks have been matched to avoid double-counting
	matchedBase := make(map[string]bool)
	matchedHead := make(map[string]bool)

	// First pass: Find exact matches (no changes)
	for normalizedContent := range baseByContent {
		if _, exists := headByContent[normalizedContent]; exists {
			matchedBase[normalizedContent] = true
			matchedHead[normalizedContent] = true
		}
	}

	// Second pass: Find modifications using fuzzy matching
	for blockType := range baseByType {
		baseBlocksOfType := baseByType[blockType]
		headBlocksOfType := headByType[blockType]

		modifications := findModificationPairs(baseBlocksOfType, headBlocksOfType, matchedBase, matchedHead)
		changes = append(changes, modifications...)
	}

	// Third pass: Find pure additions and removals (blocks that couldn't be matched)
	for normalizedContent, baseBlock := range baseByContent {
		if !matchedBase[normalizedContent] {
			changes = append(changes, BlockChange{
				BlockType:   baseBlock.Type,
				ChangeType:  "removed",
				OldSource:   baseBlock.Source,
				Description: fmt.Sprintf("%s block removed", baseBlock.Type),
			})
		}
	}

	for normalizedContent, headBlock := range headByContent {
		if !matchedHead[normalizedContent] {
			changes = append(changes, BlockChange{
				BlockType:   headBlock.Type,
				ChangeType:  "added",
				NewSource:   headBlock.Source,
				Description: fmt.Sprintf("%s block added", headBlock.Type),
			})
		}
	}

	return changes
}

// findModificationPairs finds blocks that were modified (similar but not identical content)
func findModificationPairs(baseBlocks, headBlocks []testanalysis.CodeBlock, matchedBase, matchedHead map[string]bool) []BlockChange {
	var changes []BlockChange
	usedHead := make(map[int]bool) // Track which head blocks have been matched

	for _, baseBlock := range baseBlocks {
		baseNormalized := normalizeSource(baseBlock.Source)

		// Skip if this block was already exactly matched
		if matchedBase[baseNormalized] {
			continue
		}

		bestMatch := -1
		bestSimilarity := 0.0

		// Find the best matching block in head
		for i, headBlock := range headBlocks {
			if usedHead[i] {
				continue // Already matched
			}

			headNormalized := normalizeSource(headBlock.Source)

			// Skip if this head block was already exactly matched
			if matchedHead[headNormalized] {
				continue
			}

			// Calculate similarity for potential modifications
			similarity := calculateSimilarity(baseNormalized, headNormalized)
			if similarity > bestSimilarity && similarity > 0.5 { // Threshold for considering it a modification
				bestMatch = i
				bestSimilarity = similarity
			}
		}

		// If we found a similar but not identical block, it's a modification
		if bestMatch >= 0 {
			usedHead[bestMatch] = true
			headNormalized := normalizeSource(headBlocks[bestMatch].Source)

			// Mark both as matched to prevent them from appearing as pure add/remove
			matchedBase[baseNormalized] = true
			matchedHead[headNormalized] = true

			changes = append(changes, BlockChange{
				BlockType:   baseBlock.Type,
				ChangeType:  "modified",
				OldSource:   baseBlock.Source,
				NewSource:   headBlocks[bestMatch].Source,
				Description: fmt.Sprintf("%s block modified", baseBlock.Type),
			})
		}
	}

	return changes
}

// calculateSimilarity calculates a simple similarity score between two strings
func calculateSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Simple similarity based on common substrings and length difference
	longer := a
	shorter := b
	if len(b) > len(a) {
		longer = b
		shorter = a
	}

	// Count matching characters in order
	matches := 0
	shorterIdx := 0
	for i := 0; i < len(longer) && shorterIdx < len(shorter); i++ {
		if longer[i] == shorter[shorterIdx] {
			matches++
			shorterIdx++
		}
	}

	// Similarity score: matches / longer length
	return float64(matches) / float64(len(longer))
}

// normalizeSource normalizes source code for comparison (remove whitespace differences, etc.)
func normalizeSource(source string) string {
	// Basic normalization - remove leading/trailing whitespace and normalize line endings
	lines := strings.Split(source, "\n")
	var normalized []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	return strings.Join(normalized, "\n")
}
