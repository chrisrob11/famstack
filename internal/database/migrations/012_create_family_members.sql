-- +goose Up
-- Create family members table to separate family membership from user accounts

CREATE TABLE family_members (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    name TEXT NOT NULL,
    nickname TEXT,
    member_type TEXT NOT NULL DEFAULT 'child', -- 'adult', 'child', 'pet'
    age INTEGER,
    avatar_url TEXT,
    user_id TEXT, -- Links to users table if they have an account

    -- Display and sorting
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,

    -- Metadata
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- Foreign keys
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Index for performance
CREATE INDEX idx_family_members_family_id ON family_members(family_id);
CREATE INDEX idx_family_members_user_id ON family_members(user_id);

-- Migrate existing users to family members
-- Create family members for existing users
INSERT INTO family_members (id, family_id, name, nickname, member_type, user_id, age, is_active)
SELECT
    'member-' || u.id as id,
    u.family_id,
    CASE
        WHEN u.first_name IS NOT NULL AND u.last_name IS NOT NULL
        THEN u.first_name || ' ' || u.last_name
        WHEN u.first_name IS NOT NULL
        THEN u.first_name
        ELSE u.email
    END as name,
    u.nickname,
    'adult' as member_type,
    u.id as user_id,
    NULL as age, -- Could infer from role but keeping simple
    TRUE as is_active
FROM users u;

-- Add some sample family members (kids and pets) for the Smith family
INSERT INTO family_members (family_id, name, member_type, age, display_order, is_active)
SELECT
    'family1' as family_id,
    'Emma' as name,
    'child' as member_type,
    8 as age,
    10 as display_order,
    TRUE as is_active
WHERE EXISTS (SELECT 1 FROM families WHERE id = 'family1');

INSERT INTO family_members (family_id, name, member_type, age, display_order, is_active)
SELECT
    'family1' as family_id,
    'Buddy' as name,
    'pet' as member_type,
    3 as age,
    20 as display_order,
    TRUE as is_active
WHERE EXISTS (SELECT 1 FROM families WHERE id = 'family1');

-- Update tasks to reference family members instead of users where appropriate
-- Note: We'll keep assigned_to as TEXT for now and update the API layer to handle both

-- +goose Down
-- Remove family members table and revert changes
DROP INDEX IF EXISTS idx_family_members_user_id;
DROP INDEX IF EXISTS idx_family_members_family_id;
DROP TABLE IF EXISTS family_members;