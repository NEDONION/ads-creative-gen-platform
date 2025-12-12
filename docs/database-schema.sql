-- AI 广告创意生成平台 - 数据库结构定义
-- 清空所有表
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    username VARCHAR(64) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    avatar_url VARCHAR(512),
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    last_login_at TIMESTAMPTZ
);

-- 标签表
CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    name VARCHAR(64) NOT NULL UNIQUE,
    category VARCHAR(64),
    color VARCHAR(7),
    usage_count INTEGER DEFAULT 0
);

-- 项目表
CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    owner_id BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    settings JSONB,
    CONSTRAINT fk_projects_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 项目成员表
CREATE TABLE IF NOT EXISTS project_members (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    project_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(20) DEFAULT 'member',
    CONSTRAINT fk_projects_members FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_project_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(project_id, user_id)
);

-- 创意任务表
CREATE TABLE IF NOT EXISTS creative_tasks (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    project_id BIGINT,
    title VARCHAR(255) NOT NULL,
    selling_points JSONB,
    product_image_url VARCHAR(512),
    brand_logo_url VARCHAR(512),
    requested_formats JSONB,
    requested_styles JSONB,
    num_variants INTEGER DEFAULT 3,
    cta_text VARCHAR(64),
    status VARCHAR(20) DEFAULT 'pending',
    progress INTEGER DEFAULT 0,
    error_message TEXT,
    queued_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    processing_duration INTEGER,
    CONSTRAINT fk_creative_tasks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_creative_tasks_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
);

-- 创意素材表
CREATE TABLE IF NOT EXISTS creative_assets (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    task_id BIGINT NOT NULL,
    format VARCHAR(20) NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    file_size INTEGER,
    storage_type VARCHAR(20) NOT NULL DEFAULT 'local',
    public_url VARCHAR(1024) NOT NULL,
    original_path VARCHAR(512),
    style VARCHAR(64),
    variant_index INTEGER,
    generation_prompt TEXT,
    model_name VARCHAR(128),
    model_version VARCHAR(64),
    generation_params JSONB,
    text_content JSONB,
    has_logo BOOLEAN DEFAULT FALSE,
    has_cta BOOLEAN DEFAULT FALSE,
    rank INTEGER,
    CONSTRAINT fk_creative_tasks_assets FOREIGN KEY (task_id) REFERENCES creative_tasks(id) ON DELETE CASCADE
);

-- 创意评分表
CREATE TABLE IF NOT EXISTS creative_scores (
    id BIGSERIAL PRIMARY KEY,
    creative_id BIGINT NOT NULL UNIQUE,
    quality_overall DECIMAL(4,3),
    brightness_score DECIMAL(4,3),
    contrast_score DECIMAL(4,3),
    sharpness_score DECIMAL(4,3),
    composition_score DECIMAL(4,3),
    color_harmony_score DECIMAL(4,3),
    ctr_prediction DECIMAL(5,4),
    ctr_confidence DECIMAL(4,3),
    model_version VARCHAR(64),
    cl_ip_score DECIMAL(4,3),
    aesthetic_score DECIMAL(4,3),
    nsfw_score DECIMAL(4,3),
    is_safe BOOLEAN DEFAULT TRUE,
    scored_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_creative_assets_score FOREIGN KEY (creative_id) REFERENCES creative_assets(id) ON DELETE CASCADE
);

-- 创意素材与标签关联表
CREATE TABLE IF NOT EXISTS creative_tags (
    creative_asset_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    PRIMARY KEY (creative_asset_id, tag_id),
    CONSTRAINT fk_creative_tags_creative_asset FOREIGN KEY (creative_asset_id) REFERENCES creative_assets(id) ON DELETE CASCADE,
    CONSTRAINT fk_creative_tags_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_uuid ON users(uuid);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

CREATE INDEX IF NOT EXISTS idx_tags_deleted_at ON tags(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_category ON tags(category);

CREATE INDEX IF NOT EXISTS idx_projects_uuid ON projects(uuid);
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);
CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

CREATE INDEX IF NOT EXISTS idx_project_members_deleted_at ON project_members(deleted_at);
CREATE INDEX IF NOT EXISTS idx_project_user ON project_members(project_id, user_id);

CREATE INDEX IF NOT EXISTS idx_creative_tasks_uuid ON creative_tasks(uuid);
CREATE INDEX IF NOT EXISTS idx_creative_tasks_deleted_at ON creative_tasks(deleted_at);
CREATE INDEX IF NOT EXISTS idx_creative_tasks_user_id ON creative_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_creative_tasks_project_id ON creative_tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_creative_tasks_status ON creative_tasks(status);

CREATE INDEX IF NOT EXISTS idx_creative_assets_uuid ON creative_assets(uuid);
CREATE INDEX IF NOT EXISTS idx_creative_assets_deleted_at ON creative_assets(deleted_at);
CREATE INDEX IF NOT EXISTS idx_creative_assets_task_id ON creative_assets(task_id);
CREATE INDEX IF NOT EXISTS idx_creative_assets_format ON creative_assets(format);
CREATE INDEX IF NOT EXISTS idx_creative_assets_style ON creative_assets(style);
CREATE INDEX IF NOT EXISTS idx_creative_assets_rank ON creative_assets(rank);

CREATE INDEX IF NOT EXISTS idx_creative_scores_creative_id ON creative_scores(creative_id);
CREATE INDEX IF NOT EXISTS idx_creative_scores_ctr_prediction ON creative_scores(ctr_prediction);

-- 为频繁查询的字段创建复合索引
CREATE INDEX IF NOT EXISTS idx_creative_tasks_user_status ON creative_tasks(user_id, status);
CREATE INDEX IF NOT EXISTS idx_creative_assets_task_format ON creative_assets(task_id, format);

-- 触发器：自动更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要自动更新的表添加触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_tags_updated_at ON tags;
CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_project_members_updated_at ON project_members;
CREATE TRIGGER update_project_members_updated_at BEFORE UPDATE ON project_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_creative_tasks_updated_at ON creative_tasks;
CREATE TRIGGER update_creative_tasks_updated_at BEFORE UPDATE ON creative_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_creative_assets_updated_at ON creative_assets;
CREATE TRIGGER update_creative_assets_updated_at BEFORE UPDATE ON creative_assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 为创意评分表设置特殊更新逻辑
CREATE OR REPLACE FUNCTION update_creative_score_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.scored_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_creative_scores_updated_at ON creative_scores;
CREATE TRIGGER update_creative_scores_updated_at BEFORE INSERT OR UPDATE ON creative_scores
    FOR EACH ROW EXECUTE FUNCTION update_creative_score_updated_at();