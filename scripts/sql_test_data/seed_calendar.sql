-- Seed data for family, members, and calendar events for 2025-09-22.
-- This script is designed to be self-contained and re-runnable.

-- Ensure test family 'fam1' and its members exist.
-- Using INSERT OR IGNORE to avoid errors if they already exist.
INSERT OR IGNORE INTO families (id, name) VALUES ('fam1', 'The Test Family');

INSERT OR IGNORE INTO family_members (id, family_id, first_name, last_name, member_type, email, role) VALUES
('user1', 'fam1', 'Alex', 'Test', 'adult', 'alex@test.com', 'admin'),
('user2', 'fam1', 'Bailey', 'Test', 'adult', 'bailey@test.com', 'user'),
('user3', 'fam1', 'Casey', 'Test', 'child', 'casey@test.com', 'user');


-- Clean out any previous test calendar data for this day to make this script re-runnable.
DELETE FROM unified_calendar_event_attendees WHERE event_id IN (
  SELECT id FROM unified_calendar_events WHERE date(start_time) = '2025-09-22' AND family_id = 'fam1'
);
DELETE FROM unified_calendar_events WHERE date(start_time) = '2025-09-22' AND family_id = 'fam1';

-- Event 1: Morning Standup (9:00 - 9:15 AM)
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_1', 'fam1', 'Morning Standup', '2025-09-22 09:00:00', '2025-09-22 09:15:00', 0, 'event', '#3b82f6', 'user1');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_1', 'user1');
INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_1', 'user2');

-- Event 2: Team Lunch (12:30 - 1:30 PM)
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_2', 'fam1', 'Team Lunch', '2025-09-22 12:30:00', '2025-09-22 13:30:00', 0, 'event', '#10b981', 'user1');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_2', 'user1');
INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_2', 'user2');
INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_2', 'user3');

-- Event 3: Project Deadline (All Day)
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_3', 'fam1', 'Project Deadline', '2025-09-22 00:00:00', '2025-09-22 23:59:59', 1, 'reminder', '#ef4444', 'user2');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_3', 'user1');

-- Event 4: 1:1 with Sarah (3:00 - 3:30 PM)
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_4', 'fam1', '1:1 with Sarah', '2025-09-22 15:00:00', '2025-09-22 15:30:00', 0, 'appointment', '#8b5cf6', 'user2');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_4', 'user2');

-- Event 5: Overlapping event (9:10 - 10:00 AM)
INSERT INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_5', 'fam1', 'Follow-up on Standup', '2025-09-22 09:10:00', '2025-09-22 10:00:00', 0, 'event', '#f97316', 'user3');

INSERT INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_5', 'user3');

SELECT '✅ Test families, members, and calendar events seeded successfully.';
