INSERT INTO operators (name, api_key, webhook_url, webhook_secret)
VALUES
    ('Demo Operator', 'demo-operator-key', 'https://webhook.site/51380c55-679f-4468-9eb5-78235acc527e', 'demo-secret')
    ON CONFLICT DO NOTHING;

WITH op AS (
    SELECT id FROM operators WHERE api_key = 'demo-operator-key'
)
INSERT INTO operator_limits (operator_id, max_bet, allowed_jurisdictions, daily_loss_limit, daily_win_limit)
SELECT id, 1000, ARRAY['EU', 'CA', 'RS'], 5000, 5000 FROM op
ON CONFLICT (operator_id) DO NOTHING;

WITH op AS (
    SELECT id FROM operators WHERE api_key = 'demo-operator-key'
)
INSERT INTO players (operator_id, external_player_id, jurisdiction)
SELECT id, 'player-001', 'RS' FROM op
    ON CONFLICT DO NOTHING;

WITH op AS (
    SELECT id FROM operators WHERE api_key = 'demo-operator-key'
)
INSERT INTO players (operator_id, external_player_id, jurisdiction)
SELECT id, 'player-002', 'CA' FROM op
    ON CONFLICT DO NOTHING;