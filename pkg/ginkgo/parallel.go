package ginkgo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/openshift-eng/openshift-tests-extension/pkg/dbtime"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

// escalationGrace is how long SpawnProcessToRunTest waits past the test timeout
// before interrupting the subprocess — the subprocess enforces --timeout itself
// and self-reports a structured timeout, so the grace period gives it a chance
// to do that first — and again how long it waits after SIGINT before requesting
// a stack dump via SIGABRT. It is a variable only so tests can shorten it.
var escalationGrace = time.Minute

func SpawnProcessToRunTest(ctx context.Context, testName string, timeout time.Duration) *extensiontests.ExtensionTestResult {
	// longerCtx backstop-kills a subprocess that survives the SIGINT/SIGABRT escalation below
	longerCtx, longerCancel := context.WithTimeout(ctx, timeout+15*time.Minute)
	defer longerCancel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	command := exec.CommandContext(longerCtx, os.Args[0], "run-test", "--output=json", fmt.Sprintf("--timeout=%s", timeout), testName)
	command.Stdout = stdout
	command.Stderr = stderr

	start := time.Now()
	err := command.Start()
	if err != nil {
		fmt.Fprintf(stderr, "Command Start Error: %v\n", err)
		return newTestResult(testName, extensiontests.ResultFailed, start, time.Now(), stdout, stderr)
	}

	var timedOut atomic.Bool
	done := make(chan struct{})
	escalationExited := make(chan struct{})
	go func() {
		defer close(escalationExited)
		select {
		case <-time.After(timeout + escalationGrace):
		case <-done:
			return
		case <-ctx.Done():
			// the caller aborted the whole run; longerCtx kills the process, and it is not a timeout
			return
		}
		timedOut.Store(true) // store before SIGINT so Wait() cannot return before the flag is visible
		_ = command.Process.Signal(syscall.SIGINT)
		// if the process is hung, send SIGABRT after a grace period for a stack dump
		select {
		case <-time.After(escalationGrace):
			_ = command.Process.Signal(syscall.SIGABRT)
		case <-done:
		}
	}()

	cmdErr := command.Wait()
	end := time.Now()
	close(done)
	<-escalationExited // join so timedOut can't change and no stray signal fires after this point

	// even if we have a cmdErr, if we were able to parse the result, trust the output
	result, parseErr := newTestResultFromOutput(stdout)
	if parseErr != nil {
		fmt.Fprintf(stderr, "Command Error: %v\n", cmdErr)
		fmt.Fprintf(stderr, "Deserialization Error: %v\n", parseErr)
		result = newTestResult(testName, extensiontests.ResultFailed, start, end, stdout, stderr)
	}

	// a pass or skip that raced the escalation stays as reported; only failures get the timeout marker
	if timedOut.Load() && result.Result == extensiontests.ResultFailed {
		markTimedOut(result, timeout, cmdErr)
	}
	return result
}

// markTimedOut stamps a failed result whose subprocess was interrupted by the
// timeout escalation with a leading "test timed out after ..." line — the string
// CI consumers match on in JUnit failure output — so it can't end up buried
// under a stack dump.
func markTimedOut(result *extensiontests.ExtensionTestResult, timeout time.Duration, cmdErr error) {
	msg := fmt.Sprintf("test timed out after %s", timeout)
	if cmdErr != nil {
		msg = fmt.Sprintf("%s; subprocess exit: %v", msg, cmdErr)
	}
	if result.Error != "" {
		msg += "\n" + result.Error
	}
	result.Error = msg
}

func newTestResultFromOutput(stdout *bytes.Buffer) (*extensiontests.ExtensionTestResult, error) {
	if len(stdout.Bytes()) == 0 {
		return nil, errors.New("no output from command")
	}

	jsonData, err := extractJSON(stdout.Bytes())
	if err != nil {
		return nil, err
	}

	// when the command runs correctly, we get json or json slice output
	retArray := []extensiontests.ExtensionTestResult{}
	if arrayItemErr := json.Unmarshal(jsonData, &retArray); arrayItemErr == nil {
		if len(retArray) != 1 {
			return nil, fmt.Errorf("expected 1 result, got %d results", len(retArray))
		}
		return &retArray[0], nil
	}

	// when the command runs correctly, we get json output
	ret := &extensiontests.ExtensionTestResult{}
	if singleItemErr := json.Unmarshal(jsonData, ret); singleItemErr != nil {
		return nil, singleItemErr
	}

	return ret, nil
}

// extractJSON finds the first JSON object or array in output, skipping any non-JSON
// lines that precede it (e.g. klog lines, Ginkgo reporter output). It also ignores
// trailing non-JSON content after the JSON payload. This is necessary because extension
// binaries may emit log output to stdout before or after the JSON result, which would
// otherwise cause deserialization failures.
func extractJSON(output []byte) ([]byte, error) {
	lines := bytes.Split(output, []byte("\n"))
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
			// Calculate byte offset to the start of the JSON content
			offset := 0
			for j := 0; j < i; j++ {
				offset += len(lines[j]) + 1 // +1 for the newline
			}

			var raw json.RawMessage
			dec := json.NewDecoder(bytes.NewReader(output[offset:]))
			if err := dec.Decode(&raw); err != nil {
				continue // not valid JSON, try next candidate line
			}
			return raw, nil
		}
	}

	return nil, fmt.Errorf("no JSON object or array found in output (%d bytes)", len(output))
}

func newTestResult(name string, result extensiontests.Result, start, end time.Time, stdout, stderr *bytes.Buffer) *extensiontests.ExtensionTestResult {
	duration := end.Sub(start)
	dbStart := dbtime.DBTime(start)
	dbEnd := dbtime.DBTime(end)
	ret := &extensiontests.ExtensionTestResult{
		Name:      name,
		Lifecycle: "", // lifecycle is completed one level above this.
		Duration:  int64(duration),
		StartTime: &dbStart,
		EndTime:   &dbEnd,
		Result:    result,
		Details:   nil,
	}

	if stdout != nil && stderr != nil {
		stdoutStr := stdout.String()
		stderrStr := stderr.String()

		ret.Output = fmt.Sprintf("STDOUT:\n%s\n\nSTDERR:\n%s\n", stdoutStr, stderrStr)

		// try to choose the best summary
		switch {
		case len(stderrStr) > 0 && len(stderrStr) < 5000:
			ret.Error = stderrStr
		case len(stderrStr) > 0 && len(stderrStr) >= 5000:
			ret.Error = stderrStr[len(stderrStr)-5000:]

		case len(stdoutStr) > 0 && len(stdoutStr) < 5000:
			ret.Error = stdoutStr
		case len(stdoutStr) > 0 && len(stdoutStr) >= 5000:
			ret.Error = stdoutStr[len(stdoutStr)-5000:]
		}
	}

	return ret
}
