-- ============================================================
-- nop-go 微服务平台数据库初始化脚本
-- 创建所有 20 个微服务所需的数据库
-- ============================================================
-- 执行方式: mysql -u root -p < init_databases.sql
-- 或通过 Docker entrypoint 自动执行
-- ============================================================

-- 设置字符集
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- ----------------------------------------------------------
-- 1. 用户服务数据库 (User Service)
-- 存储用户账户、个人信息、地址等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_user
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 2. 订单服务数据库 (Order Service)
-- 存储订单、购物车、订单历史等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_order
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 3. 商品服务数据库 (Product Service)
-- 存储商品信息、SKU、库存、价格等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_product
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 4. 支付服务数据库 (Payment Service)
-- 存储支付记录、退款、支付方式等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_payment
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 5. 物流服务数据库 (Shipping Service)
-- 存储配送信息、物流跟踪、仓库等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_shipping
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 6. 折扣服务数据库 (Discount Service)
-- 存储优惠券、促销活动、折扣规则等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_discount
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 7. 税务服务数据库 (Tax Service)
-- 存储税率配置、税务规则、税务记录等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_tax
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 8. 媒体服务数据库 (Media Service)
-- 存储媒体文件元数据、图片处理配置等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_media
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 9. 消息服务数据库 (Message Service)
-- 存储站内消息、通知模板、消息队列等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_message
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 10. 本地化服务数据库 (Localization Service)
-- 存储多语言资源、货币配置、时区设置等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_localization
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 11. 店铺服务数据库 (Store Service)
-- 存储店铺信息、店铺配置、店铺主题等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_store
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 12. 安全服务数据库 (Security Service)
-- 存储安全策略、审计日志、访问控制等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_security
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 13. 供应商服务数据库 (Vendor Service)
-- 存储供应商信息、采购订单、供应商评价等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_vendor
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 14. 联盟营销服务数据库 (Affiliate Service)
-- 存储推广员信息、佣金记录、推广链接等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_affiliate
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 15. 内容管理服务数据库 (Content Service)
-- 存储CMS内容、页面模板、博客文章等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_content
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 16. 目录服务数据库 (Directory Service)
-- 存储分类结构、标签、目录树等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_directory
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 17. SEO服务数据库 (SEO Service)
-- 存储SEO配置、元数据、URL映射等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_seo
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 18. GDPR合规服务数据库 (GDPR Service)
-- 存储用户同意记录、数据导出请求、删除记录等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_gdpr
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 19. 插件服务数据库 (Plugin Service)
-- 存储插件注册、插件配置、插件状态等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_plugin
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 20. 日志服务数据库 (Logging Service)
-- 存储集中式日志、错误记录、性能指标等
-- ----------------------------------------------------------
CREATE DATABASE IF NOT EXISTS nop_logging
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

-- ----------------------------------------------------------
-- 授权 nop_user 用户访问所有数据库
-- ----------------------------------------------------------
-- 创建用户（如果不存在）
CREATE USER IF NOT EXISTS 'nop_user'@'%' IDENTIFIED BY 'nop_pass';

-- 授予所有服务数据库的完整权限
GRANT ALL PRIVILEGES ON nop_user.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_order.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_product.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_payment.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_shipping.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_discount.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_tax.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_media.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_message.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_localization.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_store.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_security.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_vendor.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_affiliate.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_content.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_directory.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_seo.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_gdpr.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_plugin.* TO 'nop_user'@'%';
GRANT ALL PRIVILEGES ON nop_logging.* TO 'nop_user'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- ----------------------------------------------------------
-- 输出创建结果
-- ----------------------------------------------------------
SELECT 'Database initialization completed!' AS status;
SELECT
    SCHEMA_NAME AS 'Database Name',
    DEFAULT_CHARACTER_SET_NAME AS 'Charset',
    DEFAULT_COLLATION_NAME AS 'Collation'
FROM information_schema.SCHEMATA
WHERE SCHEMA_NAME LIKE 'nop_%'
ORDER BY SCHEMA_NAME;
