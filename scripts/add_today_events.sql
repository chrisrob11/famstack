-- Quick script to add events for today to test person identification system

-- Clean out today's events first
DELETE FROM unified_calendar_event_attendees WHERE event_id IN (
  SELECT id FROM unified_calendar_events WHERE date(start_time) = date('now') AND family_id = 'fam1'
);
DELETE FROM unified_calendar_events WHERE date(start_time) = date('now') AND family_id = 'fam1';

-- Ensure test family members exist with colors and timezone
INSERT OR IGNORE INTO families (id, name, timezone) VALUES ('fam1', 'The Test Family', 'America/New_York');

INSERT OR IGNORE INTO family_members (id, family_id, first_name, last_name, member_type, email, role, color, initial) VALUES
('user1', 'fam1', 'Alex', 'Test', 'adult', 'alex@test.com', 'admin', '#3b82f6', 'A'),
('user2', 'fam1', 'Bailey', 'Test', 'adult', 'bailey@test.com', 'user', '#10b981', 'B'),
('user3', 'fam1', 'Casey', 'Test', 'child', 'casey@test.com', 'user', '#f59e0b', 'C');

-- Update existing family members with colors
UPDATE family_members SET color = '#3b82f6', initial = 'A' WHERE id = 'user1';
UPDATE family_members SET color = '#10b981', initial = 'B' WHERE id = 'user2';
UPDATE family_members SET color = '#f59e0b', initial = 'C' WHERE id = 'user3';

-- Add today's events (times stored in UTC, representing Eastern timezone events)
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES
('event_today_1', 'fam1', 'Morning Standup',
 datetime('now', 'utc', 'start of day', '+14 hours'),
 datetime('now', 'utc', 'start of day', '+14 hours', '+30 minutes'),
 0, 'event', '#3b82f6', 'user1'),

('event_today_2', 'fam1', 'Team Lunch',
 datetime('now', 'utc', 'start of day', '+17 hours', '+30 minutes'),
 datetime('now', 'utc', 'start of day', '+18 hours', '+30 minutes'),
 0, 'event', '#10b981', 'user1'),

('event_today_3', 'fam1', 'Project Deadline',
 datetime('now', 'utc', 'start of day', '+5 hours'),
 datetime('now', 'utc', 'start of day', '+28 hours', '+59 minutes'),
 1, 'reminder', '#ef4444', 'user2'),

('event_today_4', 'fam1', 'Team Planning',
 datetime('now', 'utc', 'start of day', '+15 hours'),
 datetime('now', 'utc', 'start of day', '+17 hours', '+15 minutes'),
 0, 'event', '#f97316', 'user3');

-- Add attendees
INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES
('event_today_1', 'user1'),
('event_today_1', 'user2'),
('event_today_2', 'user1'),
('event_today_2', 'user2'),
('event_today_2', 'user3'),
('event_today_3', 'user1'),
('event_today_4', 'user3'),
('event_today_4', 'user1');

SELECT 'Events added for ' || date('now') || ' successfully!';