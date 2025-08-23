-- AI文档助手数据库初始化脚本
-- 创建时间: 2024-08-23
-- 版本: 1.0

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS ai_assistant DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ai_assistant;

-- ===================================
-- 用户表
-- ===================================
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` varchar(64) NOT NULL COMMENT '用户ID',
  `username` varchar(100) NOT NULL COMMENT '用户名',
  `email` varchar(200) DEFAULT NULL COMMENT '邮箱',
  `department` varchar(100) DEFAULT NULL COMMENT '部门',
  `role` enum('admin','user','readonly') NOT NULL DEFAULT 'user' COMMENT '角色',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像URL',
  `status` enum('active','inactive') NOT NULL DEFAULT 'active' COMMENT '状态',
  `last_login` timestamp NULL DEFAULT NULL COMMENT '最后登录时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_email` (`email`),
  KEY `idx_department` (`department`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ===================================
-- 文档表
-- ===================================
DROP TABLE IF EXISTS `documents`;
CREATE TABLE `documents` (
  `id` varchar(64) NOT NULL COMMENT '文档ID',
  `title` varchar(255) NOT NULL COMMENT '文档标题',
  `content` longtext COMMENT '文档内容',
  `document_type` varchar(50) DEFAULT NULL COMMENT '文档类型',
  `version` varchar(50) DEFAULT NULL COMMENT '版本号',
  `author` varchar(100) DEFAULT NULL COMMENT '作者',
  `department` varchar(100) DEFAULT NULL COMMENT '所属部门',
  `file_path` varchar(500) DEFAULT NULL COMMENT '文件路径',
  `file_size` bigint DEFAULT NULL COMMENT '文件大小(字节)',
  `file_hash` varchar(128) DEFAULT NULL COMMENT '文件哈希',
  `tags` text COMMENT '标签(JSON数组)',
  `metadata` json DEFAULT NULL COMMENT '元数据',
  `status` enum('active','archived','deleted') NOT NULL DEFAULT 'active' COMMENT '状态',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_title` (`title`),
  KEY `idx_author` (`author`),
  KEY `idx_department` (`department`),
  KEY `idx_document_type` (`document_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_file_hash` (`file_hash`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档表';

-- ===================================
-- 文档块表
-- ===================================
DROP TABLE IF EXISTS `document_chunks`;
CREATE TABLE `document_chunks` (
  `id` varchar(64) NOT NULL COMMENT '文档块ID',
  `document_id` varchar(64) NOT NULL COMMENT '文档ID',
  `chunk_index` int NOT NULL COMMENT '块索引',
  `content` text NOT NULL COMMENT '块内容',
  `vector_id` varchar(64) DEFAULT NULL COMMENT 'Milvus向量ID',
  `metadata` json DEFAULT NULL COMMENT '块元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_document_id` (`document_id`),
  KEY `idx_vector_id` (`vector_id`),
  KEY `idx_chunk_index` (`chunk_index`),
  CONSTRAINT `fk_chunk_document` FOREIGN KEY (`document_id`) REFERENCES `documents` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档块表';

-- ===================================
-- 查询历史表
-- ===================================
DROP TABLE IF EXISTS `query_history`;
CREATE TABLE `query_history` (
  `id` varchar(64) NOT NULL COMMENT '查询ID',
  `user_id` varchar(64) DEFAULT NULL COMMENT '用户ID',
  `query` text NOT NULL COMMENT '查询内容',
  `response` longtext COMMENT '回答内容',
  `response_time_ms` int DEFAULT NULL COMMENT '响应时间(毫秒)',
  `retrieved_docs` json DEFAULT NULL COMMENT '检索到的文档ID列表',
  `satisfaction_score` int DEFAULT NULL COMMENT '满意度评分(1-5)',
  `feedback` text COMMENT '用户反馈',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_satisfaction_score` (`satisfaction_score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='查询历史表';

-- ===================================
-- 文档分享表
-- ===================================
DROP TABLE IF EXISTS `document_shares`;
CREATE TABLE `document_shares` (
  `id` varchar(64) NOT NULL COMMENT '分享ID',
  `user_id` varchar(64) NOT NULL COMMENT '分享用户ID',
  `document_id` varchar(64) NOT NULL COMMENT '文档ID',
  `permission` enum('read','analyze','index','full') NOT NULL DEFAULT 'read' COMMENT '权限级别',
  `share_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '分享时间',
  `expire_time` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `status` enum('active','expired','revoked') NOT NULL DEFAULT 'active' COMMENT '状态',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_document_id` (`document_id`),
  KEY `idx_status` (`status`),
  CONSTRAINT `fk_share_document` FOREIGN KEY (`document_id`) REFERENCES `documents` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档分享表';

-- ===================================
-- 插入初始数据
-- ===================================

-- 插入管理员用户
INSERT INTO `users` (`id`, `username`, `email`, `department`, `role`, `status`) VALUES
('user_admin_001', 'admin', 'admin@company.com', 'IT部', 'admin', 'active'),
('user_test_001', 'test_user', 'test@company.com', '产品部', 'user', 'active');

-- 插入示例文档
INSERT INTO `documents` (`id`, `title`, `content`, `document_type`, `version`, `author`, `department`, `tags`, `status`) VALUES
('doc_001', '用户登录功能需求文档', 
'# 用户登录功能需求文档

## 1. 功能概述
设计用户登录、注册、密码重置等用户认证相关功能。

## 2. 详细需求

### 2.1 用户登录
- 支持手机号/邮箱登录
- 验证码有效期设定为5分钟
- 登录失败锁定策略：连续5次失败锁定30分钟
- 支持记住登录状态(30天)

### 2.2 用户注册
- 实名认证要求
- 手机号唯一性校验
- 密码强度要求：8-20位，包含数字、字母、特殊字符
- 注册验证码有效期3分钟

### 2.3 密码重置
- 支持手机号找回
- 支持邮箱找回
- 重置链接有效期30分钟

## 3. 非功能需求
- 响应时间: 登录请求 < 500ms
- 并发支持: 1000 TPS
- 可用性: 99.9%
- 安全性: 密码加盐哈希存储

## 4. 接口设计
- POST /api/v1/auth/login
- POST /api/v1/auth/register  
- POST /api/v1/auth/reset-password

## 5. 变更历史
- v1.0 (2023-08): 初版需求
- v2.0 (2023-12): 增加验证码有效期调整
- v2.3 (2024-03): 优化密码策略', 
'PRD', 'v2.3', '张三', '产品部', '["用户认证", "登录", "验证码"]', 'active'),

('doc_002', '支付模块技术设计文档',
'# 支付模块技术设计文档

## 1. 架构设计
采用微服务架构，支持多种支付方式。

## 2. 支付方式
- 支付宝
- 微信支付
- 银联支付
- Apple Pay

## 3. 技术实现
- 支付网关统一接口
- 异步回调处理
- 订单状态同步
- 对账文件处理

## 4. 安全措施
- API签名验证
- HTTPS加密传输
- 敏感信息脱敏
- 支付密码二次验证

## 5. 性能要求
- 支付请求响应时间 < 2秒
- 支持并发 500 TPS
- 99.9% 可用性

## 6. 版本历史
- v1.0 (2023-10): 基础支付功能
- v1.2 (2023-12): 增加微信支付
- v1.5 (2024-02): 支持分期付款',
'TDD', 'v1.5', '李四', '技术部', '["支付", "微服务", "架构设计"]', 'active'),

('doc_003', 'API接口规范文档',
'# API接口设计规范

## 1. 接口规范
所有API接口遵循RESTful设计规范。

## 2. 请求格式
- Content-Type: application/json
- 字符编码: UTF-8
- 请求方式: GET/POST/PUT/DELETE

## 3. 响应格式
```json
{
  "code": 200,
  "message": "success", 
  "data": {},
  "timestamp": 1628563200000
}
```

## 4. 错误码定义
- 200: 成功
- 400: 请求参数错误
- 401: 认证失败
- 403: 权限不足
- 404: 资源不存在
- 500: 服务器内部错误

## 5. 认证方式
使用JWT Token进行接口认证。

## 6. 限流策略
- 普通接口: 100/分钟
- 登录接口: 10/分钟
- 上传接口: 5/分钟',
'API', 'v1.0', '王五', '技术部', '["API", "接口", "规范"]', 'active');

-- 创建对应的文档块示例
INSERT INTO `document_chunks` (`id`, `document_id`, `chunk_index`, `content`, `metadata`) VALUES
('chunk_001_001', 'doc_001', 1, '用户登录验证码有效期设定为5分钟，超时后需重新获取。支持短信和邮箱两种验证方式。', '{"section": "登录功能", "page": 1}'),
('chunk_001_002', 'doc_001', 2, '登录失败锁定策略：连续5次失败锁定30分钟。可通过管理员解锁或等待时间过期。', '{"section": "安全策略", "page": 1}'),
('chunk_002_001', 'doc_002', 1, '支付模块采用微服务架构，支持支付宝、微信支付、银联支付等多种支付方式。', '{"section": "架构设计", "page": 1}'),
('chunk_003_001', 'doc_003', 1, 'API接口响应格式采用统一的JSON结构，包含code、message、data和timestamp字段。', '{"section": "接口规范", "page": 1}');

-- ===================================
-- 创建视图和存储过程
-- ===================================

-- 文档统计视图
CREATE OR REPLACE VIEW `v_document_stats` AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as doc_count,
    COUNT(DISTINCT author) as author_count,
    COUNT(DISTINCT department) as dept_count
FROM documents 
WHERE status = 'active' 
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- 用户查询统计视图  
CREATE OR REPLACE VIEW `v_user_query_stats` AS
SELECT 
    user_id,
    COUNT(*) as query_count,
    AVG(response_time_ms) as avg_response_time,
    AVG(satisfaction_score) as avg_satisfaction,
    MAX(created_at) as last_query_time
FROM query_history 
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY user_id;

-- 每日使用统计视图
CREATE OR REPLACE VIEW `v_daily_usage_stats` AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as query_count,
    COUNT(DISTINCT user_id) as unique_users,
    AVG(response_time_ms) as avg_response_time,
    AVG(CASE WHEN satisfaction_score IS NOT NULL THEN satisfaction_score END) as avg_satisfaction
FROM query_history 
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- ===================================
-- 创建索引优化
-- ===================================

-- 文档全文搜索索引
ALTER TABLE documents ADD FULLTEXT(title, content);

-- 查询历史时间范围查询优化
CREATE INDEX idx_query_history_time_range ON query_history(created_at, user_id);

-- 文档块检索优化
CREATE INDEX idx_chunk_content ON document_chunks(document_id, chunk_index);

-- ===================================
-- 设置数据库参数
-- ===================================

-- 设置innodb缓冲池大小(如果内存足够)
-- SET GLOBAL innodb_buffer_pool_size = 1073741824; -- 1GB

-- 启用慢查询日志
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 2;

-- 设置字符集
SET GLOBAL character_set_server = utf8mb4;
SET GLOBAL collation_server = utf8mb4_unicode_ci;

-- ===================================
-- 创建数据清理存储过程
-- ===================================

DELIMITER //

-- 清理过期查询历史
CREATE PROCEDURE CleanExpiredQueryHistory(IN days_to_keep INT)
BEGIN
    DELETE FROM query_history 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL days_to_keep DAY)
      AND satisfaction_score IS NULL;
    
    SELECT ROW_COUNT() as deleted_rows;
END //

-- 文档统计存储过程
CREATE PROCEDURE GetDocumentStats(IN stat_days INT)
BEGIN
    SELECT 
        COUNT(*) as total_documents,
        COUNT(CASE WHEN status = 'active' THEN 1 END) as active_documents,
        COUNT(DISTINCT author) as unique_authors,
        COUNT(DISTINCT department) as unique_departments,
        AVG(file_size) as avg_file_size
    FROM documents 
    WHERE created_at >= DATE_SUB(NOW(), INTERVAL stat_days DAY);
END //

DELIMITER ;

-- ===================================
-- 设置外键约束
-- ===================================
SET FOREIGN_KEY_CHECKS = 1;

-- 添加用户外键约束
ALTER TABLE query_history ADD CONSTRAINT fk_query_user 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE document_shares ADD CONSTRAINT fk_share_user 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ===================================
-- 创建触发器
-- ===================================

DELIMITER //

-- 文档删除时同步删除相关数据
CREATE TRIGGER tr_document_delete
BEFORE UPDATE ON documents
FOR EACH ROW
BEGIN
    IF NEW.status = 'deleted' AND OLD.status != 'deleted' THEN
        SET NEW.deleted_at = NOW();
    END IF;
END //

-- 查询历史插入时的数据验证
CREATE TRIGGER tr_query_history_insert
BEFORE INSERT ON query_history
FOR EACH ROW
BEGIN
    IF NEW.satisfaction_score IS NOT NULL THEN
        IF NEW.satisfaction_score < 1 OR NEW.satisfaction_score > 5 THEN
            SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '满意度评分必须在1-5之间';
        END IF;
    END IF;
END //

DELIMITER ;

-- 完成初始化
SELECT '✅ AI文档助手数据库初始化完成！' as status;

-- 显示初始化统计
SELECT 
    'users' as table_name, COUNT(*) as record_count FROM users
UNION ALL
SELECT 
    'documents', COUNT(*) FROM documents  
UNION ALL
SELECT 
    'document_chunks', COUNT(*) FROM document_chunks
UNION ALL
SELECT
    'query_history', COUNT(*) FROM query_history;