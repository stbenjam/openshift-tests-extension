package ginkgo

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
)

func Suite(name string) ginkgo.Labels {
	return ginkgo.Label(fmt.Sprintf("Suite:%s", name))
}
