#!/bin/bash

# Script to reorganize and manage database migrations

# Create directories if they don't exist
mkdir -p migrations/versioned

# Move schema.sql to the versioned migrations folder as the first migration
if [ -f migrations/schema.sql ]; then
  echo "Converting schema.sql to versioned migration..."
  cp migrations/schema.sql migrations/versioned/000001_initial_schema.up.sql
  
  # Generate down migration for initial schema
  echo "-- Drop all tables" > migrations/versioned/000001_initial_schema.down.sql
  echo "DROP TABLE IF EXISTS answers CASCADE;" >> migrations/versioned/000001_initial_schema.down.sql
  echo "DROP TABLE IF EXISTS participants CASCADE;" >> migrations/versioned/000001_initial_schema.down.sql
  echo "DROP TABLE IF EXISTS questions CASCADE;" >> migrations/versioned/000001_initial_schema.down.sql
  echo "DROP TABLE IF EXISTS quiz_sessions CASCADE;" >> migrations/versioned/000001_initial_schema.down.sql
  echo "DROP TABLE IF EXISTS quizzes CASCADE;" >> migrations/versioned/000001_initial_schema.down.sql
  echo "DROP TABLE IF EXISTS users CASCADE;" >> migrations/versioned/000001_initial_schema.down.sql
  echo "DROP EXTENSION IF EXISTS \"uuid-ossp\";" >> migrations/versioned/000001_initial_schema.down.sql
fi

# Convert add_quiz_code.sql to versioned migration
if [ -f migrations/add_quiz_code.sql ]; then
  echo "Converting add_quiz_code.sql to versioned migration..."
  cp migrations/add_quiz_code.sql migrations/versioned/000002_add_quiz_code.up.sql
  
  # Generate down migration for quiz code
  echo "-- Revert quiz code changes" > migrations/versioned/000002_add_quiz_code.down.sql
  echo "ALTER TABLE quizzes DROP COLUMN IF EXISTS code;" >> migrations/versioned/000002_add_quiz_code.down.sql
fi

echo "Migration files have been reorganized. You can now use the Makefile targets to manage migrations."
echo "Example: DB_URL=\"postgres://username:password@localhost:5432/quiz_db?sslmode=disable\" make migrate-up"