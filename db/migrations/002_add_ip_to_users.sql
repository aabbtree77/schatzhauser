-- Add ip column to users and index it for faster counting
ALTER TABLE users ADD COLUMN ip TEXT NOT NULL DEFAULT '';

-- Create an index on ip to speed up COUNT(*) queries
CREATE INDEX IF NOT EXISTS idx_users_ip ON users(ip);

