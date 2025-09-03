package testanalysis

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

// NewSourceExtractor creates a new source extractor
func NewSourceExtractor(workingDir string) *SourceExtractor {
	return &SourceExtractor{
		workingDir: workingDir,
	}
}

// AnalyzeTest performs complete analysis of a test's source code
func (se *SourceExtractor) AnalyzeTest(testSpec *TestSpec) (*TestSourceAnalysis, error) {
	analysis := &TestSourceAnalysis{
		TestName:        testSpec.Name,
		GinkgoBlocks:    []CodeBlock{},
		HelperFunctions: []CodeBlock{},
	}

	// Group locations by file to minimize file parsing
	locationsByFile := make(map[string][]CodeLocation)
	for _, location := range testSpec.CodeLocations {
		parts := strings.Split(location, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid location format: %s", location)
		}

		file := parts[0]
		lineStr := parts[1]
		line, err := strconv.Atoi(lineStr)
		if err != nil {
			return nil, fmt.Errorf("invalid line number in location %s: %w", location, err)
		}

		locationsByFile[file] = append(locationsByFile[file], CodeLocation{
			File: file,
			Line: line,
		})
	}

	// Extract code blocks from each file
	for file, locations := range locationsByFile {
		blocks, err := se.extractBlocksFromFile(file, locations, testSpec.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to extract blocks from %s: %w", file, err)
		}
		analysis.GinkgoBlocks = append(analysis.GinkgoBlocks, blocks...)
	}

	return analysis, nil
}

// CodeLocation represents a file and line number
type CodeLocation struct {
	File string
	Line int
}

// extractBlocksFromFile parses a Go file and extracts relevant code blocks
func (se *SourceExtractor) extractBlocksFromFile(file string, locations []CodeLocation, testName string) ([]CodeBlock, error) {
	// Resolve file path
	filePath := filepath.Join(se.workingDir, file)

	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	var blocks []CodeBlock
	targetItNode := se.findTargetItBlock(node, fset, testName)

	// Extract blocks for each location
	for _, loc := range locations {
		block, err := se.extractBlockAtLocation(node, fset, file, loc.Line, targetItNode)
		if err != nil {
			return nil, fmt.Errorf("failed to extract block at %s:%d: %w", file, loc.Line, err)
		}
		if block != nil {
			blocks = append(blocks, *block)
		}
	}

	return blocks, nil
}

// findTargetItBlock finds the specific It block for this test by matching the test name
func (se *SourceExtractor) findTargetItBlock(file *ast.File, fset *token.FileSet, testName string) ast.Node {
	var targetIt ast.Node

	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if se.isItCall(call) {
				// Check if this It call matches our test name
				if len(call.Args) > 0 {
					if lit, ok := call.Args[0].(*ast.BasicLit); ok {
						// Extract the test description from the string literal
						testDesc := strings.Trim(lit.Value, `"`)
						if strings.Contains(testName, testDesc) {
							targetIt = call
							return false // Found it, stop searching
						}
					}
				}
			}
		}
		return true
	})

	return targetIt
}

// extractBlockAtLocation extracts the AST node at the given line and converts it to source
func (se *SourceExtractor) extractBlockAtLocation(file *ast.File, fset *token.FileSet, filename string, line int, targetItNode ast.Node) (*CodeBlock, error) {
	var targetNode ast.Node
	var originalNode ast.Node // Keep reference to original for line numbers
	var blockType string

	// Find the node at the specified line - we need to find the closest CallExpr
	var bestNode ast.Node
	var bestCall *ast.CallExpr

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		pos := fset.Position(n.Pos())
		if pos.Line == line {
			// Store the node at this line
			if bestNode == nil {
				bestNode = n
			}

			// If we find a CallExpr at this line, prefer it
			if call, ok := n.(*ast.CallExpr); ok {
				bestCall = call
			}
		}
		return true
	})

	// Use the CallExpr if we found one, otherwise use the best node
	if bestCall != nil {
		originalNode = bestCall // Keep reference to original
		// Determine block type by function name
		if se.isDescribeCall(bestCall) {
			blockType = "describe"
			// For Describe blocks, filter out irrelevant It blocks
			targetNode = se.filterDescribeBlock(bestCall, targetItNode)
		} else if se.isItCall(bestCall) {
			blockType = "it"
			targetNode = bestCall
		} else if se.isBeforeEachCall(bestCall) {
			blockType = "beforeEach"
			targetNode = bestCall
		} else if se.isBeforeAllCall(bestCall) {
			blockType = "beforeAll"
			targetNode = bestCall
		} else if se.isAfterEachCall(bestCall) {
			blockType = "afterEach"
			targetNode = bestCall
		} else if se.isAfterAllCall(bestCall) {
			blockType = "afterAll"
			targetNode = bestCall
		} else {
			blockType = "unknown"
			targetNode = bestCall
		}
	} else if bestNode != nil {
		originalNode = bestNode
		blockType = "statement"
		targetNode = bestNode
	}

	if targetNode == nil {
		return nil, fmt.Errorf("no AST node found at line %d", line)
	}

	// Convert AST node back to source code
	source, err := se.nodeToSource(fset, targetNode)
	if err != nil {
		return nil, fmt.Errorf("failed to convert node to source: %w", err)
	}

	// Always use the original node for line number reporting
	startPos := fset.Position(originalNode.Pos())
	endPos := fset.Position(originalNode.End())

	// Determine if this is a filtered block
	isFiltered := (originalNode != targetNode)

	var description string
	if isFiltered {
		description = fmt.Sprintf("filtered %s block at %s:%d-%d (original range, filtered content)", blockType, filename, startPos.Line, endPos.Line)
	} else {
		description = fmt.Sprintf("%s block at %s:%d-%d", blockType, filename, startPos.Line, endPos.Line)
	}

	return &CodeBlock{
		Type:        blockType,
		Source:      source,
		File:        filename,
		StartLine:   startPos.Line,
		EndLine:     endPos.Line,
		Description: description,
		IsFiltered:  isFiltered,
	}, nil
}

// Helper functions to identify different Ginkgo call types
func (se *SourceExtractor) isDescribeCall(call *ast.CallExpr) bool {
	return se.isFunctionCall(call, "Describe")
}

func (se *SourceExtractor) isItCall(call *ast.CallExpr) bool {
	return se.isFunctionCall(call, "It")
}

func (se *SourceExtractor) isBeforeEachCall(call *ast.CallExpr) bool {
	return se.isFunctionCall(call, "BeforeEach")
}

func (se *SourceExtractor) isBeforeAllCall(call *ast.CallExpr) bool {
	return se.isFunctionCall(call, "BeforeAll")
}

func (se *SourceExtractor) isAfterEachCall(call *ast.CallExpr) bool {
	return se.isFunctionCall(call, "AfterEach")
}

func (se *SourceExtractor) isAfterAllCall(call *ast.CallExpr) bool {
	return se.isFunctionCall(call, "AfterAll")
}

// isFunctionCall checks if a CallExpr is calling a specific function name
func (se *SourceExtractor) isFunctionCall(call *ast.CallExpr, funcName string) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == funcName
	}
	return false
}

// filterDescribeBlock creates a new Describe block with only relevant content
func (se *SourceExtractor) filterDescribeBlock(describeCall *ast.CallExpr, targetItNode ast.Node) ast.Node {
	if targetItNode == nil {
		// If we don't have a target It node, return the full describe block
		return describeCall
	}

	// Create a copy of the describe call with filtered content
	filteredCall := &ast.CallExpr{
		Fun:  describeCall.Fun,
		Args: make([]ast.Expr, len(describeCall.Args)),
	}

	// Copy the first arguments (description, etc.)
	copy(filteredCall.Args, describeCall.Args)

	// Find the function literal (the last argument which contains the test body)
	if len(describeCall.Args) > 0 {
		if funcLit, ok := describeCall.Args[len(describeCall.Args)-1].(*ast.FuncLit); ok {
			// Create a new function literal with filtered body
			filteredFuncLit := &ast.FuncLit{
				Type: funcLit.Type,
				Body: &ast.BlockStmt{
					List: []ast.Stmt{},
				},
			}

			// Filter the statements in the function body
			for _, stmt := range funcLit.Body.List {
				if se.shouldIncludeStatement(stmt, targetItNode) {
					filteredFuncLit.Body.List = append(filteredFuncLit.Body.List, stmt)
				}
			}

			// Replace the last argument with our filtered function literal
			filteredCall.Args[len(filteredCall.Args)-1] = filteredFuncLit
			return filteredCall
		}
	}

	// Fallback to original if we can't parse the structure
	return describeCall
}

// shouldIncludeStatement determines if a statement should be included in the filtered Describe block
func (se *SourceExtractor) shouldIncludeStatement(stmt ast.Stmt, targetItNode ast.Node) bool {
	// Include variable declarations and other setup statements
	if _, ok := stmt.(*ast.DeclStmt); ok {
		return true
	}

	// Include assignment statements (setup code)
	if _, ok := stmt.(*ast.AssignStmt); ok {
		return true
	}

	// For expression statements, check if they're function calls
	if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
		if call, ok := exprStmt.X.(*ast.CallExpr); ok {
			// Include BeforeEach, BeforeAll, AfterEach, AfterAll
			if se.isBeforeEachCall(call) || se.isBeforeAllCall(call) ||
				se.isAfterEachCall(call) || se.isAfterAllCall(call) {
				return true
			}

			// Include only the target It block
			if se.isItCall(call) {
				return call == targetItNode
			}

			// Include other non-It function calls (helper setup, etc.)
			if !se.isItCall(call) {
				return true
			}
		}
	}

	// By default, include the statement (better to include too much than too little)
	return true
}

// nodeToSource converts an AST node back to Go source code
func (se *SourceExtractor) nodeToSource(fset *token.FileSet, node ast.Node) (string, error) {
	var buf strings.Builder
	if err := format.Node(&buf, fset, node); err != nil {
		return "", err
	}
	return buf.String(), nil
}
