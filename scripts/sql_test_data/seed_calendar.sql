-- Seed data for family, members, and calendar events for today's date.
-- This script is designed to be self-contained and re-runnable.
-- Uses date('now') to ensure events appear on the current date.

-- Ensure test family 'fam1' and its members exist.
-- Using INSERT OR IGNORE to avoid errors if they already exist.
INSERT OR IGNORE INTO families (id, name, timezone) VALUES ('fam1', 'The Test Family', 'America/New_York');

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

-- Event 1: Morning Standup (9:00 - 9:30 AM Eastern, stored as UTC)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_1', 'fam1', 'Morning Standup',
        datetime('now', 'utc', 'start of day', '+14 hours'),
        datetime('now', 'utc', 'start of day', '+14 hours', '+30 minutes'),
        0, 'event', '#3b82f6', 'user1');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_1', 'user1');
INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_1', 'user2');

-- Event 2: Team Lunch (12:30 - 1:30 PM Eastern, stored as UTC)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_2', 'fam1', 'Team Lunch',
        datetime('now', 'utc', 'start of day', '+17 hours', '+30 minutes'),
        datetime('now', 'utc', 'start of day', '+18 hours', '+30 minutes'),
        0, 'event', '#10b981', 'user1');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_2', 'user1');
INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_2', 'user2');
INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_2', 'user3');

-- Event 3: Project Deadline (All Day - stored as UTC midnight to midnight)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_3', 'fam1', 'Project Deadline',
        datetime('now', 'utc', 'start of day', '+5 hours'),
        datetime('now', 'utc', 'start of day', '+28 hours', '+59 minutes', '+59 seconds'),
        1, 'reminder', '#ef4444', 'user2');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_3', 'user1');

-- Event 4: 1:1 with Sarah (3:00 - 3:30 PM Eastern, stored as UTC)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_4', 'fam1', '1:1 with Sarah',
        datetime('now', 'utc', 'start of day', '+20 hours'),
        datetime('now', 'utc', 'start of day', '+20 hours', '+30 minutes'),
        0, 'appointment', '#8b5cf6', 'user2');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_4', 'user2');

-- Event 5: Team Planning Session (10:00 AM - 12:15 PM Eastern, stored as UTC)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_5', 'fam1', 'Team Planning Session',
        datetime('now', 'utc', 'start of day', '+15 hours'),
        datetime('now', 'utc', 'start of day', '+17 hours', '+15 minutes'),
        0, 'event', '#f97316', 'user3');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_5', 'user3');
INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_5', 'user1');

-- Event 6: Quick Check-in (2:00 - 2:30 PM Eastern, stored as UTC)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_6', 'fam1', 'Quick Check-in',
        datetime('now', 'utc', 'start of day', '+19 hours'),
        datetime('now', 'utc', 'start of day', '+19 hours', '+30 minutes'),
        0, 'event', '#06b6d4', 'user1');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_6', 'user1');
INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_6', 'user3');

-- Event 7: Afternoon Deep Work (2 hours)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_7', 'fam1', 'Afternoon Deep Work',
        datetime('now', 'utc', 'start of day', '+21 hours'),
        datetime('now', 'utc', 'start of day', '+23 hours'),
        0, 'event', '#d946ef', 'user1');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_7', 'user1');

-- Event 8: Design Review (1 hour)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_8', 'fam1', 'Design Review',
        datetime('now', 'utc', 'start of day', '+13 hours'),
        datetime('now', 'utc', 'start of day', '+14 hours'),
        0, 'event', '#84cc16', 'user2');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_8', 'user2');
INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_8', 'user3');

-- Event 9: Sync with Marketing (30 mins)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_9', 'fam1', 'Sync with Marketing',
        datetime('now', 'utc', 'start of day', '+14 hours', '+30 minutes'),
        datetime('now', 'utc', 'start of day', '+15 hours'),
        0, 'event', '#22d3ee', 'user3');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_9', 'user3');

-- Event 10: Final Review (15 mins)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_10', 'fam1', 'Final Review',
        datetime('now', 'utc', 'start of day', '+18 hours', '+30 minutes'),
        datetime('now', 'utc', 'start of day', '+18 hours', '+45 minutes'),
        0, 'event', '#fbbf24', 'user1');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_10', 'user1');

-- OVERLAPPING EVENTS FOR TESTING LAYER SYSTEM

-- Event 11: Client Call (overlaps with Morning Standup 9:15-9:45 AM)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_11', 'fam1', 'Client Call',
        datetime('now', 'utc', 'start of day', '+14 hours', '+15 minutes'),
        datetime('now', 'utc', 'start of day', '+14 hours', '+45 minutes'),
        0, 'event', '#ec4899', 'user3');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_11', 'user3');

-- Event 12: Coffee Chat (overlaps with Team Planning Session 11:00 AM - 12:00 PM)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_12', 'fam1', 'Coffee Chat',
        datetime('now', 'utc', 'start of day', '+16 hours'),
        datetime('now', 'utc', 'start of day', '+17 hours'),
        0, 'event', '#14b8a6', 'user2');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_12', 'user2');

-- Event 13: Brief Standup (15 min overlap with Team Lunch 12:45-1:00 PM)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_13', 'fam1', 'Brief Standup',
        datetime('now', 'utc', 'start of day', '+17 hours', '+45 minutes'),
        datetime('now', 'utc', 'start of day', '+18 hours'),
        0, 'event', '#7c3aed', 'user1');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_13', 'user1');
INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_13', 'user3');

-- Event 14: Code Review (overlaps with 1:1 with Sarah and Quick Check-in 2:15-3:15 PM)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_14', 'fam1', 'Code Review',
        datetime('now', 'utc', 'start of day', '+19 hours', '+15 minutes'),
        datetime('now', 'utc', 'start of day', '+20 hours', '+15 minutes'),
        0, 'event', '#ea580c', 'user2');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_14', 'user2');

-- Event 15: Strategy Meeting (overlaps partially with Afternoon Deep Work 2:30-4:00 PM)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_15', 'fam1', 'Strategy Meeting',
        datetime('now', 'utc', 'start of day', '+19 hours', '+30 minutes'),
        datetime('now', 'utc', 'start of day', '+21 hours'),
        0, 'event', '#059669', 'user3');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_15', 'user3');
INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_15', 'user1');

-- Event 16: Quick Sync (15 min event overlapping with Design Review 8:30-8:45 AM)
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, start_time, end_time, all_day, event_type, color, created_by)
VALUES ('unified_event_16', 'fam1', 'Quick Sync',
        datetime('now', 'utc', 'start of day', '+13 hours', '+30 minutes'),
        datetime('now', 'utc', 'start of day', '+13 hours', '+45 minutes'),
        0, 'event', '#dc2626', 'user1');

INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id) VALUES ('unified_event_16', 'user1');

SELECT '✅ Test families, members, and calendar events with overlapping events seeded successfully.';
