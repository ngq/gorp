package goroutine

import (
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// LoggerFromContainer tries to get framework logger.
func LoggerFromContainer(c runtimecontract.Container) observabilitycontract.Logger {
	v, err := c.Make(observabilitycontract.LogKey)
	if err != nil {
		return nil
	}
	l, _ := v.(observabilitycontract.Logger)
	return l
}
