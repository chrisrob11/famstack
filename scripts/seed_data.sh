#!/bin/bash

# This script seeds the database with test data for the calendar.

# --- Configuration ---
# The path to the SQLite database file.
DB_FILE="famstack.db"
# The path to the SQL seed script.
SQL_FILE="scripts/sql_test_data/seed_calendar.sql"

# --- Execution ---
if [ ! -f "$DB_FILE" ]; then
    echo "❌ Error: Database file not found at '$DB_FILE'. Please ensure the file exists or update the DB_FILE variable in this script."
    exit 1
fi

if [ ! -f "$SQL_FILE" ]; then
    echo "❌ Error: SQL seed file not found at '$SQL_FILE'."
    exit 1
fi

echo "🌱 Seeding database '$DB_FILE' with data from '$SQL_FILE'..."
sqlite3 "$DB_FILE" < "$SQL_FILE"

# Check the exit code of the sqlite3 command
if [ $? -eq 0 ]; then
    echo "🎉 Database seeded successfully."
else
    echo "❌ Error: Failed to seed the database. Please check for errors above."
    exit 1
fi
