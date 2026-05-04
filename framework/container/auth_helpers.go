package container

import (
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

func MustMakeJWTService(c runtimecontract.Container) securitycontract.JWTService {
	v := c.MustMake(securitycontract.AuthJWTKey)
	return v.(securitycontract.JWTService)
}
