package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd/cmdinfo"
	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd/cmdlist"
	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd/cmdrun"
	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"

	// If using ginkgo, import your tests here
	_ "github.com/openshift-eng/openshift-tests-extension/test/example"
)

func main() {
	// Extension registry
	registry := e.NewRegistry()

	ext := e.NewExtension("openshift", "payload", "default")
	ext.AddSuite(e.Suite{Name: "example/tests", Parents: []string{"openshift/conformance/parallel"}})

	// If using Ginkgo, build test specs automatically
	specs, err := g.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	// You can also manually build a test specs list from other testing tooling
	// TODO: example

	// Add label to all specs
	// specs = specs.AddLabel("SLOW")

	// Specs can be filtered...
	// specs = specs.MustFilter([]string{`name.contains("filter")`})

	// Or walked...
	// specs = specs.Walk(func(e *extensiontests.ExtensionTestSpec) {
	//	if strings.Contains(e.Name, "scale up") {
	//		e.Labels.Insert("SLOW")
	//	}
	//})

	ext.AddSpecs(specs)

	// If not using gingko build the test specs manually
	// TODO:example

	registry.Register(ext)

	root := &cobra.Command{
		Long: "OpenShift Tests Extension Example",
	}

	// TODO: Wire up extension stuff
	root.AddCommand(
		cmdrun.NewCommand("foobar"),
		cmdlist.NewCommand(registry),
		cmdinfo.NewCommand(),
	)

	gomega.RegisterFailHandler(ginkgo.Fail)

	if err := func() error {
		return root.Execute()
	}(); err != nil {
		var ex ExitError
		if errors.As(err, &ex) {
			fmt.Fprintf(os.Stderr, "Ginkgo exit error %d: %v\n", ex.Code, err)
			os.Exit(ex.Code)
		}
		os.Exit(1)
	}
}
