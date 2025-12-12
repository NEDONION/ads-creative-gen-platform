-- ============================================
-- AI 多尺寸广告创意生成平台 - 数据库设计
-- Database: ads_creative_platform
-- Version: 1.0
-- ============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS ads_creative_platform
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE ads_creative_platform;

-- ============================================
-- 1. 用户与权限管理
-- ============================================

-- 用户表
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL COMMENT 'UUID',
    username VARCHAR(64) UNIQUE NOT NULL COMMENT '用户名',
    email VARCHAR(255) UNIQUE NOT NULL COMMENT '邮箱',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    phone VARCHAR(20) COMMENT '手机号',
    avatar_url VARCHAR(512) COMMENT '头像URL',
    role ENUM('admin', 'user', 'viewer') DEFAULT 'user' COMMENT '角色',
    status ENUM('active', 'inactive', 'banned') DEFAULT 'active' COMMENT '状态',
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '软删除时间',
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 团队/项目表
CREATE TABLE projects (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL COMMENT 'UUID',
    name VARCHAR(128) NOT NULL COMMENT '项目名称',
    description TEXT COMMENT '项目描述',
    owner_id BIGINT UNSIGNED NOT NULL COMMENT '所有者ID',
    status ENUM('active', 'archived') DEFAULT 'active',
    settings JSON COMMENT '项目设置（Logo、默认风格等）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_owner (owner_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目表';

-- 项目成员表
CREATE TABLE project_members (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    role ENUM('owner', 'admin', 'member', 'viewer') DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_project_user (project_id, user_id),
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目成员表';

-- ============================================
-- 2. 创意生成核心表
-- ============================================

-- 创意生成任务表
CREATE TABLE creative_tasks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL COMMENT '任务UUID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '创建用户',
    project_id BIGINT UNSIGNED COMMENT '所属项目',

    -- 输入信息
    title VARCHAR(255) NOT NULL COMMENT '商品标题',
    selling_points JSON COMMENT '卖点列表',
    product_image_url VARCHAR(512) COMMENT '商品图URL',
    brand_logo_url VARCHAR(512) COMMENT '品牌Logo URL',

    -- 生成配置
    requested_formats JSON COMMENT '请求的尺寸格式 ["1:1", "4:5", "9:16"]',
    requested_styles JSON COMMENT '请求的风格 ["modern", "elegant", "vibrant"]',
    num_variants INT DEFAULT 3 COMMENT '每种风格的变体数',
    cta_text VARCHAR(64) COMMENT 'CTA文案',

    -- 任务状态
    status ENUM('pending', 'queued', 'processing', 'completed', 'failed', 'cancelled') DEFAULT 'pending',
    progress TINYINT DEFAULT 0 COMMENT '进度 0-100',
    error_message TEXT COMMENT '错误信息',

    -- 时间统计
    queued_at TIMESTAMP NULL COMMENT '入队时间',
    started_at TIMESTAMP NULL COMMENT '开始处理时间',
    completed_at TIMESTAMP NULL COMMENT '完成时间',
    processing_duration INT COMMENT '处理耗时（秒）',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,

    INDEX idx_user (user_id),
    INDEX idx_project (project_id),
    INDEX idx_status (status),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创意生成任务表';

-- 创意素材表
CREATE TABLE creative_assets (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL COMMENT '素材UUID',
    task_id BIGINT UNSIGNED NOT NULL COMMENT '任务ID',

    -- 素材信息
    format VARCHAR(20) NOT NULL COMMENT '尺寸格式: 1:1, 4:5, 9:16, 1200x628',
    width INT NOT NULL COMMENT '宽度（像素）',
    height INT NOT NULL COMMENT '高度（像素）',
    file_size INT COMMENT '文件大小（字节）',

    -- 存储信息
    storage_type ENUM('local', 'oss', 's3', 'minio') DEFAULT 'local',
    file_path VARCHAR(512) NOT NULL COMMENT '文件路径/Key',
    public_url VARCHAR(1024) COMMENT '公开访问URL',
    cdn_url VARCHAR(1024) COMMENT 'CDN URL',

    -- 生成元数据
    style VARCHAR(64) COMMENT '风格标签',
    variant_index INT COMMENT '变体索引',
    generation_prompt TEXT COMMENT '生成提示词',
    model_name VARCHAR(128) COMMENT '使用的模型名称',
    model_version VARCHAR(64) COMMENT '模型版本',
    generation_params JSON COMMENT '生成参数',

    -- 内容信息
    text_content JSON COMMENT '文本内容 {title, subtitle, cta}',
    has_logo BOOLEAN DEFAULT FALSE COMMENT '是否包含Logo',
    has_cta BOOLEAN DEFAULT FALSE COMMENT '是否包含CTA',

    -- 排序
    rank INT COMMENT '排名（基于CTR预测）',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    FOREIGN KEY (task_id) REFERENCES creative_tasks(id) ON DELETE CASCADE,

    INDEX idx_task (task_id),
    INDEX idx_format (format),
    INDEX idx_style (style),
    INDEX idx_rank (rank),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创意素材表';

-- ============================================
-- 3. 质量评估与评分
-- ============================================

-- 创意评分表
CREATE TABLE creative_scores (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    creative_id BIGINT UNSIGNED UNIQUE NOT NULL,

    -- 质量评分（0-1）
    quality_overall DECIMAL(4,3) COMMENT '综合质量评分',
    brightness_score DECIMAL(4,3) COMMENT '亮度评分',
    contrast_score DECIMAL(4,3) COMMENT '对比度评分',
    sharpness_score DECIMAL(4,3) COMMENT '清晰度评分',
    composition_score DECIMAL(4,3) COMMENT '构图评分',
    color_harmony_score DECIMAL(4,3) COMMENT '配色和谐度',

    -- CTR 预测
    ctr_prediction DECIMAL(5,4) COMMENT 'CTR预测值',
    ctr_confidence DECIMAL(4,3) COMMENT '预测置信度',
    model_version VARCHAR(64) COMMENT '预测模型版本',

    -- CLIP 评分（可选）
    clip_score DECIMAL(4,3) COMMENT 'CLIP图文匹配分',
    aesthetic_score DECIMAL(4,3) COMMENT '美学评分',

    -- 安全检测
    nsfw_score DECIMAL(4,3) COMMENT 'NSFW检测分数',
    is_safe BOOLEAN DEFAULT TRUE COMMENT '是否安全',

    scored_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE,

    INDEX idx_quality (quality_overall),
    INDEX idx_ctr (ctr_prediction)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创意评分表';

-- ============================================
-- 4. 实际投放性能数据
-- ============================================

-- 创意投放表现表
CREATE TABLE creative_performance (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    creative_id BIGINT UNSIGNED NOT NULL,

    -- 投放平台
    platform ENUM('facebook', 'instagram', 'tiktok', 'google', 'other') NOT NULL,
    campaign_id VARCHAR(128) COMMENT '广告活动ID',
    ad_set_id VARCHAR(128) COMMENT '广告组ID',
    ad_id VARCHAR(128) COMMENT '广告ID',

    -- 日期
    date DATE NOT NULL COMMENT '数据日期',

    -- 核心指标
    impressions INT DEFAULT 0 COMMENT '曝光量',
    clicks INT DEFAULT 0 COMMENT '点击量',
    conversions INT DEFAULT 0 COMMENT '转化量',
    spend DECIMAL(10,2) DEFAULT 0 COMMENT '花费金额',
    revenue DECIMAL(10,2) DEFAULT 0 COMMENT '收入金额',

    -- 计算指标
    ctr DECIMAL(6,4) COMMENT '点击率 = clicks/impressions',
    cvr DECIMAL(6,4) COMMENT '转化率 = conversions/clicks',
    cpc DECIMAL(8,4) COMMENT '单次点击成本',
    cpa DECIMAL(10,4) COMMENT '单次转化成本',
    roas DECIMAL(8,4) COMMENT '广告支出回报率',

    -- 元数据
    metadata JSON COMMENT '其他平台特定数据',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE,

    UNIQUE KEY uk_creative_platform_date (creative_id, platform, date),
    INDEX idx_date (date),
    INDEX idx_platform (platform),
    INDEX idx_ctr (ctr),
    INDEX idx_cvr (cvr)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创意投放表现表';

-- 性能数据汇总表（按周/月）
CREATE TABLE creative_performance_summary (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    creative_id BIGINT UNSIGNED NOT NULL,
    period_type ENUM('week', 'month') NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    total_impressions BIGINT DEFAULT 0,
    total_clicks INT DEFAULT 0,
    total_conversions INT DEFAULT 0,
    total_spend DECIMAL(12,2) DEFAULT 0,
    total_revenue DECIMAL(12,2) DEFAULT 0,

    avg_ctr DECIMAL(6,4),
    avg_cvr DECIMAL(6,4),
    avg_cpc DECIMAL(8,4),
    avg_roas DECIMAL(8,4),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE,

    UNIQUE KEY uk_creative_period (creative_id, period_type, period_start),
    INDEX idx_period (period_start, period_end)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创意性能汇总表';

-- ============================================
-- 5. A/B 测试管理
-- ============================================

-- A/B 实验表
CREATE TABLE ab_experiments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    project_id BIGINT UNSIGNED COMMENT '所属项目',
    creator_id BIGINT UNSIGNED NOT NULL COMMENT '创建者',

    name VARCHAR(255) NOT NULL COMMENT '实验名称',
    description TEXT COMMENT '实验描述',
    hypothesis TEXT COMMENT '实验假设',

    -- 实验配置
    traffic_allocation DECIMAL(3,2) DEFAULT 1.00 COMMENT '流量分配比例',
    config JSON COMMENT '实验配置',

    -- 实验状态
    status ENUM('draft', 'running', 'paused', 'completed', 'cancelled') DEFAULT 'draft',

    -- 时间范围
    start_date DATE COMMENT '开始日期',
    end_date DATE COMMENT '结束日期',
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,

    -- 结果
    winner_variant_id BIGINT UNSIGNED COMMENT '获胜变体ID',
    confidence_level DECIMAL(4,3) COMMENT '置信度',
    result_summary JSON COMMENT '结果摘要',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE,

    INDEX idx_project (project_id),
    INDEX idx_status (status),
    INDEX idx_dates (start_date, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='A/B实验表';

-- A/B 实验变体表
CREATE TABLE ab_variants (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    experiment_id BIGINT UNSIGNED NOT NULL,
    creative_id BIGINT UNSIGNED NOT NULL,

    variant_name VARCHAR(64) NOT NULL COMMENT '变体名称 A, B, C',
    traffic_allocation DECIMAL(4,3) NOT NULL COMMENT '流量分配 0-1',
    is_control BOOLEAN DEFAULT FALSE COMMENT '是否为对照组',

    -- 性能汇总
    total_impressions BIGINT DEFAULT 0,
    total_clicks INT DEFAULT 0,
    total_conversions INT DEFAULT 0,
    avg_ctr DECIMAL(6,4),
    avg_cvr DECIMAL(6,4),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (experiment_id) REFERENCES ab_experiments(id) ON DELETE CASCADE,
    FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE,

    UNIQUE KEY uk_experiment_creative (experiment_id, creative_id),
    INDEX idx_variant_name (variant_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='A/B实验变体表';

-- ============================================
-- 6. 创意模板管理
-- ============================================

-- 创意模板表
CREATE TABLE creative_templates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    name VARCHAR(128) NOT NULL COMMENT '模板名称',
    category VARCHAR(64) COMMENT '类别：电商、游戏、金融',

    -- 模板配置
    layout_config JSON COMMENT '布局配置',
    style_config JSON COMMENT '样式配置',
    default_params JSON COMMENT '默认参数',

    -- 预览
    preview_image_url VARCHAR(512) COMMENT '预览图',

    -- 使用统计
    usage_count INT DEFAULT 0 COMMENT '使用次数',
    avg_ctr DECIMAL(6,4) COMMENT '平均CTR',

    status ENUM('active', 'archived') DEFAULT 'active',
    is_public BOOLEAN DEFAULT FALSE COMMENT '是否公开',
    creator_id BIGINT UNSIGNED COMMENT '创建者',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE SET NULL,

    INDEX idx_category (category),
    INDEX idx_status (status),
    INDEX idx_usage (usage_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创意模板表';

-- ============================================
-- 7. 系统配置与额度管理
-- ============================================

-- 用户配额表
CREATE TABLE user_quotas (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED UNIQUE NOT NULL,

    -- 配额限制
    max_tasks_per_day INT DEFAULT 100 COMMENT '每日任务数上限',
    max_assets_total INT DEFAULT 10000 COMMENT '总素材数上限',

    -- 当前使用量
    tasks_today INT DEFAULT 0 COMMENT '今日已用任务数',
    assets_total INT DEFAULT 0 COMMENT '总素材数',

    -- 统计
    total_tasks_created INT DEFAULT 0 COMMENT '累计创建任务数',
    last_reset_at DATE COMMENT '最后重置日期',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户配额表';

-- API 密钥表
CREATE TABLE api_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,

    key_hash VARCHAR(255) UNIQUE NOT NULL COMMENT 'API Key哈希',
    key_prefix VARCHAR(16) NOT NULL COMMENT 'Key前缀（用于显示）',
    name VARCHAR(128) COMMENT 'Key名称',

    permissions JSON COMMENT '权限列表',

    status ENUM('active', 'revoked') DEFAULT 'active',
    last_used_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL COMMENT '过期时间',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    INDEX idx_user (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API密钥表';

-- ============================================
-- 8. 审计日志与监控
-- ============================================

-- 操作审计日志
CREATE TABLE audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED COMMENT '操作用户',

    action VARCHAR(128) NOT NULL COMMENT '操作类型',
    resource_type VARCHAR(64) COMMENT '资源类型',
    resource_id VARCHAR(128) COMMENT '资源ID',

    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT 'User Agent',

    request_params JSON COMMENT '请求参数',
    response_status INT COMMENT '响应状态码',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,

    INDEX idx_user (user_id),
    INDEX idx_action (action),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审计日志表';

-- 系统任务日志（后台任务）
CREATE TABLE system_task_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    task_type VARCHAR(64) NOT NULL COMMENT '任务类型: ctr_model_training, cleanup',
    status ENUM('pending', 'running', 'success', 'failed') DEFAULT 'pending',

    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    duration INT COMMENT '执行时长（秒）',

    params JSON COMMENT '任务参数',
    result JSON COMMENT '执行结果',
    error_message TEXT COMMENT '错误信息',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_type (task_type),
    INDEX idx_status (status),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统任务日志表';

-- ============================================
-- 9. 创意文案库
-- ============================================

-- 文案库表
CREATE TABLE copy_library (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED COMMENT '创建者',

    category ENUM('headline', 'subheadline', 'cta', 'description') NOT NULL,
    text TEXT NOT NULL COMMENT '文案内容',

    industry VARCHAR(64) COMMENT '行业',
    tags JSON COMMENT '标签',

    -- 性能数据
    usage_count INT DEFAULT 0 COMMENT '使用次数',
    avg_ctr DECIMAL(6,4) COMMENT '平均CTR',

    status ENUM('active', 'archived') DEFAULT 'active',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,

    INDEX idx_category (category),
    INDEX idx_industry (industry),
    INDEX idx_avg_ctr (avg_ctr)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文案库表';

-- ============================================
-- 10. 标签与分类
-- ============================================

-- 标签表
CREATE TABLE tags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) UNIQUE NOT NULL,
    category VARCHAR(64) COMMENT '标签分类',
    color VARCHAR(7) COMMENT '颜色代码',
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='标签表';

-- 创意标签关联表
CREATE TABLE creative_tags (
    creative_id BIGINT UNSIGNED NOT NULL,
    tag_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (creative_id, tag_id),
    FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,

    INDEX idx_tag (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创意标签关联表';

-- ============================================
-- 11. 创建初始数据
-- ============================================

-- 插入默认管理员账户
INSERT INTO users (uuid, username, email, password_hash, role, status)
VALUES (
    UUID(),
    'admin',
    'admin@example.com',
    -- 密码: admin123 (请在实际使用时修改)
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'admin',
    'active'
);

-- 插入默认标签
INSERT INTO tags (name, category, color) VALUES
    ('电商', 'industry', '#FF6B6B'),
    ('游戏', 'industry', '#4ECDC4'),
    ('金融', 'industry', '#45B7D1'),
    ('教育', 'industry', '#FFA07A'),
    ('极简风', 'style', '#95E1D3'),
    ('活力风', 'style', '#F38181'),
    ('专业风', 'style', '#AA96DA');

-- 插入默认创意模板
INSERT INTO creative_templates (uuid, name, category, status, is_public) VALUES
    (UUID(), '电商促销模板', '电商', 'active', true),
    (UUID(), '游戏推广模板', '游戏', 'active', true),
    (UUID(), '金融产品模板', '金融', 'active', true);
