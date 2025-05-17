-- Migration: Remove is_verified column from user_profiles table
ALTER TABLE user_profiles
DROP COLUMN is_verified;
