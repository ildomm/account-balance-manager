-- Change created_at to TIMESTAMP for more precise timing
ALTER TABLE game_results 
    ALTER COLUMN created_at TYPE TIMESTAMP(6) WITHOUT TIME ZONE;

-- Add UNIQUE constraint to transaction_id to prevent duplicates
ALTER TABLE game_results 
    ADD CONSTRAINT game_results_transaction_id_key UNIQUE (transaction_id);
