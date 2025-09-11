-- +goose Up
-- Create families table
CREATE TABLE IF NOT EXISTS families (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('parent', 'child', 'admin')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE
);

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
    FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Insert sample data for testing
INSERT OR IGNORE INTO families (id, name) VALUES ('fam1', 'The Smith Family');
INSERT OR IGNORE INTO users (id, family_id, name, email, password_hash, role) VALUES 
    ('user1', 'fam1', 'John Smith', 'john@smith.com', 'hash1', 'parent'),
    ('user2', 'fam1', 'Jane Smith', 'jane@smith.com', 'hash2', 'parent'),
    ('user3', 'fam1', 'Bobby Smith', 'bobby@smith.com', 'hash3', 'child');

-- Insert sample tasks
INSERT OR IGNORE INTO tasks (id, family_id, assigned_to, title, description, task_type, status, priority, created_by) VALUES
    ('task1', 'fam1', 'user3', 'Take out trash', 'Take the garbage bins to the curb', 'chore', 'pending', 1, 'user1'),
    ('task2', 'fam1', 'user2', 'Buy groceries', 'Pick up milk, bread, and eggs', 'todo', 'pending', 2, 'user1'),
    ('task3', 'fam1', 'user3', 'Clean room', 'Tidy up bedroom and make bed', 'chore', 'completed', 1, 'user1'),
    ('task4', 'fam1', NULL, 'Plan weekend trip', 'Research activities and book hotel', 'todo', 'pending', 3, 'user2');

-- +goose Down
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS families;