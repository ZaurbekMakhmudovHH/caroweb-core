-- Migration: Add is_verified column to user_profiles table
ALTER TABLE user_profiles
    ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT FALSE;