package ginkgo

import (
	"io"
)

// GinkgoDiscard is used to throw away all output, we collect the output from the spec report summaries.
var GinkgoDiscard = ginkgoDiscard{
	io.Discard,
}

type ginkgoDiscard struct {
	io.Writer
}

func (ginkgoDiscard) Print(_ ...interface{}) {
}

func (ginkgoDiscard) Printf(_ string, _ ...interface{}) {
}

func (ginkgoDiscard) Println(_ ...interface{}) {
}

func (ginkgoDiscard) TeeTo(_ io.Writer) {
}

func (ginkgoDiscard) Truncate() {
}

func (ginkgoDiscard) Bytes() []byte {
	return nil
}
func (ginkgoDiscard) Len() int {
	return 0
}

func (ginkgoDiscard) ClearTeeWriters() {
}
