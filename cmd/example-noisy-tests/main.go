package main

import (
	"fmt"
	"os"

	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd"
	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"

	// If using ginkgo, import your tests here
	_ "github.com/openshift-eng/openshift-tests-extension/test/example"
)

// example-noisy-tests is identical to example-tests, but deliberately writes
// realistic non-JSON garbage to stdout when invoked as a run-test subprocess.
// This exercises the extractJSON parsing pipeline end-to-end, ensuring that
// klog lines, Ginkgo reporter text, and other non-JSON output that commonly
// appears on stdout in extension binaries does not break result deserialization.
func main() {
	// When this binary is spawned as a subprocess by SpawnProcessToRunTest
	// (i.e. `run-test` is the first argument), emit realistic non-JSON content
	// to stdout before the cobra command runs. This simulates the kind of
	// stdout contamination that occurs in practice from klog, Ginkgo reporters,
	// and other libraries that write to stdout.
	if len(os.Args) > 1 && os.Args[1] == "run-test" {
		fmt.Fprintln(os.Stdout, `I0625 19:25:00.000000   12345 client.go:123] Connecting to apiserver...`)
		fmt.Fprintln(os.Stdout, `W0625 19:25:01.000000   12345 deprecation.go:78] API v1beta1 is deprecated, use v1`)
		fmt.Fprintln(os.Stdout, `I0625 19:25:01.100000   12345 round_trippers.go:466] GET https://api.example.com:6443/api 200 OK`)
		fmt.Fprintln(os.Stdout, `[BeforeEach] [sig-testing] openshift-tests-extension`)
		fmt.Fprintln(os.Stdout, `  /go/src/github.com/openshift-eng/openshift-tests-extension/test/example/example.go:28`)
		fmt.Fprintln(os.Stdout, `[It] [sig-testing] openshift-tests-extension should support passing tests`)
		fmt.Fprintln(os.Stdout, `STEP: Setting up test environment`)
	}

	// Extension registry
	registry := e.NewRegistry()

	ext := e.NewExtension("openshift", "payload", "example-noisy-tests")

	ext.AddSuite(
		e.Suite{
			Name:    "example/tests",
			Parents: []string{"openshift/conformance/parallel"},
		})

	ext.AddSuite(e.Suite{
		Name: "example/fast",
		Qualifiers: []string{
			`!labels.exists(l, l=="SLOW")`,
		},
	})

	ext.AddGlobalSuite(e.Suite{
		Name: "example/slow",
		Qualifiers: []string{
			`labels.exists(l, l=="SLOW")`,
		},
	})

	// If using Ginkgo, build test specs automatically
	specs, err := g.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite()
	if err != nil {
		panic(fmt.Sprintf("couldn't build extension test specs from ginkgo: %+v", err.Error()))
	}

	specs.Select(et.NameContains("[sig-testing] openshift-tests-extension should support test-skips via environment flags")).
		Include(et.PlatformEquals("aws"))

	ext.AddSpecs(specs)
	registry.Register(ext)

	root := &cobra.Command{
		Long: "OpenShift Tests Extension Example (Noisy Stdout)",
	}

	root.AddCommand(cmd.DefaultExtensionCommands(registry)...)

	if err := func() error {
		return root.Execute()
	}(); err != nil {
		os.Exit(1)
	}
}
