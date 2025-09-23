-- +goose Up
-- Add color and initial columns to family_members table for person identification
ALTER TABLE family_members ADD COLUMN color TEXT DEFAULT '#3b82f6' CHECK (color GLOB '#[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]');
ALTER TABLE family_members ADD COLUMN initial TEXT DEFAULT '';

-- Update existing family members with default initials (first letter of first name)
UPDATE family_members SET initial = UPPER(SUBSTR(first_name, 1, 1)) WHERE initial = '';

-- +goose Down
-- Remove the added columns
ALTER TABLE family_members DROP COLUMN color;
ALTER TABLE family_members DROP COLUMN initial;
