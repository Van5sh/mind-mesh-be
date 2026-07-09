-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TYPE project_role AS ENUM (
    'OWNER',
    'ADMIN',
    'EDITOR',
    'VIEWER'
);

CREATE TYPE file_permission AS ENUM (
    'READ',
    'WRITE'
);

CREATE TYPE message_role AS ENUM (
    'USER',
    'AI',
    'SYSTEM'
);

CREATE TYPE project_visibility AS ENUM (
    'PRIVATE',
    'TEAM'
);

CREATE TYPE report_status AS ENUM (
    'DRAFT',
    'GENERATING',
    'READY',
    'FAILED'
);

CREATE TYPE flowchart_status AS ENUM (
    'DRAFT',
    'GENERATING',
    'READY',
    'FAILED'
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    visibility project_visibility NOT NULL DEFAULT 'PRIVATE',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    role project_role NOT NULL DEFAULT 'EDITOR',

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(project_id, user_id)
);

CREATE TABLE folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    parent_folder_id UUID REFERENCES folders(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(project_id, parent_folder_id, name)
);

CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    folder_id UUID REFERENCES folders(id) ON DELETE SET NULL,

    uploaded_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    name VARCHAR(255) NOT NULL,

    storage_path TEXT NOT NULL,

    mime_type VARCHAR(100) NOT NULL,

    size BIGINT NOT NULL CHECK(size >= 0),

    original_name VARCHAR(255),

    extracted_text TEXT,

    is_indexed BOOLEAN DEFAULT FALSE,

    is_favorite BOOLEAN DEFAULT FALSE,

    deleted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),

    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(project_id, folder_id, name)
);

CREATE TABLE file_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,

    shared_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    shared_with UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    permission file_permission NOT NULL DEFAULT 'READ',

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(file_id, shared_with),

    CHECK (shared_by <> shared_with)
);

CREATE TABLE chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    title VARCHAR(100),

    last_activity_at TIMESTAMPTZ DEFAULT NOW(),

    created_at TIMESTAMPTZ DEFAULT NOW(),

    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    role message_role NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE flowcharts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    data JSONB NOT NULL,
    generated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    generated_by_ai BOOLEAN DEFAULT FALSE,
    status flowchart_status NOT NULL DEFAULT 'READY',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    generated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    generated_by_ai BOOLEAN DEFAULT FALSE,
    status report_status NOT NULL DEFAULT 'READY',
    format VARCHAR(20) DEFAULT 'MARKDOWN',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE file_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    embedding vector(768) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE chat_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    embedding vector(768) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_file_embeddings_vector
ON file_embeddings
USING hnsw (embedding vector_cosine_ops);

CREATE INDEX idx_chat_embeddings_vector
ON chat_embeddings
USING hnsw (embedding vector_cosine_ops);

CREATE INDEX idx_projects_owner
ON projects(owner_id);

CREATE INDEX idx_project_members_project
ON project_members(project_id);

CREATE INDEX idx_project_members_user
ON project_members(user_id);

CREATE INDEX idx_folders_project
ON folders(project_id);

CREATE INDEX idx_folders_parent
ON folders(parent_folder_id);

CREATE INDEX idx_files_project
ON files(project_id);

CREATE INDEX idx_files_folder
ON files(folder_id);

CREATE INDEX idx_files_uploaded_by
ON files(uploaded_by);

CREATE INDEX idx_files_deleted_at
ON files(deleted_at);

CREATE INDEX idx_files_favorite
ON files(is_favorite);

CREATE INDEX idx_files_indexed
ON files(is_indexed);

CREATE INDEX idx_file_shares_file
ON file_shares(file_id);

CREATE INDEX idx_file_shares_shared_with
ON file_shares(shared_with);

CREATE INDEX idx_chats_project
ON chats(project_id);

CREATE INDEX idx_chats_last_activity
ON chats(last_activity_at DESC);

CREATE INDEX idx_chat_messages_chat
ON chat_messages(chat_id);

CREATE INDEX idx_chat_messages_created_at
ON chat_messages(created_at);

CREATE INDEX idx_flowcharts_project
ON flowcharts(project_id);

CREATE INDEX idx_reports_project
ON reports(project_id);

CREATE INDEX idx_activity_logs_project
ON activity_logs(project_id);

CREATE INDEX idx_activity_logs_user
ON activity_logs(user_id);

CREATE INDEX idx_activity_logs_created_at
ON activity_logs(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS chat_embeddings CASCADE;
DROP TABLE IF EXISTS file_embeddings CASCADE;
DROP TABLE IF EXISTS activity_logs CASCADE;
DROP TABLE IF EXISTS reports CASCADE;
DROP TABLE IF EXISTS flowcharts CASCADE;
DROP TABLE IF EXISTS chat_messages CASCADE;
DROP TABLE IF EXISTS chats CASCADE;
DROP TABLE IF EXISTS file_shares CASCADE;
DROP TABLE IF EXISTS files CASCADE;
DROP TABLE IF EXISTS folders CASCADE;
DROP TABLE IF EXISTS project_members CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS flowchart_status;
DROP TYPE IF EXISTS report_status;
DROP TYPE IF EXISTS project_visibility;
DROP TYPE IF EXISTS message_role;
DROP TYPE IF EXISTS file_permission;
DROP TYPE IF EXISTS project_role;
DROP EXTENSION IF EXISTS vector;
DROP EXTENSION IF EXISTS "pgcrypto";