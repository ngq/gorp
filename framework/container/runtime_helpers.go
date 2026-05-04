package container

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	gormdb "gorm.io/gorm"
)

func MakeDBRuntime(c runtimecontract.Container) (any, error) {
	return c.Make(datacontract.DBRuntimeKey)
}

func MakeRedis(c runtimecontract.Container) (datacontract.Redis, error) {
	v, err := c.Make(datacontract.RedisKey)
	if err != nil {
		return nil, err
	}
	return v.(datacontract.Redis), nil
}

func MakeCache(c runtimecontract.Container) (datacontract.Cache, error) {
	v, err := c.Make(datacontract.CacheKey)
	if err != nil {
		return nil, err
	}
	return v.(datacontract.Cache), nil
}

func MakeGormDB(c runtimecontract.Container) (*gormdb.DB, error) {
	v, err := c.Make(datacontract.GormKey)
	if err != nil {
		return nil, err
	}
	return v.(*gormdb.DB), nil
}

func MakeSQLX(c runtimecontract.Container) (*sqlx.DB, error) {
	v, err := c.Make(datacontract.SQLXKey)
	if err != nil {
		return nil, err
	}
	return v.(*sqlx.DB), nil
}

func MakeMessagePublisher(c runtimecontract.Container) (integrationcontract.MessagePublisher, error) {
	v, err := c.Make(integrationcontract.MessagePublisherKey)
	if err != nil {
		return nil, err
	}
	return v.(integrationcontract.MessagePublisher), nil
}

func MakeMessageSubscriber(c runtimecontract.Container) (integrationcontract.MessageSubscriber, error) {
	v, err := c.Make(integrationcontract.MessageSubscriberKey)
	if err != nil {
		return nil, err
	}
	return v.(integrationcontract.MessageSubscriber), nil
}

func MakeDistributedLock(c runtimecontract.Container) (datacontract.DistributedLock, error) {
	v, err := c.Make(datacontract.DistributedLockKey)
	if err != nil {
		return nil, err
	}
	return v.(datacontract.DistributedLock), nil
}

func MakeGRPCConnFactory(c runtimecontract.Container) (transportcontract.GRPCConnFactory, error) {
	v, err := c.Make(transportcontract.GRPCConnFactoryKey)
	if err != nil {
		return nil, err
	}
	return v.(transportcontract.GRPCConnFactory), nil
}

func MakeGRPCServerRegistrar(c runtimecontract.Container) (transportcontract.GRPCServerRegistrar, error) {
	v, err := c.Make(transportcontract.GRPCServerRegistrarKey)
	if err != nil {
		return nil, err
	}
	return v.(transportcontract.GRPCServerRegistrar), nil
}

func MakeCron(c runtimecontract.Container) (runtimecontract.Cron, error) {
	v, err := c.Make(runtimecontract.CronKey)
	if err != nil {
		return nil, err
	}
	return v.(runtimecontract.Cron), nil
}

func MakeLogger(c runtimecontract.Container) (observabilitycontract.Logger, error) {
	v, err := c.Make(observabilitycontract.LogKey)
	if err != nil {
		return nil, err
	}
	return v.(observabilitycontract.Logger), nil
}

func MakeValidator(c runtimecontract.Container) (datacontract.Validator, error) {
	v, err := c.Make(datacontract.ValidatorKey)
	if err != nil {
		return nil, err
	}
	return v.(datacontract.Validator), nil
}

func MakeRetry(c runtimecontract.Container) (resiliencecontract.Retry, error) {
	v, err := c.Make(resiliencecontract.RetryKey)
	if err != nil {
		return nil, err
	}
	return v.(resiliencecontract.Retry), nil
}

func MakeHost(c runtimecontract.Container) (runtimecontract.Host, error) {
	v, err := c.Make(runtimecontract.HostKey)
	if err != nil {
		return nil, err
	}
	return v.(runtimecontract.Host), nil
}

func MakeHTTP(c runtimecontract.Container) (transportcontract.HTTP, error) {
	v, err := c.Make(transportcontract.HTTPKey)
	if err != nil {
		return nil, err
	}
	return v.(transportcontract.HTTP), nil
}

func MakeHTTPRouter(c runtimecontract.Container) (transportcontract.HTTPRouter, error) {
	httpSvc, err := MakeHTTP(c)
	if err != nil {
		return nil, err
	}
	return httpSvc.Router(), nil
}

func MustMakeLogger(c runtimecontract.Container) observabilitycontract.Logger {
	v := c.MustMake(observabilitycontract.LogKey)
	return v.(observabilitycontract.Logger)
}

func MustMakeGorm(c runtimecontract.Container) *gormdb.DB {
	v := c.MustMake(datacontract.GormKey)
	return v.(*gormdb.DB)
}

func MustMakeHTTPRouter(c runtimecontract.Container) transportcontract.HTTPRouter {
	httpSvc := MustMakeHTTP(c)
	return httpSvc.Router()
}

func MustMakeHTTP(c runtimecontract.Container) transportcontract.HTTP {
	v := c.MustMake(transportcontract.HTTPKey)
	return v.(transportcontract.HTTP)
}

func MustMakeMessagePublisher(c runtimecontract.Container) integrationcontract.MessagePublisher {
	v := c.MustMake(integrationcontract.MessagePublisherKey)
	return v.(integrationcontract.MessagePublisher)
}

func MustMakeMessageSubscriber(c runtimecontract.Container) integrationcontract.MessageSubscriber {
	v := c.MustMake(integrationcontract.MessageSubscriberKey)
	return v.(integrationcontract.MessageSubscriber)
}

func MustMakeDistributedLock(c runtimecontract.Container) datacontract.DistributedLock {
	v := c.MustMake(datacontract.DistributedLockKey)
	return v.(datacontract.DistributedLock)
}

func MustMakeGRPCConnFactory(c runtimecontract.Container) transportcontract.GRPCConnFactory {
	v := c.MustMake(transportcontract.GRPCConnFactoryKey)
	return v.(transportcontract.GRPCConnFactory)
}

func MustMakeGRPCServerRegistrar(c runtimecontract.Container) transportcontract.GRPCServerRegistrar {
	v := c.MustMake(transportcontract.GRPCServerRegistrarKey)
	return v.(transportcontract.GRPCServerRegistrar)
}

func MustMakeValidator(c runtimecontract.Container) datacontract.Validator {
	v := c.MustMake(datacontract.ValidatorKey)
	return v.(datacontract.Validator)
}

func MustMakeRetry(c runtimecontract.Container) resiliencecontract.Retry {
	v := c.MustMake(resiliencecontract.RetryKey)
	return v.(resiliencecontract.Retry)
}

func MustMakeConfig(c runtimecontract.Container) datacontract.Config {
	v := c.MustMake(datacontract.ConfigKey)
	return v.(datacontract.Config)
}

func MustMakeCache(c runtimecontract.Container) datacontract.Cache {
	v := c.MustMake(datacontract.CacheKey)
	return v.(datacontract.Cache)
}

func PingDBRuntime(dbAny any) error {
	switch db := dbAny.(type) {
	case *gormdb.DB:
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	case *sql.DB:
		return db.Ping()
	case interface{ Ping() error }:
		return db.Ping()
	default:
		return nil
	}
}

func MustMakeJWT(c runtimecontract.Container) securitycontract.JWTService {
	return MustMakeJWTService(c)
}
