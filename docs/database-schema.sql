-- ============================================
-- AI 多尺寸广告创意生成平台 - PostgreSQL 数据库设计
-- Version: 1.0
-- ============================================

-- ============================================
-- 1. 用户与权限管理
-- ============================================

-- 用户表
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL, -- UUID
    username VARCHAR(64) UNIQUE NOT NULL, -- 用户名
    email VARCHAR(255) UNIQUE NOT NULL, -- 邮箱
    password_hash VARCHAR(255) NOT NULL, -- 密码哈希
    phone VARCHAR(20), -- 手机号
    avatar_url VARCHAR(512), -- 头像URL
    role VARCHAR(20) DEFAULT 'user', -- 角色 (admin/user/viewer)
    status VARCHAR(20) DEFAULT 'active', -- 状态 (active/inactive/banned)
    last_login_at TIMESTAMP NULL, -- 最后登录时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL -- 软删除时间
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- 团队/项目表
CREATE TABLE projects (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL, -- UUID
    name VARCHAR(128) NOT NULL, -- 项目名称
    description TEXT, -- 项目描述
    owner_id BIGINT NOT NULL, -- 所有者ID
    status VARCHAR(20) DEFAULT 'active', -- 状态 (active/archived)
    settings JSONB, -- 项目设置（Logo、默认风格等）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_projects_owner ON projects(owner_id);
CREATE INDEX idx_projects_status ON projects(status);

ALTER TABLE projects ADD CONSTRAINT fk_projects_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;

-- 项目成员表
CREATE TABLE project_members (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(20) DEFAULT 'member', -- 角色 (owner/admin/member/viewer)
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_project_members_user ON project_members(user_id);
CREATE UNIQUE INDEX uk_project_user ON project_members(project_id, user_id);

ALTER TABLE project_members ADD CONSTRAINT fk_project_members_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE project_members ADD CONSTRAINT fk_project_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ============================================
-- 2. 创意生成核心表
-- ============================================

-- 创意生成任务表
CREATE TABLE creative_tasks (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL, -- 任务UUID
    user_id BIGINT NOT NULL, -- 创建用户
    project_id BIGINT, -- 所属项目

    -- 输入信息
    title VARCHAR(255) NOT NULL, -- 商品标题
    selling_points JSONB, -- 卖点列表
    product_image_url VARCHAR(512), -- 商品图URL
    brand_logo_url VARCHAR(512), -- 品牌Logo URL

    -- 生成配置
    requested_formats JSONB, -- 请求的尺寸格式 ["1:1", "4:5", "9:16"]
    requested_styles JSONB, -- 请求的风格 ["modern", "elegant", "vibrant"]
    num_variants INTEGER DEFAULT 3, -- 每种风格的变体数
    cta_text VARCHAR(64), -- CTA文案

    -- 任务状态
    status VARCHAR(20) DEFAULT 'pending', -- 状态 (pending/queued/processing/completed/failed/cancelled)
    progress SMALLINT DEFAULT 0, -- 进度 0-100
    error_message TEXT, -- 错误信息

    -- 时间统计
    queued_at TIMESTAMP NULL, -- 入队时间
    started_at TIMESTAMP NULL, -- 开始处理时间
    completed_at TIMESTAMP NULL, -- 完成时间
    processing_duration INTEGER, -- 处理耗时（秒）

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_creative_tasks_user ON creative_tasks(user_id);
CREATE INDEX idx_creative_tasks_project ON creative_tasks(project_id);
CREATE INDEX idx_creative_tasks_status ON creative_tasks(status);
CREATE INDEX idx_creative_tasks_created ON creative_tasks(created_at);

ALTER TABLE creative_tasks ADD CONSTRAINT fk_creative_tasks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE creative_tasks ADD CONSTRAINT fk_creative_tasks_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;

-- 创意素材表
CREATE TABLE creative_assets (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL, -- 素材UUID
    task_id BIGINT NOT NULL, -- 任务ID

    -- 素材信息
    format VARCHAR(20) NOT NULL, -- 尺寸格式: 1:1, 4:5, 9:16, 1200x628
    width INTEGER NOT NULL, -- 宽度（像素）
    height INTEGER NOT NULL, -- 高度（像素）
    file_size INTEGER, -- 文件大小（字节）

    -- 存储信息
    storage_type VARCHAR(20) DEFAULT 'local', -- 存储类型 (local/oss/s3/minio)
    file_path VARCHAR(512) NOT NULL, -- 文件路径/Key
    public_url VARCHAR(1024), -- 公开访问URL
    cdn_url VARCHAR(1024), -- CDN URL

    -- 生成元数据
    style VARCHAR(64), -- 风格标签
    variant_index INTEGER, -- 变体索引
    generation_prompt TEXT, -- 生成提示词
    model_name VARCHAR(128), -- 使用的模型名称
    model_version VARCHAR(64), -- 模型版本
    generation_params JSONB, -- 生成参数

    -- 内容信息
    text_content JSONB, -- 文本内容 {title, subtitle, cta}
    has_logo BOOLEAN DEFAULT FALSE, -- 是否包含Logo
    has_cta BOOLEAN DEFAULT FALSE, -- 是否包含CTA

    -- 排序
    rank INTEGER, -- 排名（基于CTR预测）

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_creative_assets_task ON creative_assets(task_id);
CREATE INDEX idx_creative_assets_format ON creative_assets(format);
CREATE INDEX idx_creative_assets_style ON creative_assets(style);
CREATE INDEX idx_creative_assets_rank ON creative_assets(rank);
CREATE INDEX idx_creative_assets_created ON creative_assets(created_at);

ALTER TABLE creative_assets ADD CONSTRAINT fk_creative_assets_task FOREIGN KEY (task_id) REFERENCES creative_tasks(id) ON DELETE CASCADE;

-- ============================================
-- 3. 质量评估与评分
-- ============================================

-- 创意评分表
CREATE TABLE creative_scores (
    id BIGSERIAL PRIMARY KEY,
    creative_id BIGINT UNIQUE NOT NULL,

    -- 质量评分（0-1）
    quality_overall NUMERIC(4,3), -- 综合质量评分
    brightness_score NUMERIC(4,3), -- 亮度评分
    contrast_score NUMERIC(4,3), -- 对比度评分
    sharpness_score NUMERIC(4,3), -- 清晰度评分
    composition_score NUMERIC(4,3), -- 构图评分
    color_harmony_score NUMERIC(4,3), -- 配色和谐度

    -- CTR 预测
    ctr_prediction NUMERIC(5,4), -- CTR预测值
    ctr_confidence NUMERIC(4,3), -- 预测置信度
    model_version VARCHAR(64), -- 预测模型版本

    -- CLIP 评分（可选）
    clip_score NUMERIC(4,3), -- CLIP图文匹配分
    aesthetic_score NUMERIC(4,3), -- 美学评分

    -- 安全检测
    nsfw_score NUMERIC(4,3), -- NSFW检测分数
    is_safe BOOLEAN DEFAULT TRUE, -- 是否安全

    scored_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_creative_scores_quality ON creative_scores(quality_overall);
CREATE INDEX idx_creative_scores_ctr ON creative_scores(ctr_prediction);

ALTER TABLE creative_scores ADD CONSTRAINT fk_creative_scores_creative FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE;

-- ============================================
-- 4. 实际投放性能数据
-- ============================================

-- 创意投放表现表
CREATE TABLE creative_performance (
    id BIGSERIAL PRIMARY KEY,
    creative_id BIGINT NOT NULL,

    -- 投放平台
    platform VARCHAR(20) NOT NULL, -- 平台 (facebook/instagram/tiktok/google/other)
    campaign_id VARCHAR(128), -- 广告活动ID
    ad_set_id VARCHAR(128), -- 广告组ID
    ad_id VARCHAR(128), -- 广告ID

    -- 日期
    date DATE NOT NULL, -- 数据日期

    -- 核心指标
    impressions INTEGER DEFAULT 0, -- 曝光量
    clicks INTEGER DEFAULT 0, -- 点击量
    conversions INTEGER DEFAULT 0, -- 转化量
    spend NUMERIC(10,2) DEFAULT 0, -- 花费金额
    revenue NUMERIC(10,2) DEFAULT 0, -- 收入金额

    -- 计算指标
    ctr NUMERIC(6,4), -- 点击率 = clicks/impressions
    cvr NUMERIC(6,4), -- 转化率 = conversions/clicks
    cpc NUMERIC(8,4), -- 单次点击成本
    cpa NUMERIC(10,4), -- 单次转化成本
    roas NUMERIC(8,4), -- 广告支出回报率

    -- 元数据
    metadata JSONB, -- 其他平台特定数据

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uk_creative_platform_date ON creative_performance(creative_id, platform, date);
CREATE INDEX idx_creative_performance_date ON creative_performance(date);
CREATE INDEX idx_creative_performance_platform ON creative_performance(platform);
CREATE INDEX idx_creative_performance_ctr ON creative_performance(ctr);
CREATE INDEX idx_creative_performance_cvr ON creative_performance(cvr);

ALTER TABLE creative_performance ADD CONSTRAINT fk_creative_performance_creative FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE;

-- 性能数据汇总表（按周/月）
CREATE TABLE creative_performance_summary (
    id BIGSERIAL PRIMARY KEY,
    creative_id BIGINT NOT NULL,
    period_type VARCHAR(10) NOT NULL, -- 周期类型 (week/month)
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    total_impressions BIGINT DEFAULT 0,
    total_clicks INTEGER DEFAULT 0,
    total_conversions INTEGER DEFAULT 0,
    total_spend NUMERIC(12,2) DEFAULT 0,
    total_revenue NUMERIC(12,2) DEFAULT 0,

    avg_ctr NUMERIC(6,4),
    avg_cvr NUMERIC(6,4),
    avg_cpc NUMERIC(8,4),
    avg_roas NUMERIC(8,4),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uk_creative_period ON creative_performance_summary(creative_id, period_type, period_start);
CREATE INDEX idx_creative_performance_summary_period ON creative_performance_summary(period_start, period_end);

ALTER TABLE creative_performance_summary ADD CONSTRAINT fk_creative_performance_summary_creative FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE;

-- ============================================
-- 5. A/B 测试管理
-- ============================================

-- A/B 实验表
CREATE TABLE ab_experiments (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    project_id BIGINT, -- 所属项目
    creator_id BIGINT NOT NULL, -- 创建者

    name VARCHAR(255) NOT NULL, -- 实验名称
    description TEXT, -- 实验描述
    hypothesis TEXT, -- 实验假设

    -- 实验配置
    traffic_allocation NUMERIC(3,2) DEFAULT 1.00, -- 流量分配比例
    config JSONB, -- 实验配置

    -- 实验状态
    status VARCHAR(20) DEFAULT 'draft', -- 状态 (draft/running/paused/completed/cancelled)

    -- 时间范围
    start_date DATE, -- 开始日期
    end_date DATE, -- 结束日期
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,

    -- 结果
    winner_variant_id BIGINT, -- 获胜变体ID
    confidence_level NUMERIC(4,3), -- 置信度
    result_summary JSONB, -- 结果摘要

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_ab_experiments_project ON ab_experiments(project_id);
CREATE INDEX idx_ab_experiments_status ON ab_experiments(status);
CREATE INDEX idx_ab_experiments_dates ON ab_experiments(start_date, end_date);

ALTER TABLE ab_experiments ADD CONSTRAINT fk_ab_experiments_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE ab_experiments ADD CONSTRAINT fk_ab_experiments_creator FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE;

-- A/B 实验变体表
CREATE TABLE ab_variants (
    id BIGSERIAL PRIMARY KEY,
    experiment_id BIGINT NOT NULL,
    creative_id BIGINT NOT NULL,

    variant_name VARCHAR(64) NOT NULL, -- 变体名称 A, B, C
    traffic_allocation NUMERIC(4,3) NOT NULL, -- 流量分配 0-1
    is_control BOOLEAN DEFAULT FALSE, -- 是否为对照组

    -- 性能汇总
    total_impressions BIGINT DEFAULT 0,
    total_clicks INTEGER DEFAULT 0,
    total_conversions INTEGER DEFAULT 0,
    avg_ctr NUMERIC(6,4),
    avg_cvr NUMERIC(6,4),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uk_experiment_creative ON ab_variants(experiment_id, creative_id);
CREATE INDEX idx_ab_variants_variant_name ON ab_variants(variant_name);

ALTER TABLE ab_variants ADD CONSTRAINT fk_ab_variants_experiment FOREIGN KEY (experiment_id) REFERENCES ab_experiments(id) ON DELETE CASCADE;
ALTER TABLE ab_variants ADD CONSTRAINT fk_ab_variants_creative FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE;

-- ============================================
-- 6. 创意模板管理
-- ============================================

-- 创意模板表
CREATE TABLE creative_templates (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    name VARCHAR(128) NOT NULL, -- 模板名称
    category VARCHAR(64), -- 类别：电商、游戏、金融

    -- 模板配置
    layout_config JSONB, -- 布局配置
    style_config JSONB, -- 样式配置
    default_params JSONB, -- 默认参数

    -- 预览
    preview_image_url VARCHAR(512), -- 预览图

    -- 使用统计
    usage_count INTEGER DEFAULT 0, -- 使用次数
    avg_ctr NUMERIC(6,4), -- 平均CTR

    status VARCHAR(20) DEFAULT 'active', -- 状态 (active/archived)
    is_public BOOLEAN DEFAULT FALSE, -- 是否公开
    creator_id BIGINT, -- 创建者

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_creative_templates_category ON creative_templates(category);
CREATE INDEX idx_creative_templates_status ON creative_templates(status);
CREATE INDEX idx_creative_templates_usage ON creative_templates(usage_count);

ALTER TABLE creative_templates ADD CONSTRAINT fk_creative_templates_creator FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE SET NULL;

-- ============================================
-- 7. 系统配置与额度管理
-- ============================================

-- 用户配额表
CREATE TABLE user_quotas (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,

    -- 配额限制
    max_tasks_per_day INTEGER DEFAULT 100, -- 每日任务数上限
    max_assets_total INTEGER DEFAULT 10000, -- 总素材数上限

    -- 当前使用量
    tasks_today INTEGER DEFAULT 0, -- 今日已用任务数
    assets_total INTEGER DEFAULT 0, -- 总素材数

    -- 统计
    total_tasks_created INTEGER DEFAULT 0, -- 累计创建任务数
    last_reset_at DATE, -- 最后重置日期

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE user_quotas ADD CONSTRAINT fk_user_quotas_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- API 密钥表
CREATE TABLE api_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,

    key_hash VARCHAR(255) UNIQUE NOT NULL, -- API Key哈希
    key_prefix VARCHAR(16) NOT NULL, -- Key前缀（用于显示）
    name VARCHAR(128), -- Key名称

    permissions JSONB, -- 权限列表

    status VARCHAR(20) DEFAULT 'active', -- 状态 (active/revoked)
    last_used_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL, -- 过期时间

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_status ON api_keys(status);

ALTER TABLE api_keys ADD CONSTRAINT fk_api_keys_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ============================================
-- 8. 审计日志与监控
-- ============================================

-- 操作审计日志
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT, -- 操作用户

    action VARCHAR(128) NOT NULL, -- 操作类型
    resource_type VARCHAR(64), -- 资源类型
    resource_id VARCHAR(128), -- 资源ID

    ip_address INET, -- IP地址 (PostgreSQL's INET type is better for IP addresses)
    user_agent TEXT, -- User Agent

    request_params JSONB, -- 请求参数
    response_status INTEGER, -- 响应状态码

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);

ALTER TABLE audit_logs ADD CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- 系统任务日志（后台任务）
CREATE TABLE system_task_logs (
    id BIGSERIAL PRIMARY KEY,
    task_type VARCHAR(64) NOT NULL, -- 任务类型: ctr_model_training, cleanup
    status VARCHAR(20) DEFAULT 'pending', -- 状态 (pending/running/success/failed)

    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    duration INTEGER, -- 执行时长（秒）

    params JSONB, -- 任务参数
    result JSONB, -- 执行结果
    error_message TEXT, -- 错误信息

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_system_task_logs_type ON system_task_logs(task_type);
CREATE INDEX idx_system_task_logs_status ON system_task_logs(status);
CREATE INDEX idx_system_task_logs_created ON system_task_logs(created_at);

-- ============================================
-- 9. 创意文案库
-- ============================================

-- 文案库表
CREATE TABLE copy_library (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT, -- 创建者

    category VARCHAR(20) NOT NULL, -- 类别 (headline/subheadline/cta/description)
    text TEXT NOT NULL, -- 文案内容

    industry VARCHAR(64), -- 行业
    tags JSONB, -- 标签

    -- 性能数据
    usage_count INTEGER DEFAULT 0, -- 使用次数
    avg_ctr NUMERIC(6,4), -- 平均CTR

    status VARCHAR(20) DEFAULT 'active', -- 状态 (active/archived)

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_copy_library_category ON copy_library(category);
CREATE INDEX idx_copy_library_industry ON copy_library(industry);
CREATE INDEX idx_copy_library_avg_ctr ON copy_library(avg_ctr);

ALTER TABLE copy_library ADD CONSTRAINT fk_copy_library_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- ============================================
-- 10. 标签与分类
-- ============================================

-- 标签表
CREATE TABLE tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(64) UNIQUE NOT NULL,
    category VARCHAR(64), -- 标签分类
    color VARCHAR(7), -- 颜色代码
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创意标签关联表
CREATE TABLE creative_tags (
    creative_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (creative_id, tag_id)
);

CREATE INDEX idx_creative_tags_tag ON creative_tags(tag_id);

ALTER TABLE creative_tags ADD CONSTRAINT fk_creative_tags_creative FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE;
ALTER TABLE creative_tags ADD CONSTRAINT fk_creative_tags_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE;

-- ============================================
-- 11. 创建初始数据
-- ============================================

-- 插入默认管理员账户
INSERT INTO users (uuid, username, email, password_hash, role, status)
VALUES (
    gen_random_uuid(),
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
    (gen_random_uuid(), '电商促销模板', '电商', 'active', true),
    (gen_random_uuid(), '游戏推广模板', '游戏', 'active', true),
    (gen_random_uuid(), '金融产品模板', '金融', 'active', true);
