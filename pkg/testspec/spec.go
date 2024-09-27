package testspec

import "fmt"

func MustLifecycle(l string) Lifecycle {
	switch Lifecycle(l) {
	case LifecycleInforming, LifecycleBlocking:
		return Lifecycle(l)
	default:
		panic(fmt.Sprintf("unknown test lifecycle: %s", l))
	}
}
