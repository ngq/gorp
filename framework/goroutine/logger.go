package goroutine

import "github.com/ngq/gorp/framework/contract"

// LoggerFromContainer tries to get framework logger.
func LoggerFromContainer(c contract.Container) contract.Logger {
	v, err := c.Make(contract.LogKey)
	if err != nil {
		return nil
	}
	l, _ := v.(contract.Logger)
	return l
}
