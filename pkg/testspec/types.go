package testspec

import "github.com/onsi/ginkgo/v2/types"

type Lifecycle string

var LifecycleInforming Lifecycle = "informing"
var LifecycleBlocking Lifecycle = "blocking"

type TestSpec struct {
	Name string `json:"name"`

	// OtherNames contains a list of historical names for this test. If the test gets renamed in the future,
	// this slice must report all the previous names for this test to preserve history.
	OtherNames []string `json:"other_names"`

	// Labels are single string values to apply to the test spec
	Labels []string `json:"labels"`

	// Tags are key:value pairs
	Tags map[string]string `json:"tags"`

	// Resources gives optional information about what's required to run this test.
	Resources Resources `json:"resources"`

	// Lifecycle informs the executor whether the test is informing only, and should not cause the
	// overall job run to fail, or if it's blocking where a failure of the test is fatal.
	// Informing lifecycle tests can be used temporarily to gather information about a test's stability.
	// Tests must not remain informing forever.
	Lifecycle Lifecycle `json:"lifecycle"`

	// Run invokes a test (TODO:relace gingko spec state with our own)
	Run func() types.SpecState `json:"-"`
}

type Resources struct {
	Isolation Isolation `json:"isolation"`
	Memory    string    `json:"memory"`
	Duration  string    `json:"duration"`
	Timeout   string    `json:"timeout"`
}

type Isolation struct {
	Mode     string   `json:"mode"`
	Conflict []string `json:"conflict"`
}
