-- +goose Up
-- Create families table
CREATE TABLE IF NOT EXISTS families (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create family_members table (combines users and family member info)
CREATE TABLE IF NOT EXISTS family_members (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    name TEXT NOT NULL,
    nickname TEXT,
    member_type TEXT NOT NULL DEFAULT 'child', -- 'adult', 'child', 'pet'
    age INTEGER,
    avatar_url TEXT,

    -- Auth fields (optional - only for members who can login)
    email TEXT UNIQUE,
    password_hash TEXT,
    role TEXT CHECK (role IN ('shared', 'user', 'admin')),
    email_verified BOOLEAN DEFAULT FALSE,
    last_login_at DATETIME,

    -- Display and sorting
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,

    -- Metadata
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE
);

-- Create indexes for family_members
CREATE INDEX idx_family_members_family_id ON family_members(family_id);
CREATE UNIQUE INDEX idx_family_members_email ON family_members(email) WHERE email IS NOT NULL;

-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    assigned_to TEXT,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    task_type TEXT NOT NULL CHECK (task_type IN ('todo', 'chore', 'appointment')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed')),
    priority INTEGER DEFAULT 0,
    due_date DATETIME,
    points INTEGER DEFAULT 0,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to) REFERENCES family_members(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES family_members(id) ON DELETE CASCADE
);

-- Insert sample data for testing
INSERT OR IGNORE INTO families (id, name) VALUES ('fam1', 'The Smith Family');

-- Insert sample family members (adults with auth, children without)
INSERT OR IGNORE INTO family_members (id, family_id, name, member_type, age, email, password_hash, role, email_verified, display_order, is_active) VALUES
    ('member1', 'fam1', 'John Smith', 'adult', 42, 'john@smith.com', '$argon2id$v=19$m=65536,t=3,p=4$hash1', 'admin', true, 1, true),
    ('member2', 'fam1', 'Jane Smith', 'adult', 39, 'jane@smith.com', '$argon2id$v=19$m=65536,t=3,p=4$hash2', 'admin', true, 2, true),
    ('member3', 'fam1', 'Bobby Smith', 'child', 12, NULL, NULL, NULL, false, 3, true);

-- Insert sample tasks
INSERT OR IGNORE INTO tasks (id, family_id, assigned_to, title, description, task_type, status, priority, created_by) VALUES
    ('task1', 'fam1', 'member3', 'Take out trash', 'Take the garbage bins to the curb', 'chore', 'pending', 1, 'member1'),
    ('task2', 'fam1', 'member2', 'Buy groceries', 'Pick up milk, bread, and eggs', 'todo', 'pending', 2, 'member1'),
    ('task3', 'fam1', 'member3', 'Clean room', 'Tidy up bedroom and make bed', 'chore', 'completed', 1, 'member1'),
    ('task4', 'fam1', NULL, 'Plan weekend trip', 'Research activities and book hotel', 'todo', 'pending', 3, 'member2');

-- +goose Down
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS family_members;
DROP TABLE IF EXISTS families;