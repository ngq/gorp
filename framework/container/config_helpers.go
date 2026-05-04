package container

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func MakeConfig(c runtimecontract.Container) (datacontract.Config, error) {
	v, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	return v.(datacontract.Config), nil
}
