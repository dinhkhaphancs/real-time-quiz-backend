-- Create database tables for the real-time quiz application

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Quizzes table
CREATE TABLE IF NOT EXISTS quizzes (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    admin_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Quiz sessions table
CREATE TABLE IF NOT EXISTS quiz_sessions (
    quiz_id UUID PRIMARY KEY REFERENCES quizzes(id) ON DELETE CASCADE,
    current_question_id UUID NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL,
    current_question_started_at TIMESTAMP NULL
);

-- Questions table
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY,
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    option_a TEXT NOT NULL,
    option_b TEXT NOT NULL,
    option_c TEXT NOT NULL,
    option_d TEXT NOT NULL,
    correct_answer CHAR(1) NOT NULL,
    time_limit INTEGER NOT NULL,
    "order" INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    joined_at TIMESTAMP NOT NULL,
    score INTEGER NOT NULL DEFAULT 0
);

-- Answers table
CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    selected_option CHAR(1) NOT NULL,
    answered_at TIMESTAMP NOT NULL,
    time_taken FLOAT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    score INTEGER NOT NULL,
    UNIQUE(user_id, question_id)
);

-- Add foreign key constraint for current_question_id
ALTER TABLE quiz_sessions 
ADD CONSTRAINT fk_quiz_sessions_current_question 
FOREIGN KEY (current_question_id) 
REFERENCES questions(id) ON DELETE SET NULL;