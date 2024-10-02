package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd"
	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"

	// If using ginkgo, import your tests here
	_ "github.com/openshift-eng/openshift-tests-extension/test/example"
)

func main() {
	// Extension registry
	registry := e.NewRegistry()

	// You can declare multiple extensions, but most people will probably only need to create one.
	ext := e.NewExtension("openshift", "payload", "example-tests")

	// Add suites to the extension. Specifying parents will cause the tests from this suite
	// to be included when a parent is invoked.
	ext.AddSuite(
		e.Suite{
			Name:    "example/tests",
			Parents: []string{"openshift/conformance/parallel"},
		})

	// The tests that a suite is composed of can be filtered by CEL expressions. By
	// default, the qualifiers only apply to tests from this extension.
	ext.AddSuite(e.Suite{
		Name: "example/fast",
		Qualifiers: []string{
			`!labels.exists(l, l=="SLOW")`,
		},
	})

	// Global suites' qualifiers will apply to all tests available, even
	// those outside of this extension (when invoked by origin).
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

	// You can also manually build a test specs list from other testing tooling
	// TODO: example

	// Modify specs:
	// Add label to all specs
	// specs = specs.AddLabel("SLOW")

	// Specs can be globally filtered...
	// specs = specs.MustFilter([]string{`name.contains("filter")`})

	// Or walked...
	// specs = specs.Walk(func(e *extensiontests.ExtensionTestSpec) {
	//	if strings.Contains(e.Name, "scale up") {
	//		e.Labels.Insert("SLOW")
	//	}
	//
	//  if val, ok := renameMap[e.Name]; ok {
	//		e.SetOtherNames(val...). // TODO
	//	}
	//})
	ext.AddSpecs(specs)
	registry.Register(ext)

	root := &cobra.Command{
		Long: "OpenShift Tests Extension Example",
	}

	root.AddCommand(cmd.DefaultExtensionCommands(registry)...)

	if err := func() error {
		return root.Execute()
	}(); err != nil {
		os.Exit(1)
	}
}
