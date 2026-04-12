-- nop-go 数据库初始化脚本
-- 创建所有微服务所需的数据库

-- 设置字符集
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 后台管理服务数据库
CREATE DATABASE IF NOT EXISTS nop_admin DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 客户服务数据库
CREATE DATABASE IF NOT EXISTS nop_customer DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 商品目录服务数据库
CREATE DATABASE IF NOT EXISTS nop_catalog DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 购物车服务数据库
CREATE DATABASE IF NOT EXISTS nop_cart DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 订单服务数据库
CREATE DATABASE IF NOT EXISTS nop_order DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- CMS 服务数据库
CREATE DATABASE IF NOT EXISTS nop_cms DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 库存服务数据库
CREATE DATABASE IF NOT EXISTS nop_inventory DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 通知服务数据库
CREATE DATABASE IF NOT EXISTS nop_notification DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 支付服务数据库
CREATE DATABASE IF NOT EXISTS nop_payment DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 价格服务数据库
CREATE DATABASE IF NOT EXISTS nop_price DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 物流服务数据库
CREATE DATABASE IF NOT EXISTS nop_shipping DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 多店铺服务数据库
CREATE DATABASE IF NOT EXISTS nop_store DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 媒体服务数据库
CREATE DATABASE IF NOT EXISTS nop_media DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 多语言服务数据库
CREATE DATABASE IF NOT EXISTS nop_localization DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 联盟营销服务数据库
CREATE DATABASE IF NOT EXISTS nop_affiliate DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- SEO 服务数据库
CREATE DATABASE IF NOT EXISTS nop_seo DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 导入导出服务数据库
CREATE DATABASE IF NOT EXISTS nop_import DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 主题服务数据库
CREATE DATABASE IF NOT EXISTS nop_theme DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- AI 服务数据库
CREATE DATABASE IF NOT EXISTS nop_ai DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 授权 root 用户访问所有数据库
GRANT ALL PRIVILEGES ON nop_*.* TO 'root'@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES;

-- 完成提示
SELECT 'nop-go 数据库初始化完成' AS message;