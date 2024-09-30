package extensiontests

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Lifecycle string

var LifecycleInforming Lifecycle = "informing"
var LifecycleBlocking Lifecycle = "blocking"

type ExtensionTestSpecs []*ExtensionTestSpec

type ExtensionTestSpec struct {
	Name string `json:"name"`

	// OtherNames contains a list of historical names for this test. If the test gets renamed in the future,
	// this slice must report all the previous names for this test to preserve history.
	OtherNames sets.Set[string] `json:"otherNames"`

	// Labels are single string values to apply to the test spec
	Labels sets.Set[string] `json:"labels"`

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
	Run func() *ExtensionTestResult `json:"-"`
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

type ExtensionTestResults []*ExtensionTestResult

type Result string

var ResultPassed Result = "passed"
var ResultSkipped Result = "skipped"
var ResultFailed Result = "failed"

type ExtensionTestResult struct {
	Name      string     `json:"name"`
	Duration  int64      `json:"duration"`
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	Result    Result     `json:"result"`
	Output    string     `json:"output"`
	Error     string     `json:"error"`
	Messages  []string   `json:"messages"`
}
