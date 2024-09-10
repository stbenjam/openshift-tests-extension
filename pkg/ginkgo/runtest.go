package ginkgo

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
)

// TestOptions handles running a single test.
type TestOptions struct {
	Out    io.Writer
	ErrOut io.Writer
}

var _ ginkgo.GinkgoTestingT = &TestOptions{}

func NewTestOptions(out io.Writer, errOut io.Writer) *TestOptions {
	return &TestOptions{
		Out:    out,
		ErrOut: errOut,
	}
}

func ListTests() []*TestCase {
	tests := testsForSuite()
	sort.Slice(tests, func(i, j int) bool { return tests[i].Name < tests[j].Name })
	return tests
}

func (opt *TestOptions) RunTest(args []string, suiteDescription string) error {
	if len(args) != 1 {
		return fmt.Errorf("only a single test name may be passed")
	}

	// Ignore the upstream suite behavior within test execution
	ginkgo.GetSuite().ClearBeforeAndAfterSuiteNodes()
	tests := testsForSuite()
	var test *TestCase
	for _, t := range tests {
		if t.Name == args[0] {
			test = t
			break
		}
	}
	if test == nil {
		return fmt.Errorf("no test exists with that name: %s", args[0])
	}

	suiteConfig, reporterConfig := ginkgo.GinkgoConfiguration()
	suiteConfig.FocusStrings = []string{fmt.Sprintf("^ %s$", regexp.QuoteMeta(test.Name))}

	// These settings are matched to upstream's ginkgo configuration.
	suiteConfig.RandomizeAllSpecs = true
	suiteConfig.Timeout = 24 * time.Hour
	reporterConfig.NoColor = true
	reporterConfig.Verbose = true

	ginkgo.SetReporterConfig(reporterConfig)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Use the suite description passed from the cobra command
	ginkgo.GetSuite().RunSpec(test.spec, ginkgo.Labels{}, suiteDescription, cwd, ginkgo.GetFailer(), ginkgo.GetWriter(), suiteConfig, reporterConfig)

	var summary types.SpecReport
	for _, report := range ginkgo.GetSuite().GetReport().SpecReports {
		if report.NumAttempts > 0 {
			summary = report
		}
	}

	return handleSummary(summary, opt)
}

func handleSummary(summary types.SpecReport, opt *TestOptions) error {
	switch {
	case summary.State == types.SpecStatePassed:
		// do nothing
	case summary.State == types.SpecStateSkipped:
		if len(summary.Failure.Message) > 0 {
			fmt.Fprintf(opt.ErrOut, "skip [%s:%d]: %s\n", lastFilenameSegment(summary.Failure.Location.FileName), summary.Failure.Location.LineNumber, summary.Failure.Message)
		}
		if len(summary.Failure.ForwardedPanic) > 0 {
			fmt.Fprintf(opt.ErrOut, "skip [%s:%d]: %s\n", lastFilenameSegment(summary.Failure.Location.FileName), summary.Failure.Location.LineNumber, summary.Failure.ForwardedPanic)
		}
		return ExitError{Code: 3}
	case summary.State == types.SpecStateFailed, summary.State == types.SpecStatePanicked, summary.State == types.SpecStateInterrupted:
		if len(summary.Failure.ForwardedPanic) > 0 {
			if len(summary.Failure.Location.FullStackTrace) > 0 {
				fmt.Fprintf(opt.ErrOut, "\n%s\n", summary.Failure.Location.FullStackTrace)
			}
			fmt.Fprintf(opt.ErrOut, "fail [%s:%d]: Test Panicked: %s\n", lastFilenameSegment(summary.Failure.Location.FileName), summary.Failure.Location.LineNumber, summary.Failure.ForwardedPanic)
			return ExitError{Code: 1}
		}
		fmt.Fprintf(opt.ErrOut, "fail [%s:%d]: %s\n", lastFilenameSegment(summary.Failure.Location.FileName), summary.Failure.Location.LineNumber, summary.Failure.Message)
		return ExitError{Code: 1}
	default:
		return fmt.Errorf("unrecognized test case outcome: %#v", summary)
	}
	return nil
}

func (opt *TestOptions) Fail() {}

func lastFilenameSegment(filename string) string {
	if parts := strings.Split(filename, "/vendor/"); len(parts) > 1 {
		return parts[len(parts)-1]
	}
	if parts := strings.Split(filename, "/src/"); len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return filename
}

func testsForSuite() []*TestCase {
	var tests []*TestCase

	if !ginkgo.GetSuite().InPhaseBuildTree() {
		if err := ginkgo.GetSuite().BuildTree(); err != nil {
			panic(err)
		}
	}

	ginkgo.GetSuite().WalkTests(func(name string, spec types.TestSpec) {
		testCase := &TestCase{
			Name:      spec.Text(),
			locations: spec.CodeLocations(),
			spec:      spec,
		}
		tests = append(tests, testCase)
	})
	return tests
}

func (e ExitError) Error() string {
	return fmt.Sprintf("exit with code %d", e.Code)
}
