-- +goose Up
-- Update users table for proper authentication

-- Add new columns for user names and authentication
ALTER TABLE users ADD COLUMN first_name TEXT;
ALTER TABLE users ADD COLUMN last_name TEXT;
ALTER TABLE users ADD COLUMN nickname TEXT;
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN last_login_at DATETIME;

-- Update role enum to use new values
-- Note: SQLite doesn't support ALTER COLUMN, so we'll handle role mapping in code

-- Migrate existing data (split name into first/last)
UPDATE users SET
    first_name = CASE
        WHEN instr(name, ' ') > 0 THEN substr(name, 1, instr(name, ' ') - 1)
        ELSE name
    END,
    last_name = CASE
        WHEN instr(name, ' ') > 0 THEN substr(name, instr(name, ' ') + 1)
        ELSE ''
    END
WHERE first_name IS NULL;

-- Note: No sessions table needed - using stateless JWT tokens

-- Update sample data with real password hashes (bcrypt)
-- Password: "password123" for all sample users
UPDATE users SET
    password_hash = '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeYQjNjJIGKyq2/OK',
    first_name = CASE
        WHEN id = 'user1' THEN 'John'
        WHEN id = 'user2' THEN 'Jane'
        WHEN id = 'user3' THEN 'Bobby'
    END,
    last_name = CASE
        WHEN id = 'user1' THEN 'Smith'
        WHEN id = 'user2' THEN 'Smith'
        WHEN id = 'user3' THEN 'Smith'
    END,
    nickname = CASE
        WHEN id = 'user3' THEN 'Bob'
        ELSE NULL
    END,
    email_verified = TRUE
WHERE id IN ('user1', 'user2', 'user3');

-- +goose Down
-- Remove added columns (SQLite limitation - can't easily drop columns)
-- Would need to recreate table, but for development this is fine