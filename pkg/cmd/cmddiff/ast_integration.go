package cmddiff

import (
	"fmt"
	"os"
	"strings"
)

// performASTAnalysis enhances the diff result with AST-based analysis
func performASTAnalysis(result *DiffResult, mergeBase string, flags *DiffFlags) (*DiffResult, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return result, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create AST analyzer
	analyzer, err := NewGitVersionAnalyzer(workingDir, flags.File)
	if err != nil {
		return result, fmt.Errorf("failed to create git analyzer: %w", err)
	}
	defer analyzer.Cleanup()

	// Analyze each potentially modified test
	var trueModified []ModifiedTest
	var unchanged []TestMetadata

	for _, potentiallyModified := range result.PotentiallyModified {
		if flags.Verbose {
			fmt.Fprintf(os.Stderr, "  Analyzing: %s\n", potentiallyModified.Name)
		}

		// Get AST analysis for both versions
		baseAnalysis, err := analyzer.AnalyzeTestAtRef(potentiallyModified.BaseTest, mergeBase)
		if err != nil {
			if flags.Verbose {
				fmt.Fprintf(os.Stderr, "    Warning: failed to analyze base version: %v\n", err)
			}
			// Keep as potentially modified if we can't analyze
			trueModified = append(trueModified, potentiallyModified)
			continue
		}

		headAnalysis, err := analyzer.AnalyzeTestAtRef(potentiallyModified.HeadTest, "HEAD")
		if err != nil {
			if flags.Verbose {
				fmt.Fprintf(os.Stderr, "    Warning: failed to analyze head version: %v\n", err)
			}
			// Keep as potentially modified if we can't analyze
			trueModified = append(trueModified, potentiallyModified)
			continue
		}

		// Compare the analyses
		astResult, err := compareTestAnalyses(baseAnalysis, headAnalysis)
		if err != nil {
			if flags.Verbose {
				fmt.Fprintf(os.Stderr, "    Warning: failed to compare analyses: %v\n", err)
			}
			// Keep as potentially modified if we can't compare
			trueModified = append(trueModified, potentiallyModified)
			continue
		}

		// Classify based on AST comparison
		if astResult.IsActuallyModified {
			if flags.Verbose {
				fmt.Fprintf(os.Stderr, "    → MODIFIED (%d changes):\n", len(astResult.ChangedBlocks))
				for _, change := range astResult.ChangedBlocks {
					fmt.Fprintf(os.Stderr, "      • %s\n", change.Description)
					if change.ChangeType == "modified" {
						fmt.Fprintf(os.Stderr, "        Old: %s\n", truncateSource(change.OldSource))
						fmt.Fprintf(os.Stderr, "        New: %s\n", truncateSource(change.NewSource))
					} else if change.ChangeType == "added" {
						fmt.Fprintf(os.Stderr, "        Added: %s\n", truncateSource(change.NewSource))
					} else if change.ChangeType == "removed" {
						fmt.Fprintf(os.Stderr, "        Removed: %s\n", truncateSource(change.OldSource))
					}
				}
			}
			// Add detailed change information to the original ModifiedTest
			enhancedModified := potentiallyModified
			// Note: We could extend ModifiedTest to include AST details if needed
			trueModified = append(trueModified, enhancedModified)
		} else {
			if flags.Verbose {
				fmt.Fprintf(os.Stderr, "    → UNCHANGED (no actual source differences)\n")
			}
			// Test content didn't actually change
			unchanged = append(unchanged, potentiallyModified.HeadTest)
		}
	}

	// Update result with refined analysis
	if len(trueModified) > 0 {
		// Move confirmed modifications to the "Modified" field
		result.Modified = trueModified
		result.PotentiallyModified = nil // Clear potentially modified since we have definitive results
	} else {
		// No actual modifications found
		result.PotentiallyModified = nil
		result.Modified = nil
	}

	// Note: We could extend DiffResult to include an "Unchanged" field if desired
	// For now, tests that were potentially modified but aren't actually modified
	// just don't appear in any modified lists anymore

	return result, nil
}

// enhanceModifiedTestWithAST adds AST analysis details to a ModifiedTest
func enhanceModifiedTestWithAST(base ModifiedTest, astResult *ASTModifiedTest) ModifiedTest {
	// For now, just return the original
	// We could extend ModifiedTest to include AST details if needed
	return base
}

// truncateSource truncates source code for readable logging
func truncateSource(source string) string {
	const maxLength = 100
	// Replace newlines with spaces for single-line display
	cleaned := strings.ReplaceAll(source, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")
	// Collapse multiple spaces
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	if len(cleaned) <= maxLength {
		return cleaned
	}
	return cleaned[:maxLength-3] + "..."
}
