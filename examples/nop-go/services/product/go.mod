module nop-go/services/product

go 1.25

require (
	github.com/ngq/gorp v0.1.3
	github.com/ngq/gorp/contrib/circuitbreaker/sentinel v0.1.3
	github.com/ngq/gorp/contrib/configsource/etcd v0.1.3
	github.com/ngq/gorp/contrib/dlock/redis v0.1.3
	github.com/ngq/gorp/contrib/dtm/dtmsdk v0.1.3
	github.com/ngq/gorp/contrib/registry/etcd v0.1.3
	github.com/ngq/gorp/contrib/serviceauth/token v0.1.3
	github.com/ngq/gorp/contrib/tracing/otel v0.1.3
	github.com/gin-gonic/gin v1.9.1
	github.com/google/wire v0.6.0
	github.com/stretchr/testify v1.9.0
	gorm.io/gorm v1.25.5
)

replace github.com/ngq/gorp => E:/project/gin_plantfrom

