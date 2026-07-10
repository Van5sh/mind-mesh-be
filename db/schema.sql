CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE TYPE project_role AS ENUM (
    'OWNER',
    'ADMIN',
    'EDITOR',
    'VIEWER'
);
CREATE TYPE report_format AS ENUM(
    'MARKDOWN',
    'PDF',
    'DOCX'
);
CREATE TYPE chat_status AS ENUM (
    'ACTIVE',
    'GENERATING',
    'ARCHIVED'
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
CREATE TYPE chat_type AS ENUM (
    'GENERAL',
    'AI_ASSISTANT'
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
CREATE TABLE user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    bio TEXT,
    avatar_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    visibility project_visibility NOT NULL DEFAULT 'PRIVATE',
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role project_role NOT NULL DEFAULT 'EDITOR',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, user_id)
);
CREATE TABLE folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    parent_folder_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CHECK (parent_folder_id IS NULL OR parent_folder_id <> id),
    UNIQUE(project_id, parent_folder_id, name)
);
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    folder_id UUID REFERENCES folders(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL CHECK(size >= 0),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, folder_id, name)
);
CREATE TABLE file_storage(
    file_id UUID PRIMARY KEY REFERENCES files(id) ON DELETE CASCADE,
    bucket_name VARCHAR(100) NOT NULL,
    object_key TEXT NOT NULL,
    etag VARCHAR(255),
    version_id VARCHAR(255),
    checksum VARCHAR(64),
    mime_type VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    uploaded_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(bucket_name, object_key)
);
CREATE TABLE file_properties (
    file_id UUID PRIMARY KEY REFERENCES files(id) ON DELETE CASCADE,
    original_name VARCHAR(255),
    is_indexed BOOLEAN NOT NULL DEFAULT FALSE,
    is_favorite BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE TABLE file_ai_metadata (
    file_id UUID PRIMARY KEY REFERENCES files(id) ON DELETE CASCADE,
    extracted_text TEXT,
    embedding_model VARCHAR(100),
    embedding_synced BOOLEAN DEFAULT FALSE,
    indexed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
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
    type chat_type NOT NULL DEFAULT 'GENERAL',
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),
    status chat_status NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id UUID REFERENCES users(id) ON DELETE SET NULL,
    role message_role NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE chat_ai_metadata (
    message_id UUID PRIMARY KEY REFERENCES chat_messages(id) ON DELETE CASCADE,
    embedding_model VARCHAR(100),
    embedding_synced BOOLEAN DEFAULT FALSE,
    indexed_at TIMESTAMPTZ
);

CREATE TABLE flowcharts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    data JSONB NOT NULL,
    generated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    generated_by_ai BOOLEAN DEFAULT FALSE,
    status flowchart_status NOT NULL DEFAULT 'READY',
    source_chat_id UUID REFERENCES chats(id) ON DELETE SET NULL,
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
    source_chat_id UUID REFERENCES chats(id) ON DELETE SET NULL,
    format report_format NOT NULL DEFAULT 'MARKDOWN',
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

CREATE INDEX idx_file_ai_metadata_synced
ON file_ai_metadata(embedding_synced);

CREATE INDEX idx_chat_ai_metadata_synced
ON chat_ai_metadata(embedding_synced);

CREATE INDEX idx_file_storage_uploaded_by
ON file_storage(uploaded_by);

CREATE INDEX idx_file_properties_deleted_at
ON file_properties(deleted_at);

CREATE INDEX idx_file_properties_favorite
ON file_properties(is_favorite);

CREATE INDEX idx_file_properties_indexed
ON file_properties(is_indexed);

CREATE INDEX idx_projects_visibility
ON projects(visibility);

CREATE INDEX idx_files_name
ON files(name);

CREATE INDEX idx_flowcharts_generated_by
ON flowcharts(generated_by);

CREATE INDEX idx_reports_generated_by
ON reports(generated_by);

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

CREATE INDEX idx_chat_messages_sender
ON chat_messages(sender_id);

CREATE INDEX idx_files_project_folder
ON files(project_id, folder_id);

CREATE INDEX idx_folders_project_parent
ON folders(project_id, parent_folder_id);

CREATE INDEX idx_chat_messages_chat_sender
ON chat_messages(chat_id, sender_id);

CREATE INDEX idx_projects_archived
ON projects(archived_at);

DROP TABLE IF EXISTS activity_logs CASCADE;
DROP TABLE IF EXISTS reports CASCADE;
DROP TABLE IF EXISTS flowcharts CASCADE;
DROP TABLE IF EXISTS chat_ai_metadata CASCADE;
DROP TABLE IF EXISTS chat_messages CASCADE;
DROP TABLE IF EXISTS chats CASCADE;
DROP TABLE IF EXISTS file_shares CASCADE;
DROP TABLE IF EXISTS file_ai_metadata CASCADE;
DROP TABLE IF EXISTS file_properties CASCADE;
DROP TABLE IF EXISTS file_storage CASCADE;
DROP TABLE IF EXISTS files CASCADE;
DROP TABLE IF EXISTS folders CASCADE;
DROP TABLE IF EXISTS project_members CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS user_profiles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS flowchart_status;
DROP TYPE IF EXISTS report_format;
DROP TYPE IF EXISTS report_status;
DROP TYPE IF EXISTS chat_type;
DROP TYPE IF EXISTS project_visibility;
DROP TYPE IF EXISTS message_role;
DROP TYPE IF EXISTS file_permission;
DROP TYPE IF EXISTS project_role;
DROP TYPE IF EXISTS chat_status;

DROP EXTENSION IF EXISTS "pgcrypto";