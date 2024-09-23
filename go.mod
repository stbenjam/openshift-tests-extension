module github.com/openshift-eng/openshift-tests-extension

go 1.22.4

require (
	github.com/onsi/ginkgo/v2 v2.20.2
	github.com/onsi/gomega v1.30.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	k8s.io/apimachinery v0.31.1
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20240827171923-fa2c70bbbfe5 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/onsi/ginkgo/v2 => github.com/openshift/onsi-ginkgo/v2 v2.6.1-0.20240314173009-2cd07f4ca53d
