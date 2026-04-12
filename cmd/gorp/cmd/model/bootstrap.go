package model

import (
	"github.com/ngq/gorp/framework"
	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/app"
	"github.com/ngq/gorp/framework/provider/config"
	"github.com/ngq/gorp/framework/provider/log"
	"github.com/ngq/gorp/framework/provider/orm/inspect"
	sqlxprovider "github.com/ngq/gorp/framework/provider/orm/sqlx"
)

// bootstrap 为 model 子命令创建一个“最小装配”的运行时。
//
// 中文说明：
// - model 相关命令只需要 app/config/log/sqlx/inspect 这些能力。
// - 因此这里没有像主 bootstrap 那样装配 HTTP、cron、gorm、ssh 等完整 provider 集合。
// - 这样可以降低命令启动成本，也更符合职责最小化。
func bootstrap() (*framework.Application, contract.Container, error) {
	appRuntime := framework.NewApplication()
	c := appRuntime.Container()

	if err := c.RegisterProvider(app.NewProvider()); err != nil {
		return nil, nil, err
	}
	if err := c.RegisterProvider(config.NewProvider()); err != nil {
		return nil, nil, err
	}
	if err := c.RegisterProvider(log.NewProvider()); err != nil {
		return nil, nil, err
	}
	if err := c.RegisterProvider(sqlxprovider.NewProvider()); err != nil {
		return nil, nil, err
	}
	if err := c.RegisterProvider(inspect.NewProvider()); err != nil {
		return nil, nil, err
	}
	return appRuntime, c, nil
}
