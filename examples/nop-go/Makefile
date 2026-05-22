# nop-go - nopCommerce 微服务重构项目
# 基于 gorp multi-independent 模板构建

HARBOR_REGISTRY ?= harbor.example.com
HARBOR_NAMESPACE ?= nop-go
IMAGE_TAG ?= latest

# 所有服务列表
SERVICES := user order product payment shipping discount tax media message localization store security vendor affiliate content directory seo gdpr plugin logging

# 镜像名称
USER_IMAGE := nop-go-user-service
ORDER_IMAGE := nop-go-order-service
PRODUCT_IMAGE := nop-go-product-service
PAYMENT_IMAGE := nop-go-payment-service
SHIPPING_IMAGE := nop-go-shipping-service
DISCOUNT_IMAGE := nop-go-discount-service
TAX_IMAGE := nop-go-tax-service
MEDIA_IMAGE := nop-go-media-service
MESSAGE_IMAGE := nop-go-message-service
LOCALIZATION_IMAGE := nop-go-localization-service
STORE_IMAGE := nop-go-store-service
SECURITY_IMAGE := nop-go-security-service
VENDOR_IMAGE := nop-go-vendor-service
AFFILIATE_IMAGE := nop-go-affiliate-service
CONTENT_IMAGE := nop-go-content-service
DIRECTORY_IMAGE := nop-go-directory-service
SEO_IMAGE := nop-go-seo-service
GDPR_IMAGE := nop-go-gdpr-service
PLUGIN_IMAGE := nop-go-plugin-service
LOGGING_IMAGE := nop-go-logging-service

.PHONY: all generate gen-wire build clean test tidy \
        run-user run-order run-product run-payment run-shipping run-discount run-tax \
        run-media run-message run-localization run-store run-security run-vendor \
        run-affiliate run-content run-directory run-seo run-gdpr run-plugin run-logging \
        deploy-local deploy-local-down build-images deploy-local-build harbor-tag harbor-push \
        init-db

all: tidy generate

generate: gen-wire

gen-wire:
	@echo ">>> 生成 Wire 依赖注入代码..."
	wire ./services/user/cmd
	wire ./services/order/cmd
	wire ./services/product/cmd
	wire ./services/payment/cmd
	wire ./services/shipping/cmd
	wire ./services/discount/cmd
	wire ./services/tax/cmd
	wire ./services/media/cmd
	wire ./services/message/cmd
	wire ./services/localization/cmd
	wire ./services/store/cmd
	wire ./services/security/cmd
	wire ./services/vendor/cmd
	wire ./services/affiliate/cmd
	wire ./services/content/cmd
	wire ./services/directory/cmd
	wire ./services/seo/cmd
	wire ./services/gdpr/cmd
	wire ./services/plugin/cmd
	wire ./services/logging/cmd

build:
	@echo ">>> 编译所有服务..."
	@mkdir -p bin
	go build -o bin/user ./services/user/cmd
	go build -o bin/order ./services/order/cmd
	go build -o bin/product ./services/product/cmd
	go build -o bin/payment ./services/payment/cmd
	go build -o bin/shipping ./services/shipping/cmd
	go build -o bin/discount ./services/discount/cmd
	go build -o bin/tax ./services/tax/cmd
	go build -o bin/media ./services/media/cmd
	go build -o bin/message ./services/message/cmd
	go build -o bin/localization ./services/localization/cmd
	go build -o bin/store ./services/store/cmd
	go build -o bin/security ./services/security/cmd
	go build -o bin/vendor ./services/vendor/cmd
	go build -o bin/affiliate ./services/affiliate/cmd
	go build -o bin/content ./services/content/cmd
	go build -o bin/directory ./services/directory/cmd
	go build -o bin/seo ./services/seo/cmd
	go build -o bin/gdpr ./services/gdpr/cmd
	go build -o bin/plugin ./services/plugin/cmd
	go build -o bin/logging ./services/logging/cmd

tidy:
	@echo ">>> 同步 Go workspace..."
	go work sync

# 运行单个服务
run-user:
	go run ./services/user/cmd

run-order:
	go run ./services/order/cmd

run-product:
	go run ./services/product/cmd

run-payment:
	go run ./services/payment/cmd

run-shipping:
	go run ./services/shipping/cmd

run-discount:
	go run ./services/discount/cmd

run-tax:
	go run ./services/tax/cmd

run-media:
	go run ./services/media/cmd

run-message:
	go run ./services/message/cmd

run-localization:
	go run ./services/localization/cmd

run-store:
	go run ./services/store/cmd

run-security:
	go run ./services/security/cmd

run-vendor:
	go run ./services/vendor/cmd

run-affiliate:
	go run ./services/affiliate/cmd

run-content:
	go run ./services/content/cmd

run-directory:
	go run ./services/directory/cmd

run-seo:
	go run ./services/seo/cmd

run-gdpr:
	go run ./services/gdpr/cmd

run-plugin:
	go run ./services/plugin/cmd

run-logging:
	go run ./services/logging/cmd

test:
	@echo ">>> 运行所有测试..."
	go test ./services/user/... ./services/order/... ./services/product/... \
		./services/payment/... ./services/shipping/... ./services/discount/... \
		./services/tax/... ./services/media/... ./services/message/... \
		./services/localization/... ./services/store/... ./services/security/... \
		./services/vendor/... ./services/affiliate/... ./services/content/... \
		./services/directory/... ./services/seo/... ./services/gdpr/... \
		./services/plugin/... ./services/logging/... ./shared/... -v -coverprofile=coverage.out

test-coverage: test
	@echo ">>> 生成覆盖率报告..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

clean:
	@echo ">>> 清理构建产物..."
	rm -rf bin/
	rm -f *.db
	rm -f coverage.out coverage.html

# 初始化数据库
init-db:
	@echo ">>> 初始化数据库..."
	mysql -h 192.168.3.250 -P 3306 -u admin -p123456 < deploy/sql/init_databases.sql

# Docker 镜像构建
build-images:
	@echo ">>> 构建 Docker 镜像..."
	docker build -f deploy/docker/Dockerfile.user -t $(USER_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.order -t $(ORDER_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.product -t $(PRODUCT_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.payment -t $(PAYMENT_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.shipping -t $(SHIPPING_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.discount -t $(DISCOUNT_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.tax -t $(TAX_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.media -t $(MEDIA_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.message -t $(MESSAGE_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.localization -t $(LOCALIZATION_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.store -t $(STORE_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.security -t $(SECURITY_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.vendor -t $(VENDOR_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.affiliate -t $(AFFILIATE_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.content -t $(CONTENT_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.directory -t $(DIRECTORY_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.seo -t $(SEO_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.gdpr -t $(GDPR_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.plugin -t $(PLUGIN_IMAGE):$(IMAGE_TAG) .
	docker build -f deploy/docker/Dockerfile.logging -t $(LOGGING_IMAGE):$(IMAGE_TAG) .

deploy-local:
	docker-compose -f deploy/compose/docker-compose.yaml up -d

deploy-local-down:
	docker-compose -f deploy/compose/docker-compose.yaml down

deploy-local-build: build-images deploy-local

harbor-tag:
	docker tag $(USER_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(USER_IMAGE):$(IMAGE_TAG)
	docker tag $(ORDER_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(ORDER_IMAGE):$(IMAGE_TAG)
	docker tag $(PRODUCT_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(PRODUCT_IMAGE):$(IMAGE_TAG)
	docker tag $(PAYMENT_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(PAYMENT_IMAGE):$(IMAGE_TAG)
	docker tag $(SHIPPING_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(SHIPPING_IMAGE):$(IMAGE_TAG)
	docker tag $(DISCOUNT_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(DISCOUNT_IMAGE):$(IMAGE_TAG)
	docker tag $(TAX_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(TAX_IMAGE):$(IMAGE_TAG)
	docker tag $(MEDIA_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(MEDIA_IMAGE):$(IMAGE_TAG)
	docker tag $(MESSAGE_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(MESSAGE_IMAGE):$(IMAGE_TAG)
	docker tag $(LOCALIZATION_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(LOCALIZATION_IMAGE):$(IMAGE_TAG)
	docker tag $(STORE_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(STORE_IMAGE):$(IMAGE_TAG)
	docker tag $(SECURITY_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(SECURITY_IMAGE):$(IMAGE_TAG)
	docker tag $(VENDOR_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(VENDOR_IMAGE):$(IMAGE_TAG)
	docker tag $(AFFILIATE_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(AFFILIATE_IMAGE):$(IMAGE_TAG)
	docker tag $(CONTENT_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(CONTENT_IMAGE):$(IMAGE_TAG)
	docker tag $(DIRECTORY_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(DIRECTORY_IMAGE):$(IMAGE_TAG)
	docker tag $(SEO_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(SEO_IMAGE):$(IMAGE_TAG)
	docker tag $(GDPR_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(GDPR_IMAGE):$(IMAGE_TAG)
	docker tag $(PLUGIN_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(PLUGIN_IMAGE):$(IMAGE_TAG)
	docker tag $(LOGGING_IMAGE):$(IMAGE_TAG) $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(LOGGING_IMAGE):$(IMAGE_TAG)

harbor-push: build-images harbor-tag
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(USER_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(ORDER_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(PRODUCT_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(PAYMENT_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(SHIPPING_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(DISCOUNT_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(TAX_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(MEDIA_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(MESSAGE_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(LOCALIZATION_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(STORE_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(SECURITY_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(VENDOR_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(AFFILIATE_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(CONTENT_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(DIRECTORY_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(SEO_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(GDPR_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(PLUGIN_IMAGE):$(IMAGE_TAG)
	docker push $(HARBOR_REGISTRY)/$(HARBOR_NAMESPACE)/$(LOGGING_IMAGE):$(IMAGE_TAG)

# 帮助信息
help:
	@echo "nop-go - nopCommerce 微服务重构项目"
	@echo ""
	@echo "常用命令:"
	@echo "  make all          - 初始化项目（tidy + generate）"
	@echo "  make generate     - 生成 Wire 依赖注入代码"
	@echo "  make build        - 编译所有服务"
	@echo "  make test         - 运行所有测试"
	@echo "  make test-coverage- 生成覆盖率报告"
	@echo "  make run-<service>- 运行指定服务（如 run-user）"
	@echo "  make deploy-local - 本地 Docker 部署"
	@echo "  make init-db      - 初始化数据库"
	@echo ""
	@echo "可用服务: user, order, product, payment, shipping, discount, tax,"
	@echo "          media, message, localization, store, security, vendor,"
	@echo "          affiliate, content, directory, seo, gdpr, plugin, logging"
