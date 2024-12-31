-- Add username column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT;

-- Make username column NOT NULL and UNIQUE
UPDATE users SET username = email WHERE username IS NULL;
ALTER TABLE users ALTER COLUMN username SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);
