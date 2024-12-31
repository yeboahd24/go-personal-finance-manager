-- Add user_id column to categories table
ALTER TABLE categories ADD COLUMN user_id UUID REFERENCES users(id);

-- Update existing categories to use the system user (we'll need to create this user first)
-- This will be handled by the application code when initializing default categories
