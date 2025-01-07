-- Remove constraints
ALTER TABLE game_results 
    DROP CONSTRAINT IF EXISTS game_results_transaction_id_key;

-- Revert created_at back to DATE
ALTER TABLE game_results 
    ALTER COLUMN created_at TYPE DATE;