-- Remove user_id column from transactions
ALTER TABLE transactions DROP COLUMN IF EXISTS user_id;
