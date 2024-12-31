-- Add user_id column to transactions
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES users(id);

-- Update existing transactions with user_id from accounts
UPDATE transactions t 
SET user_id = a.user_id 
FROM accounts a 
WHERE t.account_id = a.id 
AND t.user_id IS NULL;

-- Make user_id NOT NULL after populating data
ALTER TABLE transactions ALTER COLUMN user_id SET NOT NULL;
