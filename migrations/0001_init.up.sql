CREATE TABLE operators (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    api_key TEXT UNIQUE NOT NULL,
    webhook_url TEXT NOT NULL,
    webhook_secret TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    external_player_id TEXT NOT NULL,
    jurisdiction TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(operator_id, external_player_id)
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    player_id INT NOT NULL REFERENCES players(id),
    launch_token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE rounds (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    player_id INT NOT NULL REFERENCES players(id),
    server_seed TEXT NOT NULL,
    client_seed TEXT NOT NULL,
    outcome INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bets (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    player_id INT NOT NULL REFERENCES players(id),
    round_id INT NOT NULL REFERENCES rounds(id),
    amount NUMERIC(12,2) NOT NULL,
    outcome INT NOT NULL,
    win_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    status TEXT NOT NULL, -- pending / won / lost
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (idempotency_key, operator_id)
);

CREATE TABLE outbox (
    id SERIAL PRIMARY KEY,
    bet_id INT NOT NULL REFERENCES bets(id),
    operator_id INT NOT NULL REFERENCES operators(id),
    player_id INT NOT NULL REFERENCES players(id),
    amount NUMERIC(18,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE webhook_events (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    retries INT NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    player_id INT REFERENCES players(id),
    action TEXT NOT NULL,
    details JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE operator_limits (
    operator_id INT PRIMARY KEY REFERENCES operators(id),
    max_bet NUMERIC(18,2) NOT NULL DEFAULT 1000,
    allowed_jurisdictions TEXT[],
    daily_loss_limit NUMERIC(18,2) NOT NULL,
    daily_win_limit NUMERIC(18,2) NOT NULL
);