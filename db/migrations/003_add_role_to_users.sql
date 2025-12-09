-- 003_add_role_to_users.sql
-- Add role column to users and a small index for quick lookup

ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
