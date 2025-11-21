-- Operators table
CREATE TABLE operators (
   id SERIAL PRIMARY KEY,
   name TEXT NOT NULL,
   api_key TEXT UNIQUE NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Players table
CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    external_player_id TEXT NOT NULL,
    jurisdiction TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(operator_id, external_player_id)
);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    player_id INT NOT NULL REFERENCES players(id),
    launch_token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Rounds table
CREATE TABLE rounds (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    player_id INT NOT NULL REFERENCES players(id),
    server_seed TEXT NOT NULL,
    client_seed TEXT NOT NULL,
    outcome INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Bets table
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
    bet_id INT NOT NULL,
    operator_id INT NOT NULL,
    player_id INT NOT NULL,
    amount NUMERIC(18,2) NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    processed BOOLEAN DEFAULT FALSE
);