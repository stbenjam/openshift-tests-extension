package ginkgo

import "github.com/onsi/ginkgo/v2/types"

type TestCase struct {
	Name      string
	locations []types.CodeLocation
	spec      types.TestSpec
}

type ExitError struct {
	Code int
}
