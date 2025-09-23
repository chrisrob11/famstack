-- Seed data for family, members, and calendar events for today's date.
-- This script is designed to be self-contained and re-runnable.
-- Uses date('now') to ensure events appear on the current date.

-- Ensure test family 'fam1' and its members exist.
-- Using INSERT OR IGNORE to avoid errors if they already exist.
INSERT OR IGNORE INTO families (id, name) VALUES ('fam1', 'The Test Family');

INSERT OR IGNORE INTO family_members (id, family_id, first_name, last_name, member_type, email, role, color, initial) VALUES
('user1', 'fam1', 'Alex', 'Test', 'adult', 'alex@test.com', 'admin', '#3b82f6', 'A'),
('user2', 'fam1', 'Bailey', 'Test', 'adult', 'bailey@test.com', 'user', '#10b981', 'B'),
('user3', 'fam1', 'Casey', 'Test', 'child', 'casey@test.com', 'user', '#f59e0b', 'C');

-- Update existing family members with colors and initials if they exist
UPDATE family_members SET color = '#3b82f6', initial = 'A' WHERE id = 'user1';
UPDATE family_members SET color = '#10b981', initial = 'B' WHERE id = 'user2';
UPDATE family_members SET color = '#f59e0b', initial = 'C' WHERE id = 'user3';


-- Clean out any previous test calendar data for today to make this script re-runnable.
DELETE FROM unified_calendar_event_attendees WHERE event_id IN (
  SELECT id FROM unified_calendar_events WHERE date(start_time) = date('now') AND family_id = 'fam1'
);
DELETE FROM unified_calendar_events WHERE date(start_time) = date('now') AND family_id = 'fam1';

-- Event 1: Morning Standup (9:00 - 9:30 AM) - 30 minutes
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_1', 'fam1', 'Morning Standup',
        datetime('now', 'start of day', '+9 hours'),
        datetime('now', 'start of day', '+9 hours', '+30 minutes'),
        0, 'event', '#3b82f6', 'user1');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_1', 'user1');
INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_1', 'user2');

-- Event 2: Team Lunch (12:30 - 1:30 PM) - 1 hour
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_2', 'fam1', 'Team Lunch',
        datetime('now', 'start of day', '+12 hours', '+30 minutes'),
        datetime('now', 'start of day', '+13 hours', '+30 minutes'),
        0, 'event', '#10b981', 'user1');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_2', 'user1');
INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_2', 'user2');
INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_2', 'user3');

-- Event 3: Project Deadline (All Day)
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_3', 'fam1', 'Project Deadline',
        datetime('now', 'start of day'),
        datetime('now', 'start of day', '+23 hours', '+59 minutes', '+59 seconds'),
        1, 'reminder', '#ef4444', 'user2');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_3', 'user1');

-- Event 4: 1:1 with Sarah (3:00 - 3:30 PM) - 30 minutes
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_4', 'fam1', '1:1 with Sarah',
        datetime('now', 'start of day', '+15 hours'),
        datetime('now', 'start of day', '+15 hours', '+30 minutes'),
        0, 'appointment', '#8b5cf6', 'user2');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_4', 'user2');

-- Event 5: Team Planning Session (10:00 AM - 12:15 PM) - 2 hours 15 minutes
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_5', 'fam1', 'Team Planning Session',
        datetime('now', 'start of day', '+10 hours'),
        datetime('now', 'start of day', '+12 hours', '+15 minutes'),
        0, 'event', '#f97316', 'user3');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_5', 'user3');
INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_5', 'user1');

-- Event 6: Quick Check-in (2:00 - 2:30 PM) - 30 minutes
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_6', 'fam1', 'Quick Check-in',
        datetime('now', 'start of day', '+14 hours'),
        datetime('now', 'start of day', '+14 hours', '+30 minutes'),
        0, 'event', '#06b6d4', 'user1');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_6', 'user1');
INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_6', 'user3');

SELECT '✅ Test families, members, and calendar events seeded successfully.';
