package ginkgo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

const spawnModeEnv = "OTE_SPAWN_TEST_MODE"

// TestMain lets this test binary stand in for the extension binary that
// SpawnProcessToRunTest re-executes as os.Args[0]: when OTE_SPAWN_TEST_MODE is
// set, the process acts as the run-test subprocess instead of running tests.
func TestMain(m *testing.M) {
	switch mode := os.Getenv(spawnModeEnv); mode {
	case "":
		os.Exit(m.Run())
	case "pass":
		emitResult(extensiontests.ResultPassed)
		os.Exit(0)
	case "fail-after-sigint":
		waitForSIGINT()
		emitResult(extensiontests.ResultFailed)
		os.Exit(1)
	case "pass-after-sigint":
		waitForSIGINT()
		emitResult(extensiontests.ResultPassed)
		os.Exit(0)
	case "hang":
		signal.Ignore(syscall.SIGINT)
		for {
			time.Sleep(time.Hour) // SIGABRT from the escalation ends the process
		}
	case "garbage":
		fmt.Println("this is not json")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "unknown %s: %q\n", spawnModeEnv, mode)
		os.Exit(2)
	}
}

func emitResult(r extensiontests.Result) {
	_ = json.NewEncoder(os.Stdout).Encode([]extensiontests.ExtensionTestResult{{
		Name:   "spawned test",
		Result: r,
	}})
}

func waitForSIGINT() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
}

func spawnWithMode(t *testing.T, mode string, timeout time.Duration) *extensiontests.ExtensionTestResult {
	t.Helper()
	t.Setenv(spawnModeEnv, mode)
	oldGrace := escalationGrace
	escalationGrace = 250 * time.Millisecond
	t.Cleanup(func() { escalationGrace = oldGrace })
	return SpawnProcessToRunTest(context.Background(), "spawned test", timeout)
}

func TestSpawnProcessToRunTestPassing(t *testing.T) {
	result := spawnWithMode(t, "pass", time.Minute)
	if result.Result != extensiontests.ResultPassed {
		t.Fatalf("expected passed, got %s (error: %s)", result.Result, result.Error)
	}
	if strings.Contains(result.Error, "test timed out") {
		t.Errorf("passing test should not be marked as timed out, got: %s", result.Error)
	}
}

func TestSpawnProcessToRunTestTimeoutWithParseableOutput(t *testing.T) {
	result := spawnWithMode(t, "fail-after-sigint", 100*time.Millisecond)
	if result.Result != extensiontests.ResultFailed {
		t.Fatalf("expected failed, got %s", result.Result)
	}
	if !strings.HasPrefix(result.Error, "test timed out after 100ms; subprocess exit:") {
		t.Errorf("expected error to lead with timeout message, got: %s", result.Error)
	}
}

func TestSpawnProcessToRunTestTimeoutPassRace(t *testing.T) {
	// a pass that raced the escalation stays a pass
	result := spawnWithMode(t, "pass-after-sigint", 100*time.Millisecond)
	if result.Result != extensiontests.ResultPassed {
		t.Fatalf("expected passed, got %s (error: %s)", result.Result, result.Error)
	}
	if strings.Contains(result.Error, "test timed out") {
		t.Errorf("passing test should not be marked as timed out, got: %s", result.Error)
	}
}

func TestSpawnProcessToRunTestHungSubprocess(t *testing.T) {
	result := spawnWithMode(t, "hang", 100*time.Millisecond)
	if result.Result != extensiontests.ResultFailed {
		t.Fatalf("expected failed, got %s", result.Result)
	}
	if !strings.HasPrefix(result.Error, "test timed out after 100ms") {
		t.Errorf("expected error to lead with timeout message even with a stack dump, got: %.200s", result.Error)
	}
}

func TestSpawnProcessToRunTestCrashIsNotATimeout(t *testing.T) {
	result := spawnWithMode(t, "garbage", time.Minute)
	if result.Result != extensiontests.ResultFailed {
		t.Fatalf("expected failed, got %s", result.Result)
	}
	if strings.Contains(result.Error, "test timed out") {
		t.Errorf("fast crash should not be marked as timed out, got: %s", result.Error)
	}
	if !strings.Contains(result.Error, "Deserialization Error") {
		t.Errorf("expected deserialization error for garbage output, got: %s", result.Error)
	}
}
