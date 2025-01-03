CREATE INDEX IF NOT EXISTS game_results_pxt_user_id ON game_results (user_id);
CREATE INDEX IF NOT EXISTS game_results_pxt_transaction_id ON game_results (transaction_id);
CREATE INDEX IF NOT EXISTS game_results_pxt_created_at ON game_results (created_at);